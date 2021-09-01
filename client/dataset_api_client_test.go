package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/ONSdigital/dp-import-reporter/model"
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

	errMock = errors.New("boom")
)

var ctx = context.Background()

func TestDatasetAPIClientGetInstance(t *testing.T) {
	Convey("Given a correctly configured DatasetAPIClient", t, func() {
		body, _ := json.Marshal(validInstance)
		respBodyReader, _, httpClient, cli := setup(body, http.StatusOK)

		Convey("When GetInstance is called with valid parameters", func() {
			i, err := cli.GetInstance(ctx, testInstanceID)

			Convey("Then the returned values are as expected", func() {
				So(i, ShouldResemble, validInstance)
				So(err, ShouldBeNil)
			})

			Convey("And httpClient.Do is called 1 time with the expected parameters", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				So(httpClient.DoCalls()[0].Req.URL.String(), ShouldEqual, "http://localhost:8080/instances/"+testInstanceID)
				So(httpClient.DoCalls()[0].Req.Header.Get(authorizationHeader), ShouldEqual, auth)
			})

			Convey("And responseBodyReader.Read is called 1 time with the expected parameters", func() {
				So(len(respBodyReader.ReadCalls()), ShouldEqual, 1)
			})
		})
	})
}

func TestDatasetAPIClientGetInstanceHttpCliErr(t *testing.T) {
	Convey("Given httpClient.Do returns an error", t, func() {

		body, _ := json.Marshal(validInstance)
		respBodyReader, _, httpClient, cli := setup(body, http.StatusOK)

		httpClient.DoFunc = func(req *http.Request) (*http.Response, error) {
			return nil, errMock
		}

		Convey("When GetInstance is invoked", func() {

			i, err := cli.GetInstance(ctx, testInstanceID)

			Convey("Then the returned values are as expected", func() {
				So(i, ShouldBeNil)
				So(err.Error(), ShouldEqual, errors.Wrap(errMock, "GetInstance HTTPClient.doRequest returned an error").Error())
			})

			Convey("And httpClient.Do is called 1 time with the expected parameters", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				So(httpClient.DoCalls()[0].Req.URL.String(), ShouldEqual, "http://localhost:8080/instances/"+testInstanceID)
				So(httpClient.DoCalls()[0].Req.Header.Get(authorizationHeader), ShouldEqual, auth)
			})

			Convey("And responseBodyReader is never invoked", func() {
				So(len(respBodyReader.ReadCalls()), ShouldEqual, 0)
			})
		})
	})
}

func TestDatasetAPIClientGetInstanceHttpStatus(t *testing.T) {
	Convey("Given httpClient.Do returns an non 200 status", t, func() {

		body, _ := json.Marshal(validInstance)
		respBodyReader, response, httpClient, cli := setup(body, http.StatusBadRequest)

		httpClient.DoFunc = func(req *http.Request) (*http.Response, error) {
			return response, nil
		}

		Convey("When GetInstance is invoked", func() {

			i, err := cli.GetInstance(ctx, testInstanceID)

			Convey("Then the returned values are as expected", func() {
				So(i, ShouldBeNil)

				url := fmt.Sprintf(getInstanceURL, host, testInstanceID)
				So(err.Error(), ShouldEqual, incorrectStatusError("GetInstance", url, http.MethodGet, 200, http.StatusBadRequest).Error())
			})

			Convey("And httpClient.Do is called 1 time with the expected parameters", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				So(httpClient.DoCalls()[0].Req.URL.String(), ShouldEqual, "http://localhost:8080/instances/"+testInstanceID)
				So(httpClient.DoCalls()[0].Req.Header.Get(authorizationHeader), ShouldEqual, auth)
			})

			Convey("And responseBodyReader is never invoked", func() {
				So(len(respBodyReader.ReadCalls()), ShouldEqual, 0)
			})
		})
	})
}

