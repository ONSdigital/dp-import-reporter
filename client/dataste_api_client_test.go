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
	"strings"
	"fmt"
)

const (
	host           = "http://localhost:8080"
	auth           = "secret_password"
	testInstanceID = "1234567890"
)

var (
	event = &model.Event{
		Message:       "Error",
		Type:          "Error",
		MessageOffset: "0",
	}

	validInstance = &model.Instance{
		State:  "RED",
		Events: []*model.Event{event},
	}
)

func TestDatasetAPIClient_GetInstance(t *testing.T) {
	Convey("Given a correctly configured DatasetAPIClient", t, func() {
		body, _ := json.Marshal(validInstance)
		respBodyReader, _, httpClient, cli := setup(body, http.StatusOK)

		Convey("When GetInstance is called with valid parameters", func() {
			i, err := cli.GetInstance(testInstanceID)

			Convey("Then the returned values are as expected", func() {
				So(i, ShouldResemble, validInstance)
				So(err, ShouldBeNil)
			})

			Convey("And httpClient.Do is called 1 time with the expected parameters", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				So(httpClient.DoCalls()[0].Req.URL.String(), ShouldEqual, "http://localhost:8080/instances/"+testInstanceID)
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
		respBodyReader, _, httpClient, cli := setup(body, http.StatusOK)
		httpCliErr := errors.New("Boom!")

		httpClient.DoFunc = func(req *http.Request) (*http.Response, error) {
			return nil, httpCliErr
		}

		Convey("When GetInstance is invoked", func() {

			i, err := cli.GetInstance(testInstanceID)

			Convey("Then the returned values are as expected", func() {
				So(i, ShouldBeNil)
				So(err, ShouldResemble, wrappedDatasetAPIError(httpClientDoErr, httpCliErr))
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
		respBodyReader, response, httpClient, cli := setup(body, http.StatusBadRequest)

		httpClient.DoFunc = func(req *http.Request) (*http.Response, error) {
			return response, nil
		}

		Convey("When GetInstance is invoked", func() {

			i, err := cli.GetInstance(testInstanceID)

			Convey("Then the returned values are as expected", func() {
				So(i, ShouldBeNil)
				So(err, ShouldResemble, newDatasetAPIError(unexpectedHTTPStatus))
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
		readBodyErr := errors.New("Bork!")

		respBodyReader, response, httpClient, cli := setup(body, http.StatusOK)
		respBodyReader.ReadFunc = func(r io.Reader) ([]byte, error) {
			return nil, readBodyErr
		}

		httpClient.DoFunc = func(req *http.Request) (*http.Response, error) {
			return response, nil
		}

		Convey("When GetInstance is invoked", func() {

			i, err := cli.GetInstance(testInstanceID)

			Convey("Then the returned values are as expected", func() {
				So(i, ShouldBeNil)
				So(err, ShouldResemble, wrappedDatasetAPIError(readResponseBodyErr, readBodyErr))
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
		respBodyReader, response, httpClient, cli := setup(body, http.StatusOK)

		httpClient.DoFunc = func(req *http.Request) (*http.Response, error) {
			return response, nil
		}

		Convey("When GetInstance is invoked", func() {

			i, err := cli.GetInstance(testInstanceID)

			Convey("Then the returned values are as expected", func() {
				So(i, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(strings.Contains(err.Error(), unmarshalResponseErr), ShouldBeTrue)
			})

			Convey("And responseBodyReader is called 1 time with the expected parameters", func() {
				So(len(respBodyReader.ReadCalls()), ShouldEqual, 1)
				So(respBodyReader.ReadCalls()[0].R, ShouldResemble, response.Body)
			})
		})
	})
}

func TestDatasetAPIClient_AddEventToInstance_invalidParams(t *testing.T) {
	Convey("Given instanceID is empty", t, func() {
		_, _, httpClient, cli := setup(nil, http.StatusOK)

		Convey("When AddEventToInstance is called", func() {
			err := cli.AddEventToInstance("", nil)

			Convey("Then the DatasetAPI returns the expected error", func() {
				So(err, ShouldResemble, newDatasetAPIError(instanceIDNil))
			})

			Convey("And httpClient.Do is never called", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 0)
			})
		})
	})

	Convey("Given event is nil", t, func() {
		_, _, httpClient, cli := setup(nil, http.StatusOK)

		Convey("When AddEventToInstance is called", func() {
			err := cli.AddEventToInstance(testInstanceID, nil)

			Convey("Then the DatasetAPI returns the expected error", func() {
				So(err, ShouldResemble, newDatasetAPIError(eventNil))
			})

			Convey("And httpClient.Do is never called", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 0)
			})
		})
	})
}

func TestDatasetAPIClient_AddEventToInstance_HttpCliErr(t *testing.T) {
	Convey("Given datasetAPIClient has been configured correctly", t, func() {
		_, _, httpClient, cli := setup(nil, http.StatusOK)

		httpCliErr := errors.New("Wubba dubba dub dub")
		httpClient.DoFunc = func(req *http.Request) (*http.Response, error) {
			return nil, httpCliErr
		}

		Convey("When httpClient.Do returns an error", func() {
			err := cli.AddEventToInstance(testInstanceID, event)

			Convey("Then the expected error is returned", func() {
				So(err, ShouldResemble, wrappedDatasetAPIError(httpClientDoErr, httpCliErr))
			})

			Convey("And httpClient.Do is called 1 time", func() {
				calls := httpClient.DoCalls()
				So(len(calls), ShouldEqual, 1)
				So(calls[0].Req.URL.String(), ShouldEqual, fmt.Sprintf(addInstanceEventURL, host, testInstanceID))
			})
		})
	})
}

func TestDatasetAPIClient_AddEventToInstance_UnexpectedStatus(t *testing.T) {
	Convey("Given datasetAPIClient has been configured correctly", t, func() {
		_, _, httpClient, cli := setup(nil, http.StatusBadRequest)

		Convey("When the returned HTTP status is not 201 CREATED", func() {

			err := cli.AddEventToInstance(testInstanceID, event)

			Convey("Then the expected error is returned", func() {
				So(err, ShouldResemble, newDatasetAPIError(unexpectedHTTPStatus))
			})

			Convey("And httpClient.Do is called 1 time", func() {
				calls := httpClient.DoCalls()
				So(len(calls), ShouldEqual, 1)
				So(calls[0].Req.URL.String(), ShouldEqual, fmt.Sprintf(addInstanceEventURL, host, testInstanceID))
			})
		})
	})
}

func TestDatasetAPIClient_AddEventToInstance(t *testing.T) {
	Convey("Given datasetAPIClient has been configured correctly", t, func() {
		_, _, httpClient, cli := setup(nil, http.StatusCreated)

		Convey("When AddEventToInstance is called", func() {
			err := cli.AddEventToInstance(testInstanceID, event)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And httpClient.Do is called 1 time", func() {
				calls := httpClient.DoCalls()
				So(len(calls), ShouldEqual, 1)
				So(calls[0].Req.URL.String(), ShouldEqual, fmt.Sprintf(addInstanceEventURL, host, testInstanceID))
			})
		})
	})
}

func TestDatasetAPIClient_UpdateInstanceStatus_InvalidParams(t *testing.T) {
	Convey("Given datasetAPIClient has been configured correctly", t, func() {
		_, _, httpClient, cli := setup(nil, 0)

		Convey("When UpdateInstanceStatus is called with an empty instanceID", func() {
			err := cli.UpdateInstanceStatus("", nil)

			Convey("Then the expected error is returned", func() {
				So(err, ShouldResemble, newDatasetAPIError(instanceIDNil))
			})

			Convey("And httpClient.Do is never invoked", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 0)
			})
		})

		Convey("When UpdateInstanceStatus is called with an nil state", func() {
			err := cli.UpdateInstanceStatus(testInstanceID, nil)

			Convey("Then the expected error is returned", func() {
				So(err, ShouldResemble, newDatasetAPIError(stateNil))
			})

			Convey("And httpClient.Do is never invoked", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 0)
			})
		})
	})
}

func TestDatasetAPIClient_UpdateInstanceStatus(t *testing.T) {
	Convey("Given datasetAPIClient has been configured correctly", t, func() {
		_, response, httpClient, cli := setup(nil, http.StatusOK)

		Convey("When UpdateInstanceStatus is called", func() {
			err := cli.UpdateInstanceStatus(testInstanceID, &model.State{State: "failed"})

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And httpClient.Do is called 1 time with the expected parameters", func() {
				calls := httpClient.DoCalls()
				So(len(calls), ShouldEqual, 1)
				So(calls[0].Req.URL.String(), ShouldEqual, fmt.Sprintf(putInstanceStateURL, host, testInstanceID))
			})
		})

		Convey("When httpClient.Do returns an error", func() {
			httpCliDoErr := errors.New("Bork!")
			httpClient.DoFunc = func(req *http.Request) (*http.Response, error) {
				return nil, httpCliDoErr
			}
			err := cli.UpdateInstanceStatus(testInstanceID, &model.State{State: "failed"})

			Convey("Then the expected error is returned", func() {
				So(err, ShouldResemble, wrappedDatasetAPIError(httpClientDoErr, httpCliDoErr))
			})

			Convey("And httpClient.Do is called 1 time with the expected parameters", func() {
				calls := httpClient.DoCalls()
				So(len(calls), ShouldEqual, 1)
				So(calls[0].Req.URL.String(), ShouldEqual, fmt.Sprintf(putInstanceStateURL, host, testInstanceID))
			})
		})

		Convey("When httpClient.Do returns an incorrect HTTP status", func() {
			response.StatusCode = http.StatusBadRequest
			err := cli.UpdateInstanceStatus(testInstanceID, &model.State{State: "failed"})

			Convey("Then the expected error is returned", func() {
				So(err, ShouldResemble, newDatasetAPIError(unexpectedHTTPStatus))
			})

			Convey("And httpClient.Do is called 1 time with the expected parameters", func() {
				calls := httpClient.DoCalls()
				So(len(calls), ShouldEqual, 1)
				So(calls[0].Req.URL.String(), ShouldEqual, fmt.Sprintf(putInstanceStateURL, host, testInstanceID))
			})

		})
	})
}

func TestNewDatasetAPIClient(t *testing.T) {
	responseBodyReader, _, httpClient, _ := setup([]byte{}, http.StatusOK)

	Convey("Given an invalid host", t, func() {

		Convey("When NewDatasetAPIClient is called", func() {

			cli, err := NewDatasetAPIClient("", "", nil, nil)

			Convey("Then the expect values are returned", func() {
				So(cli, ShouldBeNil)
				So(err, ShouldResemble, newDatasetAPIError(hostEmpty))
			})
		})
	})

	Convey("Given an invalid authToken", t, func() {

		Convey("When NewDatasetAPIClient is called", func() {

			cli, err := NewDatasetAPIClient(host, "", nil, nil)

			Convey("Then the expect values are returned", func() {
				So(cli, ShouldBeNil)
				So(err, ShouldResemble, newDatasetAPIError(authTokenEmpty))
			})
		})
	})

	Convey("Given an invalid httpClient", t, func() {

		Convey("When NewDatasetAPIClient is called", func() {

			cli, err := NewDatasetAPIClient(host, auth, nil, nil)

			Convey("Then the expect values are returned", func() {
				So(cli, ShouldBeNil)
				So(err, ShouldResemble, newDatasetAPIError(httpClientNil))
			})
		})
	})

	Convey("Given an invalid ResponseBodyReader", t, func() {

		Convey("When NewDatasetAPIClient is called", func() {

			cli, err := NewDatasetAPIClient(host, auth, httpClient, nil)

			Convey("Then the expect values are returned", func() {
				So(cli, ShouldBeNil)
				So(err, ShouldResemble, newDatasetAPIError(responseBodyReaderNil))
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

func setup(body []byte, status int) (*mocks.ResponseBodyReaderMock, *http.Response, *mocks.HttpClientMock, *DatasetAPIClient) {
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

	cli := &DatasetAPIClient{
		host:           host,
		authToken:      auth,
		responseReader: respBodyReader,
		httpClient:     httpClient,
	}
	return respBodyReader, response, httpClient, cli
}
