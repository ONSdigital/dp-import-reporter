package model

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
	"errors"
	"encoding/json"
)

func TestReportEvent_GenCacheKeyAndValue(t *testing.T) {
	Convey("Given a nil reportEvent", t, func() {
		var e *ReportEvent

		Convey("When GenCacheKeyAndValue is invoked", func() {
			key, val, err := e.GenCacheKeyAndValue()

			Convey("Then the expected error, key and val are returned", func() {
				So(err, ShouldResemble, errors.New(reportEventNil))
				So(key, ShouldBeNil)
				So(val, ShouldBeNil)
			})
		})
	})

	Convey("Given a reportEvent with an empty instanceID", t, func() {
		reportEvent := &ReportEvent{}

		Convey("When GenCacheKeyAndValue is invoked", func() {
			key, val, err := reportEvent.GenCacheKeyAndValue()

			Convey("Then the expected error, key and val are returned", func() {
				So(err, ShouldResemble, errors.New(reportEventInstanceIDEmpty))
				So(key, ShouldBeNil)
				So(val, ShouldBeNil)
			})
		})
	})

	Convey("Given a reportEvent with an empty eventType", t, func() {
		reportEvent := &ReportEvent{InstanceID: "666"}

		Convey("When GenCacheKeyAndValue is invoked", func() {
			key, val, err := reportEvent.GenCacheKeyAndValue()

			Convey("Then the expected error, key and val are returned", func() {
				So(err, ShouldResemble, errors.New(reportEventTypeEmpty))
				So(key, ShouldBeNil)
				So(val, ShouldBeNil)
			})
		})
	})

	Convey("Given a reportEvent with an empty serviceName", t, func() {
		reportEvent := &ReportEvent{
			InstanceID: "666",
			EventType:  "error",
		}

		Convey("When GenCacheKeyAndValue is invoked", func() {
			key, val, err := reportEvent.GenCacheKeyAndValue()

			Convey("Then the expected error, key and val are returned", func() {
				So(err, ShouldResemble, errors.New(reportEventServiceNameEmpty))
				So(key, ShouldBeNil)
				So(val, ShouldBeNil)
			})
		})
	})

	Convey("Given a valid reportEvent", t, func() {
		reportEvent := &ReportEvent{
			InstanceID:  "666",
			EventType:   "error",
			ServiceName: "myService",
			EventMsg:    "its all gone wrong",
		}

		Convey("When GenCacheKeyAndValue is invoked", func() {
			key, val, err := reportEvent.GenCacheKeyAndValue()

			Convey("Then the expected error, key and val are returned", func() {
				expectedKey, _ := json.Marshal(&cacheKey{
					instanceID:  reportEvent.InstanceID,
					eventType:   reportEvent.EventType,
					serviceName: reportEvent.ServiceName,
				})

				expectedVal, _ := json.Marshal(reportEvent)

				So(err, ShouldBeNil)
				So(key, ShouldResemble, expectedKey)
				So(val, ShouldResemble, expectedVal)
			})
		})
	})
}
