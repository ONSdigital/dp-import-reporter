package event

import (
	"github.com/ONSdigital/dp-import-reporter/model"
	"github.com/ONSdigital/go-ns/log"
	"github.com/pkg/errors"
	"time"
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
	GetInstance(instanceID string) (*model.Instance, error)
	AddEventToInstance(instanceID string, e *model.Event) error
	UpdateInstanceStatus(instanceID string, state *model.State) error
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
func (r Handler) HandleEvent(e *model.ReportEvent) error {
	logDetails := log.Data{reportEventKey: *e}
	log.Info("handler: handling report event", logDetails)

	key, value, err := e.GenCacheKeyAndValue()
	if err != nil {
		return errors.Wrap(err, "error while attempting to generate cache key value for report event")
	}

	if _, err := r.Cache.Get(key); err != nil {
		log.Info("handler: report event not found in dp-import-reporter cache, retrieving instance from dataset API", logDetails)
		i, err := r.DatasetAPI.GetInstance(e.InstanceID)
		if err != nil {
			return errors.Wrap(err, "datasetAPI.GetInstance return an error")
		}

		timeNow := time.Now()
		newEvent := &model.Event{
			Type:          e.EventType,
			Time:          &timeNow,
			Message:       e.EventMsg,
			MessageOffset: "0",
		}

		if !i.ContainsEvent(newEvent) {
			log.Info("handler: report event not in instance.events, adding event to instance and persisting changes to dataset api", logDetails)

			if err := r.DatasetAPI.AddEventToInstance(i.InstanceID, newEvent); err != nil {
				return errors.Wrap(err, "datasetAPI.AddEventToInstance returned an error")
			}
			if e.EventType == errorType && i.State != failed {
				log.Info("handler: updating instance.status to failed and persisting changes to dataset api", logDetails)

				if err := r.DatasetAPI.UpdateInstanceStatus(i.InstanceID, statusFailed); err != nil {
					return errors.Wrap(err, "datasetAPI.UpdateInstanceStatus return an error")
				}
			}
		}

		log.Info("handler: adding report event to dp-import-reporter cache", logDetails)
		r.Cache.Set(key, value, r.ExpireSeconds)
		return nil
	}
	log.Info("handler: report event found in dp-import-reporter cache, updating cache expiry time", logDetails)
	// If the key exists in the cache delete it and set it again to reset the time to live
	r.Cache.Del(key)
	r.Cache.Set(key, value, r.ExpireSeconds)
	return nil
}
