package client

import (
	"net/http"
	"fmt"
	"github.com/ONSdigital/go-ns/log"
	"encoding/json"
	"errors"
	"io"
	"github.com/ONSdigital/dp-import-reporter/model"
)

//go:generate moq -out ../mocks/dataset_api_generated_mocks.go -pkg mocks . HttpClient ResponseBodyReader

const (
	getInstanceURL  = "%s/instances/%s"
	authTokenHeader = "internal-token"
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
		return nil, errors.New("datasetAPIClient requires a non empty host")
	}
	if len(authToken) == 0 {
		return nil, errors.New("datasetAPIClient requires a non empty authToken")
	}
	if httpCli == nil {
		return nil, errors.New("datasetAPIClient requires a non nil HttpClient")
	}
	if respBodyReader == nil {
		return nil, errors.New("datasetAPIClient requires a non nil ResponseBodyReader")
	}

	return &DatasetAPIClient{
		host:           host,
		authToken:      authToken,
		httpClient:     httpCli,
		responseReader: respBodyReader,
	}, nil
}

// GetInstance get an instance from the Dataset API.
func (cli *DatasetAPIClient) GetInstance(instanceID string) (*model.Instance, error) {
	url := fmt.Sprintf(getInstanceURL, cli.host, instanceID)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.ErrorC("unexpected error when attempting to create get instance request", err, nil)
		return nil, err
	}

	req.Header.Set(authTokenHeader, cli.authToken)

	resp, err := cli.httpClient.Do(req)
	if err != nil {
		log.ErrorC("unexpected error returned from http client when making get instance request", err, nil)
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := errors.New("unexpected status code returned from dataset cli")
		log.ErrorC("unexpected status code returned from dataset cli", err, log.Data{
			"expected": http.StatusOK,
			"actual":   resp.StatusCode,
		})
		return nil, err
	}

	body, err := cli.responseReader.Read(resp.Body)
	if err != nil {
		log.ErrorC("unexpected error when attempting to read get instance response body", err, nil)
		return nil, err
	}

	var instance model.Instance
	if err := json.Unmarshal(body, &instance); err != nil {
		log.ErrorC("unexpected error when attempting to unmarshal get instance response", err, nil)
		return nil, err
	}
	return &instance, err
}
