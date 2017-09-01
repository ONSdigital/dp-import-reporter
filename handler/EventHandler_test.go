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
	ImportAuthToken:  "D0108EA-825D-411C-9B1D-41EF7727F465",
	BindAddress:      ":22200",
}

//wrong ImportAPI
var cfg1 = &config.Config{
	NewInstanceTopic: "event-reporter",
	Brokers:          []string{"localhost:9092"},
	ImportAPIURL:     "http://localho:21800",
	ImportAuthToken:  "D0108EA-825D-411C-9B1D-41EF7727F465",
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

// func TestPutJobStatus(t *testing.T) {
//
// 	Convey("Internal method which changes the job status ", t, func() {
// 		cfg3, _ := config.Get()
//
// 		err := e.putJobStatus(httpClient, cfg3)
// 		Convey("A complete working run through of all the code in a positive manner", func() {
// 			So(err, ShouldBeNil)
// 		})
//
// 		err1 := e.putJobStatus(httpClient, cfg1)
// 		Convey("A run through with an incomplete url", func() {
// 			So(err1, ShouldNotBeNil)
// 		})
//
// 		err2 := e.putJobStatus(httpClient, cfg2)
// 		Convey("A run through without the auth token", func() {
// 			So(err2, ShouldNotBeNil)
// 		})
// 	})
// }
func TestPutEvents(t *testing.T) {
	t.Parallel()

	Convey("internal method which puts the events into that instance", t, func() {
		json := []byte(`{"type":"` + "error" + `","time":"` + time.Now().String() + `","message":"` + "message" + `","messageOffset":"` + "msgOff" + `"}`)
		cfg4, _ := config.Get()
		err := e.putEvent(httpClient, json, cfg4, "")
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

func TestArraySlicing(t *testing.T) {
	Convey("A method which slices a events array up", t, func() {
		_, events, err := e.checkInstance(httpClient, cfg)
		Convey("It brings back a valid instance", func() {
			So(err, ShouldBeNil)
		})
		var aE = event{
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

// func TestCache(t *testing.T) {
// 	t.Parallel()
// 	Convey("Checks for eventhandler caching", t, func() {
// 		cacheSize := 100 * 1024 * 1024
// 		c := freecache.NewCache(cacheSize)
// 		debug.SetGCPercent(20)
//
// 		key, err := c.Get([]byte("a4695fee-f0a2-49c4-b136-e3ca8dd40476"))
// 		Convey("There should be no error as its should be added", func() {
// 			So(err, ShouldNotBeNil)
// 			So(string(key), ShouldBeBlank)
// 		})
// 		c.Set([]byte("a4695fee-f0a2-49c4-b136-e3ca8dd40476"), []byte("value"), 30)
// 		test, err := c.Get([]byte("a4695fee-f0a2-49c4-b136-e3ca8dd40476"))
// 		Convey("To see if the instance has been added to the cache", func() {
// 			So(err, ShouldBeNil)
// 			So(string(test), ShouldNotBeBlank)
// 		})
// 	})
// }
//
// func TestInstanceCheckSuccesful(t *testing.T) {
//
// 	t.Parallel()
// 	Convey("Checks if a instance exists", t, func() {
// 		instanceID := "a4695fee-f0a2-49c4-b136-e3ca8dd40476"
// 		r, err := http.Get("http://localhost:21800/instances/" + instanceID)
// 		So(err, ShouldBeNil)
// 		So(r.StatusCode, ShouldEqual, http.StatusOK)
// 	})
// }
// func TestInstanceCheckUnsuccesful(t *testing.T) {
//
// 	t.Parallel()
// 	Convey("Checks if a instance exists but will be unsuccesful", t, func() {
// 		instanceID := "a4695fee-f0a2-49c4-b136-e3ca8dd4047"
// 		r, err := http.Get("http://localhost:21800/instances/" + instanceID)
// 		So(err, ShouldBeNil)
// 		So(r.StatusCode, ShouldEqual, 404)
// 	})
// }
//
// func TestPutEventSuccessful(t *testing.T) {
// 	t.Parallel()
// 	Convey("An event is added to the instance", t, func() {
// 		jsonUpload := []byte(`{"type":"` + "error" + `","time":"` + "21/6/2015" + `","message":"` + "message" + `","messageOffset":"` + "msgOff" + `"}`)
//
// 		r, err := http.NewRequest("PUT", "http://localhost:21800/instances/a4695fee-f0a2-49c4-b136-e3ca8dd40476/events", bytes.NewBuffer(jsonUpload))
// 		So(err, ShouldBeNil)
// 		r.Header.Set("Internal-token", "FD0108EA-825D-411C-9B1D-41EF7727F465")
// 		httpClient := &http.Client{}
// 		res, err := httpClient.Do(r)
// 		So(err, ShouldBeNil)
// 		So(res.StatusCode, ShouldEqual, 201)
// 	})
// }
// func TestPutEventUnsuccessful(t *testing.T) {
// 	t.Parallel()
// 	Convey("No token added to the response and then sent to the events endpoint", t, func() {
// 		jsonUpload := []byte(`{"type":"` + "error" + `","time":"` + "21/6/2015" + `","message":"` + "message" + `","messageOffset":"` + "msgOff" + `"}`)
//
// 		r, err := http.NewRequest("PUT", "http://localhost:21800/instances/a4695fee-f0a2-49c4-b136-e3ca8dd40476/events", bytes.NewBuffer(jsonUpload))
// 		So(err, ShouldBeNil)
// 		httpClient := &http.Client{}
// 		res, err := httpClient.Do(r)
// 		So(err, ShouldBeNil)
// 		So(res.StatusCode, ShouldEqual, 403)
// 	})
//
// }
//
// func TestPutEventUnsuccessfulJSON(t *testing.T) {
// 	t.Parallel()
// 	Convey("An invalid JSON response is sent to the events endpoint", t, func() {
// 		jsonUpload := []byte(`{"type":"` + "error" + `","time":"` + "21/6/2015" + `","messageOffset":"` + "msgOff" + `"}`)
// 		log.Debug("HHello world", nil)
// 		r, err := http.NewRequest("PUT", "http://localhost:21800/instances/a4695fee-f0a2-49c4-b136-e3ca8dd40476/events", bytes.NewBuffer(jsonUpload))
// 		So(err, ShouldBeNil)
// 		r.Header.Set("Internal-token", "FD0108EA-825D-411C-9B1D-41EF7727F465")
// 		httpClient := &http.Client{}
// 		res, err := httpClient.Do(r)
// 		So(err, ShouldBeNil)
// 		So(res.StatusCode, ShouldEqual, 400)
// 	})
// }
//
// func TestPutJobStatusSuccessful(t *testing.T) {
// 	t.Parallel()
// 	Convey("A successful change of job status", t, func() {
// 		jsonUpload := []byte(`{"state":"` + "failed" + `"}`)
// 		r, err := http.NewRequest("PUT", "http://localhost:21800/instances/a4695fee-f0a2-49c4-b136-e3ca8dd40476", bytes.NewBuffer(jsonUpload))
// 		So(err, ShouldBeNil)
// 		r.Header.Set("Internal-token", "FD0108EA-825D-411C-9B1D-41EF7727F465")
// 		httpClient := &http.Client{}
// 		res, err := httpClient.Do(r)
// 		So(err, ShouldBeNil)
// 		So(res.StatusCode, ShouldEqual, http.StatusOK)
// 	})
// }
//
// func TestPutJobStatusUnsuccessful(t *testing.T) {
// 	t.Parallel()
// 	Convey("A unsuccessful change of job status", t, func() {
// 		jsonUpload := []byte(`{"state":"` + "failed" + `"}`)
// 		r, err := http.NewRequest("PUT", "http://localhost:21800/instances/a4695fee-f0a2-49c4-b136-e3ca8dd40476", bytes.NewBuffer(jsonUpload))
// 		So(err, ShouldBeNil)
//
// 		httpClient := &http.Client{}
// 		res, err := httpClient.Do(r)
// 		So(err, ShouldBeNil)
// 		So(res.StatusCode, ShouldNotEqual, http.StatusOK)
// 	})
// }
