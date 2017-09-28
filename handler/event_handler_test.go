package handler

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/ONSdigital/dp-import-reporter/model"
	"github.com/ONSdigital/dp-import-reporter/mocks"
	"errors"
	"encoding/json"
)

var (
	testInstanceID = "1234567890"

	instance = &model.Instance{
		InstanceID: testInstanceID,
		Events:     make([]*model.Event, 0),
		State:      "pending",
	}

	event = &model.ReportEvent{
		InstanceID: testInstanceID,
		EventMsg:   "Its all gone horribly wrong!",
		EventType:  "error",
	}
)

func TestHandleEvent_NotInCacheOrDatasetAPI(t *testing.T) {
	Convey("Given the handle has been configured correctly", t, func() {
		datasetAPI, cacheMock := setup()

		reportEventHandler := ReportEventHandler{
			DatasetAPI:    datasetAPI,
			Cache:         cacheMock,
			ExpireSeconds: 60,
		}

		Convey("When handle is invoked with an event that's not in cacheMock or in the dataset instance.events", func() {
			err := reportEventHandler.HandleEvent(event)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the datasetAPI is called as expected with the expected parameters", func() {
				getCalls := datasetAPI.GetInstanceCalls()
				So(len(getCalls), ShouldEqual, 1)
				So(getCalls[0].InstanceID, ShouldEqual, testInstanceID)

				addCalls := datasetAPI.AddEventToInstanceCalls()
				So(len(addCalls), ShouldEqual, 1)
				So(addCalls[0].InstanceID, ShouldEqual, testInstanceID)
				So(addCalls[0].E.Message, ShouldEqual, event.EventMsg)
				So(addCalls[0].E.Type, ShouldEqual, event.EventType)
				So(addCalls[0].E.MessageOffset, ShouldEqual, "0")

				updateCalls := datasetAPI.UpdateInstanceStatusCalls()
				So(len(updateCalls), ShouldEqual, 1)
				So(updateCalls[0].State, ShouldResemble, statusFailed)
			})

			Convey("And the Cache is called as expected with the expected parameters", func() {
				So(len(cacheMock.SetCalls()), ShouldEqual, 1)

				expectedKey, _ := json.Marshal(event)
				So(cacheMock.SetCalls()[0].Key, ShouldResemble, expectedKey)
				So(cacheMock.SetCalls()[0].Value, ShouldResemble, expectedKey)
				So(cacheMock.SetCalls()[0].ExpireSeconds, ShouldResemble, 60)
			})
		})
	})
}

func TestReportEventHandler_HandleEvent_EventInCache(t *testing.T) {
	datasetAPI, cacheMock := setup()

	cacheMock.GetFunc = func(key []byte) ([]byte, error) {
		return nil, nil
	}

	Convey("Given the reportEventHandler has been correctly configured", t, func() {
		reportEventHandler := ReportEventHandler{
			DatasetAPI:    datasetAPI,
			Cache:         cacheMock,
			ExpireSeconds: 60,
		}

		var handlerErrors error

		Convey("When the cache contains the event being handled", func() {
			handlerErrors = reportEventHandler.HandleEvent(event)
		})

		Convey("Then no error is returned", func() {
			So(handlerErrors, ShouldBeNil)
		})

		Convey("And no calls are made to the DatasetAPI", func() {
			So(len(datasetAPI.GetInstanceCalls()), ShouldEqual, 0)
			So(len(datasetAPI.AddEventToInstanceCalls()), ShouldEqual, 0)
			So(len(datasetAPI.UpdateInstanceStatusCalls()), ShouldEqual, 0)
		})

		Convey("And the cache is called as expected with the correct parameters", func() {
			expectedKey, _ := json.Marshal(event)

			So(len(cacheMock.DelCalls()), ShouldEqual, 1)
			So(cacheMock.DelCalls()[0].Key, ShouldResemble, expectedKey)

			So(len(cacheMock.GetCalls()), ShouldEqual, 1)
			So(cacheMock.GetCalls()[0].Key, ShouldResemble, expectedKey)

			So(len(cacheMock.SetCalls()), ShouldEqual, 1)
			So(cacheMock.SetCalls()[0].Key, ShouldResemble, expectedKey)
			So(cacheMock.SetCalls()[0].Value, ShouldResemble, expectedKey)
			So(cacheMock.SetCalls()[0].ExpireSeconds, ShouldEqual, 60)
		})
	})
}

func setup() (*mocks.DatasetAPICliMock, *mocks.CacheMock) {
	return &mocks.DatasetAPICliMock{
		AddEventToInstanceFunc: func(instanceID string, e *model.Event) error {
			return nil
		},
		GetInstanceFunc: func(instanceID string) (*model.Instance, error) {
			return instance, nil
		},
		UpdateInstanceStatusFunc: func(instanceID string, state model.State) error {
			return nil
		},
	}, &mocks.CacheMock{
		GetFunc: func(key []byte) ([]byte, error) {
			return nil, errors.New("not found")
		},
		DelFunc: func(key []byte) bool {
			return true
		},
		SetFunc: func(key []byte, value []byte, expireSeconds int) error {
			return nil
		},
	}
}
