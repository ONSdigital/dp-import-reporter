package handler

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"runtime/debug"
	"testing"
	"time"

	"github.com/ONSdigital/dp-import-reporter/config"
	"github.com/ONSdigital/go-ns/errorhandler/models"
	"github.com/coocood/freecache"
	. "github.com/smartystreets/goconvey/convey"
)

//TODO all the commented out sections need to be mocked/ part of the mocking process

//Correct
var cfg = &config.Config{
	NewInstanceTopic: "event-reporter",
	Brokers:          []string{"localhost:9092"},
	DatasetAPIURL:    "http://localhost:22000",
	ImportAuthToken:  "FD0108EA-825D-411C-9B1D-41EF7727F465",
	BindAddress:      ":22200",
}

//wrong DatasetAPI
var cfgBadURL = &config.Config{
	NewInstanceTopic: "event-reporter",
	Brokers:          []string{"localhost:9092"},
	DatasetAPIURL:    "http://localho:21800",
	ImportAuthToken:  "FD0108EA-825D-411C-9B1D-41EF7727F465",
	BindAddress:      ":22200",
	// DatasetAPIURL: "http://localhost:21800",
}

//wrong auth token
var cfgBadAuth = &config.Config{
	NewInstanceTopic: "event-reporter",
	Brokers:          []string{"localhost:9092"},
	DatasetAPIURL:    "http://localhost:22000",
	ImportAuthToken:  "D0108EA-825D-411C-9B12-41EF7727F465",
	BindAddress:      ":22200",
}

//correct
var e = &errorModel.EventReport{
	InstanceID: "479dcd03-09b1-4273-b49a-58533f084add",
	EventType:  "error",
	EventMsg:   "Broken on something.",
}

//wrong instance
var eWrongInstance = &errorModel.EventReport{
	InstanceID: "a4695fee-f0a2-49c4-b136-eWithRandomMsgca8dd40612",
	EventType:  "error",
	EventMsg:   "Broken on something.",
}
var eWrongInstanceMsg = &errorModel.EventReport{
	InstanceID: "a4695fee-f0a2-49c4-b136-eWithRandomMsgca8dd40612",
	EventType:  "error",
	EventMsg:   "Broken on ",
}
var eWithRandomMsg = &errorModel.EventReport{
	InstanceID: "479dcd03-09b1-4273-b49a-58533f084add",
	EventType:  "error",
	EventMsg:   "Broken on something." + string(rand.Int()),
}
var httpClient = &http.Client{}

func TestCheckInstance(t *testing.T) {

	Convey("Given when you need to check an instance ", t, func() {
		Convey("When I pass through instance information", func() {

			_, _, err := checkInstance(cfgBadURL, e)
			Convey("URL should not parse", func() {
				So(err, ShouldNotBeNil)
			})

			// state, _, err := checkInstance(cfg, e)
			// Convey("Complete run through with 200 status response", func() {
			// 	So(err, ShouldBeNil)
			// 	So(state, ShouldNotBeNil)
			// 	// So(events, ShouldNotBeNil)
			// })

			stateWrongInstance, events1, err := checkInstance(cfgBadURL, eWrongInstance)
			Convey("Complete run through with incorrect instanceID", func() {
				So(err, ShouldNotBeNil)
				So(stateWrongInstance, ShouldEqual, "")
				So(events1, ShouldNotBeNil)
			})
		})
	})

}

func TestPutJobStatus(t *testing.T) {

	Convey("Given when an instance job status needs to change", t, func() {
		Convey("When the instance information is passed through", func() {

			err1 := putJobStatus(cfgBadURL, e)
			Convey("A run through with an incomplete url", func() {
				So(err1, ShouldNotBeNil)
			})

			err2 := putJobStatus(cfgBadAuth, e)
			Convey("A run through without the auth token", func() {
				So(err2, ShouldNotBeNil)
			})
		})
	})
}
func TestinsertEvents(t *testing.T) {
	t.Parallel()

	Convey("Given when an event needs to be added to an instance", t, func() {
		Convey("When the instance information is passed through", func() {
			timeNow := time.Now()
			json, JSONerr := json.Marshal(errorModel.Event{
				Type:          e.EventType,
				Time:          &timeNow,
				Message:       e.EventMsg,
				MessageOffset: "0",
			})
			Convey("No errors when marshalling an event", func() {
				So(JSONerr, ShouldBeNil)
			})
			err1 := insertEvent(json, cfgBadAuth, "", e)
			Convey("Should through a status code error as it doesnt have authorisation", func() {
				So(err1, ShouldNotBeNil)
			})
			err2 := insertEvent(json, cfgBadURL, "", e)
			Convey("Should throw an error when trying to request the the put job status within the insertEvent method", func() {
				So(err2, ShouldNotBeNil)
			})
		})
	})
}
func TestHandleEvents(t *testing.T) {
	t.Parallel()
	Convey("Given when a complete event comes to the Handle", t, func() {
		Convey("When an event is passed to the handler", func() {

			cacheSize := cfg.CacheSize
			c := freecache.NewCache(cacheSize)
			debug.SetGCPercent(20)
			// err := HandleEvent(c, cfg, e)
			// Convey("Complete run through", func() {
			// 	So(err, ShouldBeNil)
			// })
			err1 := HandleEvent(c, cfg, eWrongInstance)
			Convey("Pass through an incorrect instance ID", func() {
				So(err1, ShouldNotBeNil)
			})
			// err2 := HandleEvent(c, cfg, eWithRandomMsg)
			// Convey("Should add the event to the events log ", func() {
			// 	So(err2, ShouldBeNil)
			// })
		})
	})

}
func TestErrorHandler(t *testing.T) {
	Convey("Given that a status code is provided", t, func() {
		Convey("When an http response is sent", func() {
			Convey("These response should not responde with an error", func() {
				r := responseStatus(200)
				So(r, ShouldBeNil)
				r1 := responseStatus(201)
				So(r1, ShouldBeNil)
			})
			Convey("These response should return errors", func() {
				r2 := responseStatus(404)
				So(r2, ShouldNotBeNil)
				r3 := responseStatus(401)
				So(r3, ShouldNotBeNil)
				r4 := responseStatus(400)
				So(r4, ShouldNotBeNil)
				r5 := responseStatus(60000)
				So(r5, ShouldNotBeNil)
			})

		})
	})
}

func TestArraySlicing(t *testing.T) {
	Convey("A method which slices a events array up", t, func() {
		// _, events, err := checkInstance(cfg, e)
		// Convey("It brings back a valid instance", func() {
		// 	So(err, ShouldBeNil)
		// })
		events := make([]*errorModel.InstanceEvent, 0)
		var aE = &errorModel.InstanceEvent{
			Type:          "error",
			Message:       "i am a message",
			MessageOffset: "1",
		}
		array := arraySlicing(aE, events)
		Convey("Returns false because this event doesn't exist", func() {
			So(array, ShouldBeFalse)
		})
	})
}
