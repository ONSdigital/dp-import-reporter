package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/ONSdigital/dp-import-reporter/logging"
	"github.com/ONSdigital/dp-import-reporter/model"
	"github.com/ONSdigital/go-ns/log"
	"github.com/pkg/errors"
)

//go:generate moq -out ../mocks/dataset_api_generated_mocks.go -pkg mocks . HTTPClient ResponseBodyReader

const (
	getInstanceURL      = "%s/instances/%s"
	putInstanceStateURL = getInstanceURL
	addInstanceEventURL = "%s/instances/%s/events"
	authTokenHeader     = "Internal-Token"
	authorizationHeader = "Authorization"
	uriKey              = "uri"
	instanceKey         = "instance"
	instanceIDKey       = "instanceID"
	stateKey            = "state"
	httpMethodKey       = "httpMethod"
	requestBodyKey      = "requestBody"
)

var (
	errValidation = errors.New("dataset api client validation error")
	logger        = logging.Logger{Prefix: "client.DatasetAPI"}
)

// HTTPClient defines a client for making http requests.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// ResponseBodyReader defines an http response reader.
type ResponseBodyReader interface {
	Read(r io.Reader) ([]byte, error)
}

// DatasetAPIClient client for making http requests to the Dataset API.
type DatasetAPIClient struct {
	authToken           string
	datasetAPIAuthToken string
	host                string
	httpClient          HTTPClient
	responseReader      ResponseBodyReader
}

// NewDatasetAPIClient create a new DatasetAPIClient using the supplied configuration
func NewDatasetAPIClient(authToken string, host string, datasetAPIAuthToken string, httpCli HTTPClient, respBodyReader ResponseBodyReader) (*DatasetAPIClient, error) {

	if len(authToken) == 0 {
		return nil, errors.Wrap(errValidation, "non empty service authToken required")
	}
	if len(host) == 0 {
		return nil, errors.Wrap(errValidation, "non empty host required")
	}
	// TODO Remove check for dataset api auth token as we now only require an auth token for the "Authorization" header
	if len(datasetAPIAuthToken) == 0 {
		return nil, errors.Wrap(errValidation, "non empty dataset API authToken required")
	}
	if httpCli == nil {
		return nil, errors.Wrap(errValidation, "non nil HTTPClient required")
	}
	if respBodyReader == nil {
		return nil, errors.Wrap(errValidation, "non nil ResponseBodyReader required")
	}

	return &DatasetAPIClient{
		authToken:           authToken,
		datasetAPIAuthToken: datasetAPIAuthToken,
		host:                host,
		httpClient:          httpCli,
		responseReader:      respBodyReader,
	}, nil
}

// GetInstance make a HTTP GET request to the Dataset API to get the specified Instance
func (cli *DatasetAPIClient) GetInstance(instanceID string) (*model.Instance, error) {
	if len(instanceID) == 0 {
		return nil, errors.Wrap(errValidation, "GetInstance requires a non empty instanceID")
	}
	url := fmt.Sprintf(getInstanceURL, cli.host, instanceID)

	logData := log.Data{
		instanceIDKey: instanceID,
		uriKey:        url,
	}

	resp, err := cli.doRequest(url, http.MethodGet, nil, logData)
	if err != nil {
		return nil, errors.Wrap(err, "GetInstance HTTPClient.doRequest returned an error")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, incorrectStatusError("GetInstance", url, http.MethodGet, http.StatusOK, resp.StatusCode)
	}

	body, err := cli.responseReader.Read(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "GetInstance error while attempting to read HTTP response body")
	}

	var instance model.Instance
	if err = json.Unmarshal(body, &instance); err != nil {
		return nil, errors.Wrap(err, "GetInstance error while attempting to unmarshal HTTP response body into model.Instance")
	}
	logger.Info("GetInstance completed successfully", log.Data{instanceKey: instance})
	return &instance, err
}

// AddEventToInstance make a post request to the dataset API to add a report event to get the specified Instance
func (cli *DatasetAPIClient) AddEventToInstance(instanceID string, e *model.Event) error {
	if len(instanceID) == 0 {
		return errors.Wrap(errValidation, "AddEventToInstance requires a non empty instanceID")
	}
	if e == nil {
		return errors.Wrap(errValidation, "AddEventToInstance requires a non empty event")
	}

	url := fmt.Sprintf(addInstanceEventURL, cli.host, instanceID)

	logData := log.Data{
		instanceIDKey: instanceID,
		uriKey:        url,
	}

	resp, err := cli.doRequest(url, http.MethodPost, e, logData)
	if err != nil {
		return errors.Wrap(err, "AddEventToInstance HTTPClient.doRequest returned an error")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return incorrectStatusError("AddEventToInstance", url, http.MethodPost, http.StatusOK, resp.StatusCode)
	}
	logger.Info("AddEventToInstance completed successfully", logData)
	return nil
}

// UpdateInstanceStatus send a PUT request to the dataset API to update the status of the specified instance
func (cli *DatasetAPIClient) UpdateInstanceStatus(instanceID string, state *model.State) error {
	if len(instanceID) == 0 {
		return errors.Wrap(errValidation, "UpdateInstanceStatus requires a non empty instanceID")
	}
	if state == nil {
		return errors.Wrap(errValidation, "UpdateInstanceStatus requires a non nil state")
	}

	url := fmt.Sprintf(putInstanceStateURL, cli.host, instanceID)

	logData := log.Data{
		instanceIDKey: instanceID,
		uriKey:        url,
		stateKey:      state,
	}
	resp, err := cli.doRequest(url, http.MethodPut, state, logData)
	if err != nil {
		return errors.Wrap(err, "UpdateInstanceStatus HTTPClient.doRequest returned an error")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return incorrectStatusError("UpdateInstanceStatus", url, http.MethodPut, http.StatusOK, resp.StatusCode)
	}
	logger.Info("UpdateInstanceStatus completed successfully", logData)
	return nil
}

// doRequest builds a Dataset API request from the specified url, method & payload & sets the required authentication header
func (cli *DatasetAPIClient) doRequest(url string, httpMethod string, payload interface{}, logData log.Data) (*http.Response, error) {
	var body []byte
	var err error
	var req *http.Request
	var reader *bytes.Reader

	logData[uriKey] = url
	logData[httpMethodKey] = httpMethod

	if payload != nil {
		body, err = json.Marshal(payload)
		logData[requestBodyKey] = string(body)

		if err != nil {
			log.Error(err, logData)
			return nil, errors.Wrap(err, fmt.Sprintf("error when attempting to marshal %s to json", reflect.TypeOf(payload).Name()))
		}

		reader = bytes.NewReader(body)
		req, err = http.NewRequest(httpMethod, url, reader)
	} else {
		req, err = http.NewRequest(httpMethod, url, nil)
	}

	if err != nil {
		return nil, errors.Wrap(err, "error while attempting to create HTTP request")
	}

	// TODO Remove authTokenHeader header, now uses "Authorization" header
	req.Header.Set(authTokenHeader, cli.datasetAPIAuthToken)
	req.Header.Set(authorizationHeader, cli.authToken)

	logger.Info("making HTTP request", logData)
	resp, err := cli.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	logger.Info("HTTP response received", nil)
	return resp, nil
}

func incorrectStatusError(context string, url string, method string, expected int, actual int) error {
	return errors.Errorf("%s HTTPClient.doRequest returned an incorrect response status: url: %s, method: %s, expected status: %d, actual status: %d", context, url, method, expected, actual)
}
