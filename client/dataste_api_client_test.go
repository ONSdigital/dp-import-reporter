package client

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/ONSdigital/dp-import-reporter/mocks"

	"io"
	"github.com/ONSdigital/dp-import-reporter/model"
	"encoding/json"
	"net/http"
	"io/ioutil"
	"bytes"
	"errors"
	"reflect"
)

const (
	host           = "http://localhost:8080"
	auth           = "secret_password"
	testInstanceID = "1234567890"
)

var (
	validInstance = &model.Instance{
		State: "RED",
		Events: []*model.InstanceEvent{
			{
				Message:       "Error",
				Type:          "Error",
				MessageOffset: "99",
			},
		},
	}
)

func TestDatasetAPIClient_GetInstance(t *testing.T) {
	Convey("Given a correctly configured DatasetAPIClient", t, func() {
		body, _ := json.Marshal(validInstance)
		respBodyReader, _, httpClient := setup(body, http.StatusOK)

		cli := &DatasetAPIClient{
			host:           host,
			authToken:      auth,
			responseReader: respBodyReader,
			httpClient:     httpClient,
		}

		Convey("When GetInstance is called with valid parameters", func() {
			i, err := cli.GetInstance(testInstanceID)

			Convey("Then the returned values are as expected", func() {
				So(i, ShouldResemble, validInstance)
				So(err, ShouldBeNil)
			})

			Convey("And httpClient.Do is called 1 time with the expected parameters", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				So(httpClient.DoCalls()[0].Req.URL.String(), ShouldEqual, "http://localhost:8080/instances/"+testInstanceID)
				So(httpClient.DoCalls()[0].Req.Header.Get(authTokenHeader), ShouldEqual, auth)
			})

			Convey("And responseBodyReader.Read is called 1 time with the expected parameters", func() {
				So(len(respBodyReader.ReadCalls()), ShouldEqual, 1)
			})
		})
	})
}

func TestDatasetAPIClient_GetInstance_HttpCliErr(t *testing.T) {
	Convey("Given httpClient.Do returns an error", t, func() {

		body, _ := json.Marshal(validInstance)
		respBodyReader, _, httpClient := setup(body, http.StatusOK)

		httpClient.DoFunc = func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("Boom!")
		}

		cli := &DatasetAPIClient{
			host:           host,
			authToken:      auth,
			responseReader: respBodyReader,
			httpClient:     httpClient,
		}

		Convey("When GetInstance is invoked", func() {

			i, err := cli.GetInstance(testInstanceID)

			Convey("Then the returned values are as expected", func() {
				So(i, ShouldBeNil)
				So(err, ShouldResemble, errors.New("Boom!"))
			})

			Convey("And responseBodyReader is never invoked", func() {
				So(len(respBodyReader.ReadCalls()), ShouldEqual, 0)
			})
		})
	})
}

func TestDatasetAPIClient_GetInstance_HttpStatus(t *testing.T) {
	Convey("Given httpClient.Do returns an non 200 status", t, func() {

		body, _ := json.Marshal(validInstance)
		respBodyReader, response, httpClient := setup(body, http.StatusOK)
		response.StatusCode = http.StatusBadRequest

		httpClient.DoFunc = func(req *http.Request) (*http.Response, error) {
			return response, nil
		}

		cli := &DatasetAPIClient{
			host:           host,
			authToken:      auth,
			responseReader: respBodyReader,
			httpClient:     httpClient,
		}

		Convey("When GetInstance is invoked", func() {

			i, err := cli.GetInstance(testInstanceID)

			Convey("Then the returned values are as expected", func() {
				So(i, ShouldBeNil)
				So(err, ShouldResemble, errors.New("unexpected status code returned from dataset cli"))
			})

			Convey("And responseBodyReader is never invoked", func() {
				So(len(respBodyReader.ReadCalls()), ShouldEqual, 0)
			})
		})
	})
}

func TestDatasetAPIClient_GetInstance_ResponseBodyReadErr(t *testing.T) {
	Convey("Given responseBodyReader returns an error", t, func() {

		body, _ := json.Marshal(validInstance)
		respBodyReader, response, httpClient := setup(body, http.StatusOK)
		respBodyReader.ReadFunc = func(r io.Reader) ([]byte, error) {
			return nil, errors.New("Bork!")
		}

		httpClient.DoFunc = func(req *http.Request) (*http.Response, error) {
			return response, nil
		}

		cli := &DatasetAPIClient{
			host:           host,
			authToken:      auth,
			responseReader: respBodyReader,
			httpClient:     httpClient,
		}

		Convey("When GetInstance is invoked", func() {

			i, err := cli.GetInstance(testInstanceID)

			Convey("Then the returned values are as expected", func() {
				So(i, ShouldBeNil)
				So(err, ShouldResemble, errors.New("Bork!"))
			})

			Convey("And responseBodyReader is called 1 time with the expected parameters", func() {
				So(len(respBodyReader.ReadCalls()), ShouldEqual, 1)
				So(respBodyReader.ReadCalls()[0].R, ShouldResemble, response.Body)
			})
		})
	})
}

