package model

import (
	"time"
	"reflect"
	"encoding/json"
	"errors"
)

const (
	reportEventNil              = "reportEvent.GenCacheKeyAndValue requires non nil reportEvent"
	reportEventInstanceIDEmpty  = "reportEvent.GenCacheKeyAndValue requires a non empty instanceID"
	reportEventTypeEmpty        = "reportEvent.GenCacheKeyAndValue requires a non empty eventType"
	reportEventServiceNameEmpty = "reportEvent.GenCacheKeyAndValue requires a non empty serviceName"
)

type cacheKey struct {
	instanceID  string
	serviceName string
	eventType   string
}

func (e *ReportEvent) GenCacheKeyAndValue() ([]byte, []byte, error) {
	if e == nil {
		return nil, nil, errors.New(reportEventNil)
	}
	if len(e.InstanceID) == 0 {
		return nil, nil, errors.New(reportEventInstanceIDEmpty)
	}
	if len(e.EventType) == 0 {
		return nil, nil, errors.New(reportEventTypeEmpty)
	}
	if len(e.ServiceName) == 0 {
		return nil, nil, errors.New(reportEventServiceNameEmpty)
	}

	cacheKey := &cacheKey{
		instanceID:  e.InstanceID,
		eventType:   e.EventType,
		serviceName: e.ServiceName,
	}

	key, err := json.Marshal(cacheKey)
	if err != nil {
		return nil, nil, err
	}

	val, err := json.Marshal(e)
	if err != nil {
		return nil, nil, err
	}
	return key, val, err
}

//struct for eventhandler which handles the instance and the start of the api
type ReportEvent struct {
	InstanceID  string `avro:"instance_id"`
	EventType   string `avro:"event_type"`
	EventMsg    string `avro:"event_message"`
	ServiceName string `avro:"service_name"`
}

//Instance provides a struct for all the instance information
type Instance struct {
	InstanceID                string   `json:"instance_id"`
	NumberOfObservations      int64    `json:"total_observations"`
	TotalInsertedObservations int64    `json:"total_inserted_observations,omitempty"`
	State                     string   `json:"state"`
	Events                    []*Event `json:"events, omit"`
}

func (i *Instance) ContainsEvent(target *Event) bool {
	if target == nil {
		return false
	}

	for _, event := range i.Events {
		if event.EqualsIgnoreTime(target) {
			return true
		}
	}
	return false
}

//Event struct including the time
type Event struct {
	Type          string     `bson:"type,omitempty"           json:"type"`
	Service       string     `bson:"serviceName,omitempty"    json:"serviceName"`
	Time          *time.Time `bson:"time,omitempty"           json:"time"`
	Message       string     `bson:"message,omitempty"        json:"message"`
	MessageOffset string     `bson:"message_offset,omitempty" json:"message_offset"`
}

func (this *Event) EqualsIgnoreTime(that *Event) bool {
	if this == nil && that == nil {
		return false
	}
	if this == nil && that != nil {
		return false
	}
	if this != nil && that == nil {
		return false
	}
	return reflect.DeepEqual(this.Type, that.Type) && reflect.DeepEqual(this.MessageOffset, that.MessageOffset) && reflect.DeepEqual(this.Message, that.Message)
}

// State struct representing the state of the dataset.
type State struct {
	State string `json:"state"`
}
