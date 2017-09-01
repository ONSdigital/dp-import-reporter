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

	instanceEvents := instanceEvent{
		Type:          e.EventType,
		Message:       e.EventMsg,
		MessageOffset: "0",
	}

	timeNow := time.Now().String()

	jsonUpload, err := json.Marshal(Event{
		Type:          e.EventType,
		Time:          timeNow,
		Message:       e.EventMsg,
		MessageOffset: "0",
	})

	key := []byte(e.InstanceID)
	value := []byte(e.EventMsg)
	expire := 25

	check := arraySlicing(instanceEvents, *events)
	if check {
		_, err := c.Get(key)
		if err != nil {
			c.Set(key, value, expire)
		} else {
			c.Del(key)
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
	req.Header.Set("Internal-token", cfg.ImportAuthToken)
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

	err = errorhandler(res.StatusCode)
	if err != nil {
		return err
	}
	return nil

}

func (e *EventReport) checkInstance(httpClient *http.Client, cfg *config.Config) (string, *[]instanceEvent, error) {
	path := cfg.ImportAPIURL + "/instances/" + e.InstanceID
	event := &[]instanceEvent{}
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
	err = errorhandler(res.StatusCode)
	if err != nil {
		return "", event, err
	}
	return instance.State, instance.Events, nil
}

// This will put a error status in the state
func (e *EventReport) putJobStatus(httpClient *http.Client, cfg *config.Config) error {

	path := cfg.ImportAPIURL + "/instances/" + e.InstanceID
	log.Info("Instance id: "+e.InstanceID, log.Data{"instance_id": e.InstanceID})

	var URL *url.URL

	URL, err := url.Parse(path)
	if err != nil {
		return err
	}
	errorhandle := failed
	jsonUpload, err := json.Marshal(&State{
		State: errorhandle,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", URL.String(), bytes.NewBuffer(jsonUpload))
	if err != nil {
		return err
	}

	req.Header.Set("Internal-token", cfg.ImportAuthToken)
	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	err = errorhandler(res.StatusCode)
	if err != nil {
		return err
	}
	return nil
}
func errorhandler(statusCode int) error {
	if statusCode == 200 || statusCode == 201 {
		log.Info("Successfully updated Event state in the import api", log.Data{"Status code": statusCode})
		return nil
	} else if statusCode == 404 {
		log.Info("Could not find instance", log.Data{"Status code": statusCode})
		return errors.New("Could not find instance")
	} else if statusCode == 401 {
		log.Info("Unauthorised access", log.Data{"Status code": statusCode})
		return errors.New("Unauthorised access")
	} else if statusCode == 400 {
		log.Info("Bad client request", log.Data{"Status code": statusCode})
		return errors.New("JSON was incorrect")
	} else {
		log.Info("Unrecoginsed error", log.Data{"Status code": statusCode})
		return errors.New("Unrecoginsed error.")
	}
}
func arraySlicing(a instanceEvent, event []instanceEvent) bool {
	for _, b := range event {
		// fmt.Println(b)
		if b == a {
			return true
		}
	}
	return false
}
