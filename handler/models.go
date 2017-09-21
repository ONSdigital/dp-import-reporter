package handler

import "time"

//EventReport is struct for eventhandler which handles the messages being created/sent
type EventReport struct {
	InstanceID string `avro:"instance_id"`
	EventType  string `avro:"event_type"`
	EventMsg   string `avro:"event_message"`
}

//Instance provides a struct for all the instance information
type Instance struct {
	InstanceID                string           `json:"instance_id"`
	NumberOfObservations      int64            `json:"total_observations"`
	TotalInsertedObservations int64            `json:"total_inserted_observations,omitempty"`
	State                     string           `json:"state"`
	Events                    []*InstanceEvent `json:"events, omitempty"`
}

//I have removed the time from this instanceEvent making event checks easier
type InstanceEvent struct {
	Type          string `bson:"type,omitempty"           json:"type"`
	Message       string `bson:"message,omitempty"        json:"message"`
	MessageOffset string `bson:"message_offset,omitempty" json:"message_offset"`
}

//Event struct including the time
type Event struct {
	Type          string     `bson:"type,omitempty"           json:"type"`
	Time          *time.Time `bson:"time,omitempty"           json:"time"`
	Message       string     `bson:"message,omitempty"        json:"message"`
	MessageOffset string     `bson:"message_offset,omitempty" json:"message_offset"`
}

//State struct is for assigning the state of the instance
type State struct {
	State string `json:"state"`
}