func TestDatasetAPIClient_GetInstance_UnmarshallErr(t *testing.T) {
	Convey("Given unmarshalling the response body returns an error", t, func() {

		body := []byte("This is not a valid response")
		respBodyReader, response, httpClient := setup(body, http.StatusOK)

		httpClient.DoFunc = func(req *http.Request) (*http.Response, error) {
			return response, nil
		}

		cli := &DatasetAPIClient{
			host:           host,
			authToken:      auth,
			responseReader: respBodyReader,
			httpClient:     httpClient,
		}

		Convey("When GetInstance is invoked", func() {

			i, err := cli.GetInstance(testInstanceID)

			Convey("Then the returned values are as expected", func() {
				So(i, ShouldBeNil)

				expectedType := reflect.TypeOf((*json.SyntaxError)(nil))
				actualType := reflect.TypeOf(err)
				So(actualType, ShouldEqual, expectedType)
			})

			Convey("And responseBodyReader is called 1 time with the expected parameters", func() {
				So(len(respBodyReader.ReadCalls()), ShouldEqual, 1)
				So(respBodyReader.ReadCalls()[0].R, ShouldResemble, response.Body)
			})
		})
	})
}

func TestNewDatasetAPIClient(t *testing.T) {
	responseBodyReader, _, httpClient := setup([]byte{}, http.StatusOK)

	Convey("Given an invalid host", t, func() {

		Convey("When NewDatasetAPIClient is called", func() {

			cli, err := NewDatasetAPIClient("", "", nil, nil)

			Convey("Then the expect values are returned", func() {
				So(cli, ShouldBeNil)
				So(err, ShouldResemble, errors.New("datasetAPIClient requires a non empty host"))
			})
		})
	})

	Convey("Given an invalid authToken", t, func() {

		Convey("When NewDatasetAPIClient is called", func() {

			cli, err := NewDatasetAPIClient(host, "", nil, nil)

			Convey("Then the expect values are returned", func() {
				So(cli, ShouldBeNil)
				So(err, ShouldResemble, errors.New("datasetAPIClient requires a non empty authToken"))
			})
		})
	})

	Convey("Given an invalid httpClient", t, func() {

		Convey("When NewDatasetAPIClient is called", func() {

			cli, err := NewDatasetAPIClient(host, auth, nil, nil)

			Convey("Then the expect values are returned", func() {
				So(cli, ShouldBeNil)
				So(err, ShouldResemble, errors.New("datasetAPIClient requires a non nil HttpClient"))
			})
		})
	})

	Convey("Given an invalid ResponseBodyReader", t, func() {

		Convey("When NewDatasetAPIClient is called", func() {

			cli, err := NewDatasetAPIClient(host, auth, httpClient, nil)

			Convey("Then the expect values are returned", func() {
				So(cli, ShouldBeNil)
				So(err, ShouldResemble, errors.New("datasetAPIClient requires a non nil ResponseBodyReader"))
			})
		})
	})

	Convey("Given an valid parameters", t, func() {

		Convey("When NewDatasetAPIClient is called", func() {

			cli, err := NewDatasetAPIClient(host, auth, httpClient, responseBodyReader)

			Convey("Then the expect values are returned", func() {
				So(err, ShouldBeNil)
				So(cli.host, ShouldEqual, host)
				So(cli.authToken, ShouldEqual, auth)
				So(cli.httpClient, ShouldEqual, httpClient)
				So(cli.responseReader, ShouldEqual, responseBodyReader)
			})
		})
	})
}

func setup(body []byte, status int) (*mocks.ResponseBodyReaderMock, *http.Response, *mocks.HttpClientMock) {
	reader := bytes.NewReader(body)
	readeCloser := ioutil.NopCloser(reader)

	respBodyReader := &mocks.ResponseBodyReaderMock{
		ReadFunc: func(r io.Reader) ([]byte, error) {
			return body, nil
		},
	}

	response := &http.Response{
		Body:       readeCloser,
		StatusCode: status,
	}

	httpClient := &mocks.HttpClientMock{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return response, nil
		},
	}

	return respBodyReader, response, httpClient
}
