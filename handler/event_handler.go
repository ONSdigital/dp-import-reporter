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
	"github.com/ONSdigital/go-ns/errorhandler/models"
	"github.com/ONSdigital/go-ns/log"
	"github.com/coocood/freecache"
)

const failed = "failed"

//main function that triggers everything else
func HandleEvent(c *freecache.Cache, cfg *config.Config, e *errorModel.EventReport) error {
	log.Info("Starting error handle", log.Data{"INSTANCE_ID": e.InstanceID, "ERROR_MSG": e.EventMsg})

	status, events, err := checkInstance(cfg, e)
	if err != nil {
		return err
	}

	log.Info("Successfully checked instance", log.Data{"INSTANCE_ID": e.InstanceID, "INSTANCE_STATE": status})
	instanceEvents := &errorModel.InstanceEvent{
		Type:          e.EventType,
		Message:       e.EventMsg,
		MessageOffset: "0",
	}

	timeNow := time.Now()

	jsonUpload, err := json.Marshal(errorModel.Event{
		Type:          e.EventType,
		Time:          &timeNow,
		Message:       e.EventMsg,
		MessageOffset: "0",
	})

	key := []byte(e.InstanceID)
	value := []byte(e.EventMsg)
	expire := 25

	check := arraySlicing(instanceEvents, events)
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
		err := insertEvent(jsonUpload, cfg, status, e)
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
func insertEvent(json []byte, cfg *config.Config, status string, e *errorModel.EventReport) error {

	path := cfg.DatasetAPIURL + "/instances/" + e.InstanceID + "/events"

	URL, err := urlParser(path)
	if err != nil || URL == nil {
		return err
	}

	res, err := apiRequests(URL, "POST", json, cfg)
	if err != nil {
		return err
	}
	if e.EventType == "error" && status != failed {
		err := putJobStatus(cfg, e)
		if err != nil {
			return err
		}
	}

	err = responseStatus(res.StatusCode)
	if err != nil {
		return err
	}
	return nil

}

func checkInstance(cfg *config.Config, e *errorModel.EventReport) (string, []*errorModel.InstanceEvent, error) {
	log.Info("Checking instance avaiable:", log.Data{"Instance_ID": e.InstanceID})

	path := cfg.DatasetAPIURL + "/instances/" + e.InstanceID

	event := make([]*errorModel.InstanceEvent, 0)

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

	var instance errorModel.Instance

	log.Info("Reading response body...", nil)
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.ErrorC("Error reading reponse body", err, log.Data{"STATUS_RETURNED": res.StatusCode})
		return "", event, err
	}
	log.Info("Successfully read response body...", nil)

	log.Info("Attempting unmarshalling response body...", nil)
	if err := json.Unmarshal(body, &instance); err != nil {
		err := responseStatus(res.StatusCode)
		if err != nil {
			log.ErrorC("Error unmarshalling response body", err, log.Data{"STATUS_CODE": res.StatusCode})
			return "", event, err
		}
	}
	log.Info("Successfully unmarshalled data", log.Data{"INSTANCE": e.InstanceID})

	err = responseStatus(res.StatusCode)
	if err != nil {
		log.ErrorC("Non 200 or 201 response status returned", err, log.Data{"STATUS_CODE": res.StatusCode})
		return "", event, err
	}
	return instance.State, instance.Events, nil
}

// This will put a error status in the state
func putJobStatus(cfg *config.Config, e *errorModel.EventReport) error {

	path := cfg.DatasetAPIURL + "/instances/" + e.InstanceID

	URL, err := urlParser(path)
	if err != nil || URL == nil {
		return err
	}

	log.Info("Attempting to marshal state...", nil)
	jsonUpload, err := json.Marshal(&errorModel.State{
		State: failed,
	})

	if err != nil {
		log.ErrorC("Unsuccessful marshal of state", err, nil)
		return err
	}
	log.Info("Successfully marshaled state", nil)

	res, err := apiRequests(URL, "PUT", jsonUpload, cfg)
	if err != nil {
		return err
	}
	err = responseStatus(res.StatusCode)
	if err != nil {
		return err
	}
	return nil
}

func apiRequests(URL *url.URL, request string, jsonUpload []byte, cfg *config.Config) (*http.Response, error) {
	log.Info("Attempting request", log.Data{"REQUESTED_URL": URL.String(), "REQUEST_METHOD": request})
	req, err := http.NewRequest(request, URL.String(), bytes.NewBuffer(jsonUpload))
	if err != nil {
		log.ErrorC("Unsuccessful making request", err, log.Data{"REQUESTED_URL": URL.String()})
		return nil, err
	}

	req.Header.Set("Internal-token", cfg.ImportAuthToken)
	log.Info("Token set... Requesting httpclient...", nil)
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.ErrorC("Error requesting httpclient", err, nil)
		return nil, err
	}
	log.Info("Successfully requested httpclient", nil)

	defer res.Body.Close()

	return res, nil
}
func urlParser(path string) (*url.URL, error) {
	var url *url.URL
	log.Info("Attempting parsing path: "+path, nil)

	url, err := url.Parse(path)
	if err != nil {
		log.ErrorC("Unsuccessful parsing of path", err, nil)
		return nil, err
	}
	log.Info("Successfully parsed path", log.Data{"URL": url.String()})
	return url, nil
}

func responseStatus(statusCode int) error {
	switch statusCode {
	case 200, 201:
		log.Info("Successfully connection", log.Data{"Status code": statusCode})
		return nil
	case 400:
		log.Info("Bad client request", log.Data{"Status code": statusCode})
		return errors.New("JSON was incorrect")
	case 401:
		log.Info("Unauthorised access", log.Data{"Status code": statusCode})
		return errors.New("Unauthorised access")
	case 404:
		log.Info("Could not find instance", log.Data{"Status code": statusCode})
		return errors.New("Could not find instance")
	default:
		log.Info("Unrecoginsed error", log.Data{"Status code": statusCode})
		return errors.New("Unrecoginsed error")
	}
}

func arraySlicing(a *errorModel.InstanceEvent, event []*errorModel.InstanceEvent) bool {
	for _, b := range event {
		if reflect.DeepEqual(*a, *b) {
			return true
		}
	}
	return false
}
