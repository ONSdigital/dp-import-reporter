package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/ONSdigital/dp-import-reporter/config"

	"github.com/ONSdigital/go-ns/log"
	"github.com/coocood/freecache"
)

var failed = "failed"

//main function that triggers everything else
func (e *EventReport) HandleEvent(httpClient *http.Client, c *freecache.Cache, cfg *config.Config) error {

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
		log.Info("What is the json", log.Data{"json": jsonUpload})
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
func (e *EventReport) putEvent(httpClient *http.Client, json []byte, cfg *config.Config, status string) error {

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
	log.Info(cfg.ImportAuthToken, nil)
	//this needs to change but for testing purposes leave it in
	req.Header.Set("Internal-token", "D0108EA-825D-411C-9B1D-41EF7727F465")
	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	// if e.EventType == "error" && status != failed {
	// 	err := e.putJobStatus(httpClient, cfg)
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	defer res.Body.Close()
	log.Info("here?", log.Data{"instance_id": e.InstanceID})

	if res.StatusCode == 201 {
		log.Info("Successfully updated events in the import api", log.Data{"instance_id": e.InstanceID})
		return nil
	} else if res.StatusCode == 404 {
		log.Info("Could not find instance", log.Data{"instance_id": e.InstanceID})
		return errors.New("Could not find instance")
	} else if res.StatusCode == 401 {
		log.Info("Unauthorised access", log.Data{"instance_id": e.InstanceID})
		return errors.New("Unauthorised access")
	} else if res.StatusCode == 400 {
		return errors.New("Bad client request received.")
	} else {
		return errors.New("Critical error.")
	}

}

func (e *EventReport) checkInstance(httpClient *http.Client, cfg *config.Config) (string, *[]event, error) {
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

	var instance Instance
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
// func (e *EventReport) putJobStatus(httpClient *http.Client, cfg *config.Config) error {
//
// 	path := cfg.ImportAPIURL + "/instances/" + e.InstanceID
// 	log.Info("Instance id: "+e.InstanceID, log.Data{"instance_id": e.InstanceID})
//
// 	var URL *url.URL
//
// 	URL, err := url.Parse(path)
// 	if err != nil {
// 		return err
// 	}
// 	errorhandle := failed
// 	jsonUpload, err := json.Marshal(&state{
// 		state: errorhandle,
// 	})
// 	if err != nil {
// 		return err
// 	}
//
// 	req, err := http.NewRequest("PUT", URL.String(), bytes.NewBuffer(jsonUpload))
// 	if err != nil {
// 		return err
// 	}
//
// 	req.Header.Set("Internal-token", cfg.ImportAuthToken)
// 	res, err := httpClient.Do(req)
// 	if err != nil {
// 		return err
// 	}
//
// 	defer res.Body.Close()
// 	if res.StatusCode == 200 {
// 		log.Info("Successfully updated job state in the import ap", log.Data{"instance_id": e.InstanceID})
// 		return nil
// 	} else if res.StatusCode == 404 {
// 		log.Info("Could not find instance", log.Data{"instance_id": e.InstanceID})
// 		return errors.New("Could not find instance")
// 	} else if res.StatusCode == 403 {
// 		log.Info("Unauthorised access", log.Data{"instance_id": e.InstanceID})
// 		return errors.New("Unauthorised access")
// 	} else if res.StatusCode == 400 {
// 		log.Info("Bad client request", nil)
// 		return errors.New("JSON was incorrect")
// 	} else {
// 		return errors.New("CRITICAL ERROR.")
// 	}
//
// }

func arraySlicing(a event, event []event) bool {
	for _, b := range event {
		// fmt.Println(b)
		if b == a {
			return true
		}
	}
	return false
}
