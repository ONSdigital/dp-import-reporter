package handler

//struct for eventhandler which handles the instance and the start of the api
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
	Events                    []*instanceEvent `json:"events, omitempty"`
}

//I have removed the time from this instanceEvent making event checks easier
type instanceEvent struct {
	Type          string `json:"type"`
	Message       string `json:"message"`
	MessageOffset string `json:"messageOffset"`
}

//Event struct including the time
type Event struct {
	Type          string `json:"type"`
	Message       string `json:"message"`
	Time          string `json:"time"`
	MessageOffset string `json:"messageOffset"`
}

//State struct is for assigning the state of the instance
type State struct {
	State string `json:"state"`
}
