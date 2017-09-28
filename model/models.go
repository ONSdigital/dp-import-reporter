package model

import (
	"time"
	"reflect"
)

type CacheKey struct {
	InstanceID string
	Service    string
	EventType  string
}

//struct for eventhandler which handles the instance and the start of the api
type ReportEvent struct {
	InstanceID string `avro:"instance_id"`
	EventType  string `avro:"event_type"`
	EventMsg   string `avro:"event_message"`
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
	Service       string     `bson:"service,omitempty"        json:"service"`
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

//State struct is for assigning the state of the instance
type State struct {
	State string `json:"state"`
}