func TestDatasetAPIClientGetInstanceResponseBodyReadErr(t *testing.T) {
	Convey("Given responseBodyReader returns an error", t, func() {

		body, _ := json.Marshal(validInstance)

		respBodyReader, response, httpClient, cli := setup(body, http.StatusOK)
		respBodyReader.ReadFunc = func(r io.Reader) ([]byte, error) {
			return nil, errMock
		}

		httpClient.DoFunc = func(req *http.Request) (*http.Response, error) {
			return response, nil
		}

		Convey("When GetInstance is invoked", func() {

			i, err := cli.GetInstance(ctx, testInstanceID)

			Convey("Then the returned values are as expected", func() {
				So(i, ShouldBeNil)
				So(err.Error(), ShouldEqual, errors.Wrap(errMock, "GetInstance error while attempting to read HTTP response body").Error())
			})

			Convey("And httpClient.Do is called 1 time with the expected parameters", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				So(httpClient.DoCalls()[0].Req.URL.String(), ShouldEqual, "http://localhost:8080/instances/"+testInstanceID)
				So(httpClient.DoCalls()[0].Req.Header.Get(authorizationHeader), ShouldEqual, auth)
			})

			Convey("And responseBodyReader is called 1 time with the expected parameters", func() {
				So(len(respBodyReader.ReadCalls()), ShouldEqual, 1)
				So(respBodyReader.ReadCalls()[0].R, ShouldResemble, response.Body)
			})
		})
	})
}

func TestDatasetAPIClientGetInstanceUnmarshallErr(t *testing.T) {
	Convey("Given that unmarshalling the response body returns an error", t, func() {

		body := []byte("This is not a valid response")
		respBodyReader, response, httpClient, cli := setup(body, http.StatusOK)

		httpClient.DoFunc = func(req *http.Request) (*http.Response, error) {
			return response, nil
		}

		Convey("When GetInstance is invoked", func() {

			i, err := cli.GetInstance(ctx, testInstanceID)

			Convey("Then the returned values are as expected", func() {
				So(i, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(strings.Contains(err.Error(), "GetInstance error while attempting to unmarshal HTTP response body into model.Instance"), ShouldBeTrue)
			})

			Convey("And httpClient.Do is called 1 time with the expected parameters", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				So(httpClient.DoCalls()[0].Req.URL.String(), ShouldEqual, "http://localhost:8080/instances/"+testInstanceID)
				So(httpClient.DoCalls()[0].Req.Header.Get(authorizationHeader), ShouldEqual, auth)
			})

			Convey("And responseBodyReader is called 1 time with the expected parameters", func() {
				So(len(respBodyReader.ReadCalls()), ShouldEqual, 1)
				So(respBodyReader.ReadCalls()[0].R, ShouldResemble, response.Body)
			})
		})
	})
}

func TestDatasetAPIClientAddEventToInstanceInvalidParams(t *testing.T) {
	Convey("Given instanceID is empty", t, func() {
		_, _, httpClient, cli := setup(nil, http.StatusOK)

		Convey("When AddEventToInstance is called", func() {
			err := cli.AddEventToInstance(ctx, "", nil)

			Convey("Then the datasetAPI returns the expected error", func() {
				So(err.Error(), ShouldResemble, errors.Wrap(errValidation, "AddEventToInstance requires a non empty instanceID").Error())
			})

			Convey("And httpClient.Do is never called", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 0)
			})
		})
	})

	Convey("Given event is nil", t, func() {
		_, _, httpClient, cli := setup(nil, http.StatusOK)

		Convey("When AddEventToInstance is called", func() {
			err := cli.AddEventToInstance(ctx, testInstanceID, nil)

			Convey("Then the datasetAPI returns the expected error", func() {
				So(err.Error(), ShouldResemble, errors.Wrap(errValidation, "AddEventToInstance requires a non empty event").Error())
			})

			Convey("And httpClient.Do is never called", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 0)
			})
		})
	})
}

func TestDatasetAPIClientAddEventToInstanceHttpCliErr(t *testing.T) {
	Convey("Given datasetAPIClient has been configured correctly", t, func() {
		_, _, httpClient, cli := setup(nil, http.StatusOK)

		httpClient.DoFunc = func(req *http.Request) (*http.Response, error) {
			return nil, errMock
		}

		Convey("When httpClient.Do returns an error", func() {
			err := cli.AddEventToInstance(ctx, testInstanceID, event)

			Convey("Then the expected error is returned", func() {
				So(err.Error(), ShouldEqual, errors.Wrap(errMock, "AddEventToInstance HTTPClient.doRequest returned an error").Error())
			})

			Convey("And httpClient.Do is called 1 time", func() {
				calls := httpClient.DoCalls()
				So(len(calls), ShouldEqual, 1)
				So(calls[0].Req.URL.String(), ShouldEqual, fmt.Sprintf(addInstanceEventURL, host, testInstanceID))
				So(httpClient.DoCalls()[0].Req.Header.Get(authorizationHeader), ShouldEqual, auth)
			})
		})
	})
}

