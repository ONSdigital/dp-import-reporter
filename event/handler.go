package event

import (
	"time"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/dp-import-reporter/model"
)

//go:generate moq -out ../mocks/event_generated_mocks.go -pkg mocks . DatasetAPICli Cache EventHandler

const (
	failed                    = "failed"
	errorType                 = "error"
	datasetAPIGetInstErr      = "datasetAPI.GetInstance return an error"
	datasetAPIAddEventErr     = "datasetAPI.AddEventToInstance return an error"
	datasetAPIUpdateStatusErr = "datasetAPI.UpdateInstanceStatus return an error"
	eventNotInCache           = "report event not found in cache, checking dataset API"
	eventNotInInstance        = "report not found in instance.events, updating dataset api"
	updatingDSInstance        = "updating dataset api to set instance.status to failed"
	addingToLocalCache        = "adding report event to local cache"
	updatingCacheTimeout      = "report event found in cache, updating cache expiry time"
	handlingEvent             = "Handling report event"
	reportEventKey            = "reportEvent"
	generateCacheKeyValueErr  = "error while attempting to generate cache key value for report event"
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

// HandleEvent handle the reportEvent
func (r Handler) HandleEvent(e *model.ReportEvent) error {
	logDetails := log.Data{reportEventKey: *e}
	log.Info(handlingEvent, logDetails)

	key, value, err := e.GenCacheKeyAndValue()
	if err != nil {
		log.ErrorC(generateCacheKeyValueErr, err, logDetails)
		return err
	}

	// err != nil means not in cache.
	if _, err := r.Cache.Get(key); err != nil {
		log.Info(eventNotInCache, logDetails)
		i, err := r.DatasetAPI.GetInstance(e.InstanceID)
		if err != nil {
			log.ErrorC(datasetAPIGetInstErr, err, logDetails)
			return err
		}

		timeNow := time.Now()
		newEvent := &model.Event{
			Type:          e.EventType,
			Time:          &timeNow,
			Message:       e.EventMsg,
			MessageOffset: "0",
		}

		if !i.ContainsEvent(newEvent) {
			log.Info(eventNotInInstance, logDetails)

			if err := r.DatasetAPI.AddEventToInstance(i.InstanceID, newEvent); err != nil {
				log.ErrorC(datasetAPIAddEventErr, err, logDetails)
				return err
			}
			if e.EventType == errorType && i.State != failed {
				log.Info(updatingDSInstance, logDetails)

				if err := r.DatasetAPI.UpdateInstanceStatus(i.InstanceID, statusFailed); err != nil {
					log.ErrorC(datasetAPIUpdateStatusErr, err, logDetails)
					return err
				}
			}
		}

		log.Info(addingToLocalCache, logDetails)
		r.Cache.Set(key, value, r.ExpireSeconds)
		return nil
	}
	log.Info(updatingCacheTimeout, logDetails)

	// It is in the cache so reset the time to live.
	r.Cache.Del(key)
	r.Cache.Set(key, value, r.ExpireSeconds)
	return nil
}
