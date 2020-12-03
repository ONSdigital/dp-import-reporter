package event

import (
	"context"
	"time"

	"github.com/ONSdigital/dp-import-reporter/model"
	"github.com/ONSdigital/log.go/log"
	"github.com/pkg/errors"
)

//go:generate moq -out ../mocks/event_generated_mocks.go -pkg mocks . DatasetAPICli Cache EventHandler

const (
	failed         = "failed"
	errorType      = "error"
	reportEventKey = "reportEvent"
)

var (
	statusFailed = &model.State{State: failed}
)

// DatasetAPICli defines that interface for a client of the DatasetAPI
type DatasetAPICli interface {
	GetInstance(ctx context.Context, instanceID string) (*model.Instance, error)
	AddEventToInstance(ctx context.Context, instanceID string, e *model.Event) error
	UpdateInstanceStatus(ctx context.Context, instanceID string, state *model.State) error
}

// Cache defines the behaviour of an in memory cache
type Cache interface {
	Get(key []byte) (value []byte, err error)
	Set(key, value []byte, expireSeconds int) (err error)
	Del(key []byte) (affected bool)
	TTL(key []byte) (timeLeft uint32, err error)
}

// Handler a struct for handling a ReportEvent
type Handler struct {
	DatasetAPI    DatasetAPICli
	Cache         Cache
	ExpireSeconds int
}

// HandleEvent if the event does not exist in the local cache add it to the dataset instance events (if it does not
// already exist) & add to the local cache, otherwise update the cache time to live.
func (h Handler) HandleEvent(ctx context.Context, e *model.ReportEvent) error {
	logDetails := log.Data{reportEventKey: *e}
	log.Event(ctx, "handling report event", log.INFO, logDetails)

	key, value, err := e.GenCacheKeyAndValue()
	if err != nil {
		return errors.Wrap(err, "error while attempting to generate cache key value for report event")
	}

	if _, err := h.Cache.Get(key); err != nil {
		log.Event(ctx, "report event not found in dp-import-reporter cache, retrieving instance from dataset API", log.INFO, logDetails)
		i, err := h.DatasetAPI.GetInstance(ctx, e.InstanceID)
		if err != nil {
			return errors.Wrap(err, "datasetAPI.GetInstance return an error")
		}

		timeNow := time.Now()
		newEvent := &model.Event{
			Type:          e.EventType,
			Time:          &timeNow,
			Message:       e.EventMsg,
			Service:       e.ServiceName,
			MessageOffset: "0", // TODO need to update GONS to be able to get this
		}

		if !i.ContainsEvent(newEvent) {
			log.Event(ctx, "report event not in instance.events, adding event to instance and persisting changes to dataset api", log.INFO, logDetails)

			if err := h.DatasetAPI.AddEventToInstance(ctx, i.InstanceID, newEvent); err != nil {
				return errors.Wrap(err, "datasetAPI.AddEventToInstance returned an error")
			}
			if e.EventType == errorType && i.State != failed {
				log.Event(ctx, "updating instance.status to failed and persisting changes to dataset api", log.INFO, logDetails)

				if err := h.DatasetAPI.UpdateInstanceStatus(ctx, i.InstanceID, statusFailed); err != nil {
					return errors.Wrap(err, "datasetAPI.UpdateInstanceStatus return an error")
				}
			}
		}

		log.Event(ctx, "adding report event to dp-import-reporter cache", log.INFO, logDetails)
		h.Cache.Set(key, value, h.ExpireSeconds)
		return nil
	}
	log.Event(ctx, "report event found in dp-import-reporter cache, updating cache expiry time", log.INFO, logDetails)
	// If the key exists in the cache delete it and set it again to reset the time to live
	h.Cache.Del(key)
	h.Cache.Set(key, value, h.ExpireSeconds)
	return nil
}