func TestDatasetAPIClientAddEventToInstanceUnexpectedStatus(t *testing.T) {
	Convey("Given datasetAPIClient has been configured correctly", t, func() {
		_, _, httpClient, cli := setup(nil, http.StatusBadRequest)

		Convey("When the returned HTTP status is not 201 CREATED", func() {

			err := cli.AddEventToInstance(ctx, testInstanceID, event)

			Convey("Then the expected error is returned", func() {
				url := fmt.Sprintf(addInstanceEventURL, host, testInstanceID)
				So(err.Error(), ShouldEqual, incorrectStatusError("AddEventToInstance", url, http.MethodPost, http.StatusCreated, http.StatusBadRequest).Error())
			})

			Convey("And httpClient.Do is called 1 time", func() {
				calls := httpClient.DoCalls()
				So(len(calls), ShouldEqual, 1)
				So(calls[0].Req.URL.String(), ShouldEqual, fmt.Sprintf(addInstanceEventURL, host, testInstanceID))
				So(httpClient.DoCalls()[0].Req.Header.Get(authorizationHeader), ShouldEqual, auth)
			})
		})
	})
}

func TestDatasetAPIClientAddEventToInstance(t *testing.T) {
	Convey("Given datasetAPIClient has been configured correctly", t, func() {
		_, _, httpClient, cli := setup(nil, http.StatusCreated)

		Convey("When AddEventToInstance is called", func() {
			err := cli.AddEventToInstance(ctx, testInstanceID, event)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And httpClient.Do is called 1 time", func() {
				calls := httpClient.DoCalls()
				So(len(calls), ShouldEqual, 1)
				So(calls[0].Req.URL.String(), ShouldEqual, fmt.Sprintf(addInstanceEventURL, host, testInstanceID))
				So(httpClient.DoCalls()[0].Req.Header.Get(authorizationHeader), ShouldEqual, auth)
			})
		})
	})
}

func TestDatasetAPIClientUpdateInstanceStatusInvalidParams(t *testing.T) {
	Convey("Given datasetAPIClient has been configured correctly", t, func() {
		_, _, httpClient, cli := setup(nil, 0)

		Convey("When UpdateInstanceStatus is called with an empty instanceID", func() {
			err := cli.UpdateInstanceStatus(ctx, "", nil)

			Convey("Then the expected error is returned", func() {
				So(err.Error(), ShouldEqual, errors.Wrap(errValidation, "UpdateInstanceStatus requires a non empty instanceID").Error())
			})

			Convey("And httpClient.Do is never invoked", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 0)
			})
		})

		Convey("When UpdateInstanceStatus is called with an nil state", func() {
			err := cli.UpdateInstanceStatus(ctx, testInstanceID, nil)

			Convey("Then the expected error is returned", func() {
				So(err.Error(), ShouldEqual, errors.Wrap(errValidation, "UpdateInstanceStatus requires a non nil state").Error())
			})

			Convey("And httpClient.Do is never invoked", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 0)
			})
		})
	})
}

