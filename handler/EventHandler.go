package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/ONSdigital/go-ns/log"
	"github.com/coocood/freecache"
)

//struct for eventhandler which handles the instance and the start of the api
type EventReport struct {
	InstanceID string `avro:"instance_id"`
	EventType  string `avro:"event_type"`
	EventMsg   string `avro:"event_message"`
}

type config struct {
	ImportAPIURL string
	AuthToken    string
}

type instance struct {
	InstanceID                string   `json:"instance_id"`
	NumberOfObservations      int64    `json:"total_observations"`
	TotalInsertedObservations int64    `json:"total_inserted_observations,omitempty"`
	State                     string   `json:"state"`
	Events                    *[]event `json:"events, omitempty"`
}

type event struct {
	Type          string `json:"type"`
	Message       string `json:"message"`
	MessageOffset string `json:"messageOffset"`
}

var failed = "failed"

//method to set an event not really needed but testing made things easier
func (e *EventReport) EventSetter(InstanceID string, EventType string, EventMessage string) {
	e.InstanceID = InstanceID
	e.EventType = EventType
	e.EventMsg = EventMessage
}

//main function that triggers everything else
func (e *EventReport) HandleEvent(httpClient *http.Client, authToken string, c *freecache.Cache) error {
	cfg := config{
		ImportAPIURL: "http://localhost:21800",
		AuthToken:    authToken,
	}

	status, events, err := e.checkInstance(httpClient, cfg)
	if err != nil {
		return err
	}

	eventTypes := e.EventType
	timeNow := time.Now().String()
	message := e.EventMsg
	msgOff := "0"

	Event := event{
		Type:          e.EventType,
		Message:       e.EventMsg,
		MessageOffset: "0",
	}

	jsonUpload := []byte(`{"type":"` + eventTypes + `","time":"` + timeNow + `","message":"` + message + `","messageOffset":"` + msgOff + `"}`)

	key := []byte(e.InstanceID)
	value := []byte(e.EventMsg)
	expire := 25

	check := arraySlicing(Event, *events)
	if check {
		_, err := c.Get(key)
		if err != nil {
			c.Set(key, value, expire)
		}
	}

	got, err := c.Get(key)
	if err != nil {
		c.Set(key, value, expire)
		err := e.putEvent(httpClient, jsonUpload, cfg, status)
		if err != nil {
			return err
		} else {
			return nil
		}
	} else {
		log.Info("This instance is saved inmemory already.", log.Data{e.InstanceID: "this instance is saved inmemory", "lock": got})
		return nil
	}
}

/*this puts an event into the database under the instance you chose
it does some checks to make sure the instance exists and checks the status
if the status isn't already failed it will turn that instance to failed */
func (e *EventReport) putEvent(httpClient *http.Client, json []byte, cfg config, status string) error {

	path := cfg.ImportAPIURL + "/instances/" + e.InstanceID + "/events"

	var URL *url.URL
	URL, err := url.Parse(path)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", URL.String(), bytes.NewBuffer(json))
	if err != nil {
		return err
	}

	//this needs to change but for testing purposes leave it in
	req.Header.Set("Internal-token", cfg.AuthToken)
	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	if e.EventType == "error" && status != failed {
		err := e.putJobStatus(httpClient, cfg)
		if err != nil {
			return err
		}
	}
	defer res.Body.Close()

	if res.StatusCode != 201 {
		err = errors.New("Bad response while updating inserts for Events")
	} else {
		log.Info("Successfully updated events in the import api", log.Data{"instance_id": e.InstanceID})
	}

	return err
}

func (e *EventReport) checkInstance(httpClient *http.Client, cfg config) (string, *[]event, error) {
	path := cfg.ImportAPIURL + "/instances/" + e.InstanceID
	event := &[]event{}
	var URL *url.URL

	URL, err := url.Parse(path)
	if err != nil {
		return "", event, err
	}
	res, err := http.Get(URL.String())
	if err != nil {
		return "", event, err
	}

	defer res.Body.Close()

	var instance instance
	// json := json.NewDecoder(res.Body).Decode(&Instance{})
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", event, err
	}
	if err := json.Unmarshal(body, &instance); err != nil {
		return "", event, err
	}
	if res.StatusCode == 200 {
		log.Info("Successfully found instance", log.Data{"instance_id": e.InstanceID})
		return instance.State, instance.Events, nil
	} else if res.StatusCode == 404 {
		return "", event, err
	} else {
		return "", event, err
	}
}

//This will put a error status in the state
func (e *EventReport) putJobStatus(httpClient *http.Client, cfg config) error {

	path := cfg.ImportAPIURL + "/instances/" + e.InstanceID
	log.Info("Instance id: "+e.InstanceID, log.Data{"instance_id": e.InstanceID})

	var URL *url.URL

	URL, err := url.Parse(path)
	if err != nil {
		return err
	}
	errorhandle := failed

	jsonUpload := []byte(`{"state":"` + errorhandle + `"}`)

	req, err := http.NewRequest("PUT", URL.String(), bytes.NewBuffer(jsonUpload))
	if err != nil {
		return err
	}

	req.Header.Set("Internal-token", cfg.AuthToken)
	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return errors.New("Bad response while updating inserts for Events")
	} else {
		log.Info("Successfully updated job state in the import api", log.Data{"instance_id": e.InstanceID})
		return nil
	}

}

func arraySlicing(a event, event []event) bool {
	for _, b := range event {
		// fmt.Println(b)
		if b == a {
			return true
		}
	}
	return false
}
