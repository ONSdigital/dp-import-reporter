package handler

import (
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
var cfg1 = &config.Config{
	NewInstanceTopic: "event-reporter",
	Brokers:          []string{"localhost:9092"},
	ImportAPIURL:     "http://localho:21800",
	ImportAuthToken:  "FD0108EA-825D-411C-9B1D-41EF7727F465",
	BindAddress:      ":22200",
	// ImportAPIURL: "http://localhost:21800",
}

//wrong auth token
var cfg2 = &config.Config{
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
var e1 = EventReport{
	InstanceID: "a4695fee-f0a2-49c4-b136-e3ca8dd40612",
	EventType:  "error",
	EventMsg:   "Broken on something.",
}
var e2 = EventReport{
	InstanceID: "a4695fee-f0a2-49c4-b136-e3ca8dd40612",
	EventType:  "error",
	EventMsg:   "Broken on ",
}
var e3 = EventReport{
	InstanceID: "a4695fee-f0a2-49c4-b136-e3ca8dd40476",
	EventType:  "error",
	EventMsg:   "Broken on something." + string(rand.Int()),
}
var httpClient = &http.Client{}

func TestCheckInstance(t *testing.T) {

	Convey("Internal method which checks if the instance exists and returns the status and events ", t, func() {

		_, _, err := e.checkInstance(httpClient, cfg1)
		Convey("URL should not parse", func() {
			So(err, ShouldNotBeNil)
		})

		state, events, err := e.checkInstance(httpClient, cfg)
		Convey("Complete run through with 200 status response", func() {
			So(err, ShouldBeNil)
			So(state, ShouldNotBeNil)
			So(events, ShouldNotBeNil)
		})

		state1, events1, err := e1.checkInstance(httpClient, cfg1)
		Convey("Complete run through with incorrect instanceID", func() {
			So(err, ShouldNotBeNil)
			So(state1, ShouldEqual, "")
			So(events1, ShouldNotBeNil)
		})
	})

}

func TestPutJobStatus(t *testing.T) {

	Convey("Internal method which changes the job status ", t, func() {
		cfg3, _ := config.Get()

		err := e.putJobStatus(httpClient, cfg3)
		Convey("A complete working run through of all the code in a positive manner", func() {
			So(err, ShouldBeNil)
		})

		err1 := e.putJobStatus(httpClient, cfg1)
		Convey("A run through with an incomplete url", func() {
			So(err1, ShouldNotBeNil)
		})

		err2 := e.putJobStatus(httpClient, cfg2)
		Convey("A run through without the auth token", func() {
			So(err2, ShouldNotBeNil)
		})
	})
}
func TestPutEvents(t *testing.T) {
	t.Parallel()

	Convey("internal method which puts the events into that instance", t, func() {
		json := []byte(`{"type":"` + "error" + `","time":"` + time.Now().String() + `","message":"` + "message" + `","messageOffset":"` + "msgOff" + `"}`)
		err := e.putEvent(httpClient, json, cfg, "")
		Convey("A complete run through with a postive response with it being added", func() {
			So(err, ShouldBeNil)
		})
		err1 := e.putEvent(httpClient, json, cfg2, "")
		Convey("Should through a status code error as it doesnt have authorisation", func() {
			So(err1, ShouldNotBeNil)
		})
		err2 := e.putEvent(httpClient, json, cfg1, "")
		Convey("Should throw an error when trying to request the the put job status within the putevent method", func() {
			So(err2, ShouldNotBeNil)
		})

		// err3 := e1.putEvent(httpClient, json, cfg, "")
		// Convey("Should throw an error when trying to request the the put job status within the putevent method", func() {
		// 	So(err3, ShouldNotBeNil)
		// })
	})
}
func TestHandleEvents(t *testing.T) {
	t.Parallel()
	Convey("Method which inits all the HandleEvent functionality", t, func() {
		cacheSize := 100 * 1024 * 1024
		c := freecache.NewCache(cacheSize)
		debug.SetGCPercent(20)
		err := e.HandleEvent(httpClient, c, cfg)
		Convey("Complete run through", func() {
			So(err, ShouldBeNil)
		})
		err1 := e1.HandleEvent(httpClient, c, cfg)
		Convey("Pass through an incorrect instance ID", func() {
			So(err1, ShouldNotBeNil)
		})
		err2 := e3.HandleEvent(httpClient, c, cfg)
		Convey("Should add the event to the events log ", func() {
			So(err2, ShouldBeNil)
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
