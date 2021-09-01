package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDropCache(t *testing.T) {
	Convey("Given a valid request & response", t, func() {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/dropcache", nil)

		cacheMock := &ClearableCacheMock{
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
		r := httptest.NewRequest(http.MethodGet, "/healthcheck", nil)

		Convey("When HealthCheck is invoked", func() {
			HealthCheck(w, r)

			Convey("And the response is as expected", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})
	})
}
