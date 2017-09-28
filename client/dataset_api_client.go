package client

import (
	"net/http"
	"fmt"
	"github.com/ONSdigital/go-ns/log"
	"encoding/json"
	"errors"
	"io"
	"github.com/ONSdigital/dp-import-reporter/model"
	"bytes"
)

//go:generate moq -out ../mocks/dataset_api_generated_mocks.go -pkg mocks . HttpClient ResponseBodyReader

const (
	getInstanceURL        = "%s/instances/%s"
	putInstanceStateURL   = getInstanceURL
	addInstanceEventURL   = "%s/instances/%s/events"
	authTokenHeader       = "internal-token"
	expectedKey           = "expected"
	actualKey             = "actual"
	uriKey                = "uri"
	instanceIDKey         = "instanceID"
	eventKey              = "event"
	stateKey              = "state"
	httpClientDoErr       = "httpClient.Do returned an unexpected error"
	unexpectedHTTPStatus  = "unexpected status code returned from dataset api client"
	readResponseBodyErr   = "unexpected error while attempting to read HTTP response body"
	unmarshalResponseErr  = "unexpected error while attempting to unmarshal HTTP response body into domain object"
	instanceIDNil         = "instanceID is a required but was empty"
	eventNil              = "event required but was nil"
	marshallEventErr      = "unexpected error when attempting to marshal event to json"
	createRequestErr      = "unexpected error when attempting to create HTTP request"
	hostEmpty             = "requires a non empty host"
	authTokenEmpty        = "authToken required but was empty"
	httpClientNil         = "HttpClient required but was nil"
	responseBodyReaderNil = "ResponseBodyReader required but was nil"
	stateNil              = "state is required but was nil"
)

// HttpClient defines a client for making http requests.
type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// ResponseBodyReader defines an http response reader.
type ResponseBodyReader interface {
	Read(r io.Reader) ([]byte, error)
}

// DatasetAPIClient client for making http requests to the Dataset API.
type DatasetAPIClient struct {
	host           string
	authToken      string
	httpClient     HttpClient
	responseReader ResponseBodyReader
}

// NewDatasetAPIClient create a new DatasetAPIClient using the supplied configuration
func NewDatasetAPIClient(host string, authToken string, httpCli HttpClient, respBodyReader ResponseBodyReader) (*DatasetAPIClient, error) {
	if len(host) == 0 {
		return nil, newDatasetAPIError(hostEmpty)
	}
	if len(authToken) == 0 {
		return nil, newDatasetAPIError(authTokenEmpty)
	}
	if httpCli == nil {
		return nil, newDatasetAPIError(httpClientNil)
	}
	if respBodyReader == nil {
		return nil, newDatasetAPIError(responseBodyReaderNil)
	}

	return &DatasetAPIClient{
		host:           host,
		authToken:      authToken,
		httpClient:     httpCli,
		responseReader: respBodyReader,
	}, nil
}

// GetInstance make a HTTP GET request to the Dataset API to get the specified Instance
func (cli *DatasetAPIClient) GetInstance(instanceID string) (*model.Instance, error) {
	url := fmt.Sprintf(getInstanceURL, cli.host, instanceID)

	logData := log.Data{
		instanceIDKey: instanceID,
		uriKey:        url,
	}

	resp, err := cli.doRequest(url, http.MethodGet, nil, logData)
	if err != nil {
		log.Error(err, logData)
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := newDatasetAPIError(unexpectedHTTPStatus)
		logData[expectedKey] = http.StatusOK
		logData[actualKey] = resp.StatusCode
		log.Error(err, logData)
		return nil, err
	}

	body, err := cli.responseReader.Read(resp.Body)
	if err != nil {
		log.Error(err, logData)
		return nil, wrappedDatasetAPIError(readResponseBodyErr, err)
	}

	var instance model.Instance
	if err := json.Unmarshal(body, &instance); err != nil {
		log.Error(err, logData)
		return nil, wrappedDatasetAPIError(unmarshalResponseErr, err)
	}
	return &instance, err
}

// AddEventToInstance make a post request to the dataset API to add a report event to rge specified Instance
func (cli *DatasetAPIClient) AddEventToInstance(instanceID string, e *model.Event) error {
	if len(instanceID) == 0 {
		return newDatasetAPIError(instanceIDNil)
	}
	if e == nil {
		return newDatasetAPIError(eventNil)
	}

	url := fmt.Sprintf(addInstanceEventURL, cli.host, instanceID)

	logData := log.Data{
		instanceIDKey: instanceID,
		uriKey:        url,
		eventKey:      e,
	}

	resp, err := cli.doRequest(url, http.MethodPost, e, logData)
	if err != nil {
		log.Error(err, logData)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		err := newDatasetAPIError(unexpectedHTTPStatus)
		logData[expectedKey] = http.StatusCreated
		logData[actualKey] = resp.StatusCode
		log.Error(err, logData)

		return err
	}
	return nil
}

func (cli *DatasetAPIClient) UpdateInstanceStatus(instanceID string, state *model.State) error {
	if len(instanceID) == 0 {
		return newDatasetAPIError(instanceIDNil)
	}
	if state == nil {
		return newDatasetAPIError(stateNil)
	}

	url := fmt.Sprintf(putInstanceStateURL, cli.host, instanceID)

	logData := log.Data{
		instanceIDKey: instanceID,
		uriKey:        url,
		stateKey:      state,
	}
	resp, err := cli.doRequest(url, http.MethodPut, state, logData)
	if err != nil {
		log.Error(err, logData)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := newDatasetAPIError(unexpectedHTTPStatus)
		logData[expectedKey] = http.StatusCreated
		logData[actualKey] = resp.StatusCode
		log.Error(err, logData)

		return err
	}
	return nil
}

func (cli *DatasetAPIClient) doRequest(url string, httpMethod string, payload interface{}, logData log.Data) (*http.Response, error) {
	var body []byte
	var err error
	var req *http.Request
	var reader *bytes.Reader

	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			log.Error(err, logData)
			return nil, wrappedDatasetAPIError(marshallEventErr, err)
		}
		reader = bytes.NewReader(body)
		req, err = http.NewRequest(httpMethod, url, reader)
	} else {
		req, err = http.NewRequest(httpMethod, url, nil)
	}

	if err != nil {
		log.Error(err, logData)
		return nil, wrappedDatasetAPIError(createRequestErr, err)
	}

	resp, err := cli.httpClient.Do(req)
	if err != nil {
		log.Error(err, logData)
		return nil, wrappedDatasetAPIError(httpClientDoErr, err)
	}
	return resp, nil
}

func wrappedDatasetAPIError(context string, originalErr error) error {
	details := fmt.Sprintf("datasetAPIClient %s: %s", context, originalErr.Error())
	return errors.New(details)
}

func newDatasetAPIError(details string) error {
	return errors.New(fmt.Sprintf("datasetAPIClient: %s", details))
}