func TestDatasetAPIClientUpdateInstanceStatus(t *testing.T) {
	Convey("Given datasetAPIClient has been configured correctly", t, func() {
		_, response, httpClient, cli := setup(nil, http.StatusOK)

		Convey("When UpdateInstanceStatus is called", func() {
			err := cli.UpdateInstanceStatus(ctx, testInstanceID, &model.State{State: "failed"})

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And httpClient.Do is called once with the expected parameters", func() {
				calls := httpClient.DoCalls()
				So(len(calls), ShouldEqual, 1)
				So(calls[0].Req.URL.String(), ShouldEqual, fmt.Sprintf(putInstanceStateURL, host, testInstanceID))
			})
		})

		Convey("When httpClient.Do returns an error", func() {
			httpClient.DoFunc = func(req *http.Request) (*http.Response, error) {
				return nil, errMock
			}
			err := cli.UpdateInstanceStatus(ctx, testInstanceID, &model.State{State: "failed"})

			Convey("Then the expected error is returned", func() {
				So(err.Error(), ShouldEqual, errors.Wrap(errMock, "UpdateInstanceStatus HTTPClient.doRequest returned an error").Error())
			})

			Convey("And httpClient.Do is called 1 time with the expected parameters", func() {
				calls := httpClient.DoCalls()
				So(len(calls), ShouldEqual, 1)
				So(calls[0].Req.URL.String(), ShouldEqual, fmt.Sprintf(putInstanceStateURL, host, testInstanceID))
				So(httpClient.DoCalls()[0].Req.Header.Get(authorizationHeader), ShouldEqual, auth)
			})
		})

		Convey("When httpClient.Do returns an incorrect HTTP status", func() {
			response.StatusCode = http.StatusBadRequest
			err := cli.UpdateInstanceStatus(ctx, testInstanceID, &model.State{State: "failed"})

			Convey("Then the expected error is returned", func() {
				url := fmt.Sprintf(putInstanceStateURL, host, testInstanceID)
				So(err.Error(), ShouldEqual, incorrectStatusError("UpdateInstanceStatus", url, http.MethodPut, http.StatusOK, http.StatusBadRequest).Error())
			})

			Convey("And httpClient.Do is called 1 time with the expected parameters", func() {
				calls := httpClient.DoCalls()
				So(len(calls), ShouldEqual, 1)
				So(calls[0].Req.URL.String(), ShouldEqual, fmt.Sprintf(putInstanceStateURL, host, testInstanceID))
				So(httpClient.DoCalls()[0].Req.Header.Get(authorizationHeader), ShouldEqual, auth)
			})

		})
	})
}

func TestNewDatasetAPIClient(t *testing.T) {
	responseBodyReader, _, httpClient, _ := setup([]byte{}, http.StatusOK)

	Convey("Given an invalid service authToken", t, func() {

		Convey("When NewDatasetAPIClient is called", func() {

			cli, err := NewDatasetAPIClient("", "", "", nil, nil)

			Convey("Then the expect values are returned", func() {
				So(cli, ShouldBeNil)
				So(err.Error(), ShouldEqual, errors.Wrap(errValidation, "non empty service authToken required").Error())
			})
		})
	})

	Convey("Given an invalid host", t, func() {

		Convey("When NewDatasetAPIClient is called", func() {

			cli, err := NewDatasetAPIClient(auth, "", "", nil, nil)

			Convey("Then the expect values are returned", func() {
				So(cli, ShouldBeNil)
				So(err.Error(), ShouldEqual, errors.Wrap(errValidation, "non empty host required").Error())
			})
		})
	})

	Convey("Given an invalid authToken", t, func() {

		Convey("When NewDatasetAPIClient is called", func() {

			cli, err := NewDatasetAPIClient(auth, host, "", nil, nil)

			Convey("Then the expect values are returned", func() {
				So(cli, ShouldBeNil)
				So(err.Error(), ShouldEqual, errors.Wrap(errValidation, "non empty dataset API authToken required").Error())
			})
		})
	})

	Convey("Given an invalid httpClient", t, func() {

		Convey("When NewDatasetAPIClient is called", func() {

			cli, err := NewDatasetAPIClient(auth, host, auth, nil, nil)

			Convey("Then the expect values are returned", func() {
				So(cli, ShouldBeNil)
				So(err.Error(), ShouldEqual, errors.Wrap(errValidation, "non nil HTTPClient required").Error())
			})
		})
	})

	Convey("Given an invalid ResponseBodyReader", t, func() {

		Convey("When NewDatasetAPIClient is called", func() {

			cli, err := NewDatasetAPIClient(auth, host, auth, httpClient, nil)

			Convey("Then the expect values are returned", func() {
				So(cli, ShouldBeNil)
				So(err.Error(), ShouldEqual, errors.Wrap(errValidation, "non nil ResponseBodyReader required").Error())
			})
		})
	})

	Convey("Given an valid parameters", t, func() {

		Convey("When NewDatasetAPIClient is called", func() {

			cli, err := NewDatasetAPIClient(auth, host, auth, httpClient, responseBodyReader)

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

func setup(body []byte, status int) (*ResponseBodyReaderMock, *http.Response, *HTTPClientMock, *DatasetAPIClient) {
	reader := bytes.NewReader(body)
	readeCloser := ioutil.NopCloser(reader)

	respBodyReader := &ResponseBodyReaderMock{
		ReadFunc: func(r io.Reader) ([]byte, error) {
			return body, nil
		},
	}

	response := &http.Response{
		Body:       readeCloser,
		StatusCode: status,
	}

	httpClient := &HTTPClientMock{
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
