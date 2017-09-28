package handler

import (
	"time"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/dp-import-reporter/model"
	"encoding/json"
)

//go:generate moq -out ../mocks/handler_generated_mocks.go -pkg mocks . DatasetAPICli Cache

const (
	failed                    = "failed"
	errorType                 = "error"
	datasetAPIGetInstErr      = "datasetAPI.GetInstance return an error"
	datasetAPIAddEventErr     = "datasetAPI.AddEventToInstance return an error"
	datasetAPIUpdateStatusErr = "datasetAPI.UpdateInstanceStatus return an error"
	eventNotInCache           = "reportEvent not found in cache, checking dataset API"
	eventNotInInstance        = "event not found in instance.events, updating dataset api"
	updatingDSInstance        = "updating dataset api to set instance.status to failed"
	addingToLocalCache        = "adding reportEvent to local cache"
	updatingCacheTimeout      = "reportEvent found in cache, updating cache expiry time"
	handlingEvent             = "Handling report event"
	reportEventKey            = "reportEvent"
)

var (
	statusFailed = &model.State{State: failed}
)

type DatasetAPICli interface {
	GetInstance(instanceID string) (*model.Instance, error)
	AddEventToInstance(instanceID string, e *model.Event) error
	UpdateInstanceStatus(instanceID string, state *model.State) error
}

type Cache interface {
	Get(key []byte) (value []byte, err error)
	Set(key, value []byte, expireSeconds int) (err error)
	Del(key []byte) (affected bool)
	TTL(key []byte) (timeLeft uint32, err error)
}

type ReportEventHandler struct {
	DatasetAPI    DatasetAPICli
	Cache         Cache
	ExpireSeconds int
}

func (r ReportEventHandler) HandleEvent(e *model.ReportEvent) error {
	logDetails := log.Data{reportEventKey: e}
	log.Info(handlingEvent, logDetails)

	key, value, err := generateCacheKey(e)
	if err != nil {
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
	// It is in the cache so reset the time to live.
	ttl, _ := r.Cache.TTL(key)
	logDetails["key"] = key
	logDetails["TTL"] = ttl
	log.Info(updatingCacheTimeout, logDetails)
	r.Cache.Del(key)
	r.Cache.Set(key, value, r.ExpireSeconds)
	return nil
}

func generateCacheKey(r *model.ReportEvent) (key []byte, value []byte, err error) {
	cacheKey := model.CacheKey{
		InstanceID: r.InstanceID,
		EventType:  r.EventType,
		Service:    "TODO",
	}

	key, err = json.Marshal(cacheKey)
	if err != nil {
		return nil, nil, err
	}
	value, err = json.Marshal(r)
	if err != nil {
		return nil, nil, err
	}
	return key, value, err
}
