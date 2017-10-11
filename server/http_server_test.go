package server

import (
	"github.com/ONSdigital/dp-import-reporter/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDropCache(t *testing.T) {
	Convey("Given a valid request & response", t, func() {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, dropCachePath, nil)

		cacheMock := &mocks.ClearableCacheMock{
			ClearFunc: func() {
				// Do nothing
			},
		}
		clearCacheHandler := ClearCache(cacheMock)

		Convey("When DropCache is invoked", func() {
			clearCacheHandler.ServeHTTP(w, r)

			Convey("Then cache.Clear is called once", func() {
				So(len(cacheMock.ClearCalls()), ShouldEqual, 1)
			})

			Convey("And the response is as expected", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})
	})
}

func TestHealthCheck(t *testing.T) {
	Convey("Given a valid request & response", t, func() {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, healthCheckPath, nil)

		Convey("When HealthCheck is invoked", func() {
			HealthCheck(w, r)

			Convey("And the response is as expected", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})
	})
}