package handler

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"runtime/debug"
	"testing"
	"time"

	"github.com/ONSdigital/dp-import-reporter/config"
	"github.com/coocood/freecache"
	. "github.com/smartystreets/goconvey/convey"
)

//Correct
var cfg = &config.Config{
	NewInstanceTopic: "event-reporter",
	Brokers:          []string{"localhost:9092"},
	ImportAPIURL:     "http://localhost:21800",
	ImportAuthToken:  "FD0108EA-825D-411C-9B1D-41EF7727F465",
	BindAddress:      ":22200",
}

//wrong ImportAPI
var cfgBadURL = &config.Config{
	NewInstanceTopic: "event-reporter",
	Brokers:          []string{"localhost:9092"},
	ImportAPIURL:     "http://localho:21800",
	ImportAuthToken:  "FD0108EA-825D-411C-9B1D-41EF7727F465",
	BindAddress:      ":22200",
	// ImportAPIURL: "http://localhost:21800",
}

//wrong auth token
var cfgBadAuth = &config.Config{
	NewInstanceTopic: "event-reporter",
	Brokers:          []string{"localhost:9092"},
	ImportAPIURL:     "http://localhost:21800",
	ImportAuthToken:  "D0108EA-825D-411C-9B12-41EF7727F465",
	BindAddress:      ":22200",
}

//correct
var e = EventReport{
	InstanceID: "a4695fee-f0a2-49c4-b136-e3ca8dd40476",
	EventType:  "error",
	EventMsg:   "Broken on something.",
}

//wrong instance
var eWrongInstance = EventReport{
	InstanceID: "a4695fee-f0a2-49c4-b136-eWithRandomMsgca8dd40612",
	EventType:  "error",
	EventMsg:   "Broken on something.",
}
var eWrongInstanceMsg = EventReport{
	InstanceID: "a4695fee-f0a2-49c4-b136-eWithRandomMsgca8dd40612",
	EventType:  "error",
	EventMsg:   "Broken on ",
}
var eWithRandomMsg = EventReport{
	InstanceID: "a4695fee-f0a2-49c4-b136-e3ca8dd40476",
	EventType:  "error",
	EventMsg:   "Broken on something." + string(rand.Int()),
}
var httpClient = &http.Client{}

func TestCheckInstance(t *testing.T) {

	Convey("Given when you need to check an instance ", t, func() {
		Convey("When I pass through instance information", func() {

			_, _, err := e.checkInstance(httpClient, cfgBadURL)
			Convey("URL should not parse", func() {
				So(err, ShouldNotBeNil)
			})

			state, events, err := e.checkInstance(httpClient, cfg)
			Convey("Complete run through with 200 status response", func() {
				So(err, ShouldBeNil)
				So(state, ShouldNotBeNil)
				So(events, ShouldNotBeNil)
			})

			stateWrongInstance, events1, err := eWrongInstance.checkInstance(httpClient, cfgBadURL)
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

			config, _ := config.Get()

			err := e.putJobStatus(httpClient, config)
			Convey("A complete working run through of all the code in a positive manner", func() {
				So(err, ShouldBeNil)
			})

			err1 := e.putJobStatus(httpClient, cfgBadURL)
			Convey("A run through with an incomplete url", func() {
				So(err1, ShouldNotBeNil)
			})

			err2 := e.putJobStatus(httpClient, cfgBadAuth)
			Convey("A run through without the auth token", func() {
				So(err2, ShouldNotBeNil)
			})
		})
	})
}
func TestPutEvents(t *testing.T) {
	t.Parallel()

	Convey("Given when an event needs to be added to an instance", t, func() {
		Convey("When the instance information is passed through", func() {

			json, JSONerr := json.Marshal(Event{
				Type:          e.EventType,
				Time:          time.Now().String(),
				Message:       e.EventMsg,
				MessageOffset: "0",
			})
			Convey("No errors when marshalling an event", func() {
				So(JSONerr, ShouldBeNil)
			})
			err := e.putEvent(httpClient, json, cfg, "")
			Convey("A complete run through with a postive response with it being added", func() {
				So(err, ShouldBeNil)
			})
			err1 := e.putEvent(httpClient, json, cfgBadAuth, "")
			Convey("Should through a status code error as it doesnt have authorisation", func() {
				So(err1, ShouldNotBeNil)
			})
			err2 := e.putEvent(httpClient, json, cfgBadURL, "")
			Convey("Should throw an error when trying to request the the put job status within the putevent method", func() {
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
			err := e.HandleEvent(httpClient, c, cfg)
			Convey("Complete run through", func() {
				So(err, ShouldBeNil)
			})
			err1 := eWrongInstance.HandleEvent(httpClient, c, cfg)
			Convey("Pass through an incorrect instance ID", func() {
				So(err1, ShouldNotBeNil)
			})
			err2 := eWithRandomMsg.HandleEvent(httpClient, c, cfg)
			Convey("Should add the event to the events log ", func() {
				So(err2, ShouldBeNil)
			})
		})
	})

}
func TestErrorHandler(t *testing.T) {
	Convey("Given that a status code is provided", t, func() {
		Convey("When an http response is sent", func() {
			Convey("These response should not responde with an error", func() {
				r := errorhandler(200)
				So(r, ShouldBeNil)
				r1 := errorhandler(201)
				So(r1, ShouldBeNil)
			})
			Convey("These response should return errors", func() {
				r2 := errorhandler(404)
				So(r2, ShouldNotBeNil)
				r3 := errorhandler(401)
				So(r3, ShouldNotBeNil)
				r4 := errorhandler(400)
				So(r4, ShouldNotBeNil)
				r5 := errorhandler(60000)
				So(r5, ShouldNotBeNil)
			})

		})
	})
}

func TestArraySlicing(t *testing.T) {
	Convey("A method which slices a events array up", t, func() {
		_, events, err := e.checkInstance(httpClient, cfg)
		Convey("It brings back a valid instance", func() {
			So(err, ShouldBeNil)
		})
		var aE = instanceEvent{
			Type:          "error",
			Message:       "i am a message",
			MessageOffset: "1",
		}
		array := arraySlicing(aE, *events)
		Convey("Returns false because this event doesn't exist", func() {
			So(array, ShouldBeFalse)
		})
	})
}
