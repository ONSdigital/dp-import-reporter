package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"time"

	"github.com/ONSdigital/dp-import-reporter/config"
	"github.com/ONSdigital/go-ns/log"
	"github.com/coocood/freecache"
)

const failed = "failed"

//main function that triggers everything else
func (e *EventReport) HandleEvent(httpClient *http.Client, c *freecache.Cache, cfg *config.Config) error {
	log.Info("Starting error handle", log.Data{"instance_id": e.InstanceID, "error_msg": e.EventMsg})

	status, events, err := e.checkInstance(httpClient, cfg)
	if err != nil {
		return err
	}
	log.Info("Successfully checked instance", log.Data{"INSTANCE_ID": e.InstanceID, "INSTANCE_STATE": status})
	instanceEvents := &InstanceEvent{
		Type:          e.EventType,
		Message:       e.EventMsg,
		MessageOffset: "0",
	}

	timeNow := time.Now()

	jsonUpload, err := json.Marshal(Event{
		Type:          e.EventType,
		Time:          &timeNow,
		Message:       e.EventMsg,
		MessageOffset: "0",
	})

	key := []byte(e.InstanceID)
	value := []byte(e.EventMsg)
	expire := 25

	if ok := arraySlicing(instanceEvents, events); ok {
		if _, err := c.Get(key); err != nil {
			c.Set(key, value, expire)
		} else {
			c.Del(key)
		}
	}

	got, err := c.Get(key)
	if err != nil {
		c.Set(key, value, expire)
		err := e.insertEvent(httpClient, jsonUpload, cfg, status)
		if err != nil {
			return err
		}
		return nil
	}
	log.Info("This instance is saved inmemory already.", log.Data{e.InstanceID: "this instance is saved inmemory", "lock": got})
	return nil
}

/*this puts an event into the database under the instance you chose
it does some checks to make sure the instance exists and checks the status
if the status isn't already failed it will turn that instance to failed */
func (e *EventReport) insertEvent(httpClient *http.Client, json []byte, cfg *config.Config, status string) error {

	path := cfg.DatasetAPIURL + "/instances/" + e.InstanceID + "/events"

	URL, err := urlParser(path)
	if err != nil || URL == nil {
		return err
	}

	res, err := apiRequests(URL, "POST", json, httpClient, cfg)
	if err != nil {
		return err
	}
	if e.EventType == "error" && status != failed {
		err := e.putJobStatus(httpClient, cfg)
		if err != nil {
			return err
		}
	}

	err = errorhandler(res.StatusCode)
	if err != nil {
		return err
	}
	return nil

}

func (e *EventReport) checkInstance(httpClient *http.Client, cfg *config.Config) (string, []*InstanceEvent, error) {
	log.Info("Checking instance avaiable:", log.Data{"Instance_ID": e.InstanceID})

	path := cfg.DatasetAPIURL + "/instances/" + e.InstanceID

	event := make([]*InstanceEvent, 0)

	URL, err := urlParser(path)
	if err != nil || URL == nil {
		return "", event, err
	}

	log.Info("Making GET request to URL...", log.Data{"URL_REQUESTED": URL.String()})
	res, err := http.Get(URL.String())
	if err != nil {
		log.ErrorC("Error making GET Request", err, log.Data{"URL_REQUESTED": URL.String()})
		return "", event, errors.New("Invalid instance ID")
	}
	log.Info("Successfully made GET request", log.Data{"URL_REQUESTED": URL.String()})

	defer res.Body.Close()

	var instance Instance

	log.Info("Reading response body...", nil)
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.ErrorC("Error reading reponse body", err, log.Data{"STATUS_RETURNED": res.StatusCode})
		return "", event, err
	}
	log.Info("Successfully read response body...", nil)

	log.Info("Attempting unmarshalling response body...", nil)
	if err := json.Unmarshal(body, &instance); err != nil {
		err := errorhandler(res.StatusCode)
		if err != nil {
			log.ErrorC("Error unmarshalling response body", err, log.Data{"STATUS_CODE": res.StatusCode})
			return "", event, err
		}
	}
	log.Info("Successfully unmarshalled data", log.Data{"INSTANCE": e.InstanceID})

	err = errorhandler(res.StatusCode)
	if err != nil {
		log.ErrorC("Non 200 or 201 response status returned", err, log.Data{"STATUS_CODE": res.StatusCode})
		return "", event, err
	}
	return instance.State, instance.Events, nil
}

// This will put a error status in the state
func (e *EventReport) putJobStatus(httpClient *http.Client, cfg *config.Config) error {

	path := cfg.DatasetAPIURL + "/instances/" + e.InstanceID

	url, err := urlParser(path)
	if err != nil || url == nil {
		return err
	}

	log.Info("Attempting to marshal state...", nil)
	jsonUpload, err := json.Marshal(&State{
		State: failed,
	})

	if err != nil {
		log.ErrorC("Unsuccessful marshal of state", err, nil)
		return err
	}
	log.Info("Successfully marshaled state", nil)

	res, err := apiRequests(url, "PUT", jsonUpload, httpClient, cfg)
	if err != nil {
		return err
	}
	err = errorhandler(res.StatusCode)
	if err != nil {
		return err
	}
	return nil
}

func apiRequests(URL *url.URL, request string, jsonUpload []byte, httpClient *http.Client, cfg *config.Config) (*http.Response, error) {
	log.Info("Attempting request", log.Data{"REQUESTED_URL": URL.String(), "REQUEST_METHOD": request})
	req, err := http.NewRequest(request, URL.String(), bytes.NewBuffer(jsonUpload))
	if err != nil {
		log.ErrorC("Unsuccessful making request", err, log.Data{"REQUESTED_URL": URL.String()})
		return nil, err
	}

	req.Header.Set("Internal-token", cfg.ImportAuthToken)
	log.Info("Token set... Requesting httpclient...", nil)
	res, err := httpClient.Do(req)
	if err != nil {
		log.ErrorC("Error requesting httpclient", err, nil)
		return nil, err
	}
	log.Info("Successfully requested httpclient", nil)

	defer res.Body.Close()

	return res, nil
}
func urlParser(path string) (*url.URL, error) {
	var URL *url.URL
	log.Info("Attempting parsing path: "+path, nil)

	URL, err := url.Parse(path)
	if err != nil {
		log.ErrorC("Unsuccessful parsing of path", err, nil)
		return nil, err
	}
	log.Info("Successfully parsed path", log.Data{"URL": URL.String()})
	return URL, nil
}

func errorhandler(statusCode int) error {
	switch statusCode {
	case 200, 201:
		log.Info("Successful connection", log.Data{"status_code": statusCode})
		return nil
	case 400:
		log.Info("Bad client request", log.Data{"status_code": statusCode})
		return errors.New("JSON was incorrect")
	case 401:
		log.Info("Unauthorised access", log.Data{"status_code": statusCode})
		return errors.New("Unauthorised access")
	case 404:
		log.Info("Could not find instance", log.Data{"status_code": statusCode})
		return errors.New("Could not find instance")
	default:
		log.Info("Unrecoginsed error", log.Data{"status_code": statusCode})
		return errors.New("Unrecoginsed error")
	}
}

func arraySlicing(a *InstanceEvent, event []*InstanceEvent) bool {
	for _, b := range event {
		if reflect.DeepEqual(*a, *b) {
			return true
		}
	}
	return false
}
