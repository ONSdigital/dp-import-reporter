package config

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfig(t *testing.T) {
	cfg, err := Get()
	Convey("Given that the enviroment has no enviroment variables", t, func() {

		Convey("When the values are retrieved", func() {

			Convey("There should be no errors returned", func() {
				So(err, ShouldBeNil)
			})
			Convey("Config variables should match those of the enviroment", func() {
				So(cfg.NewInstanceTopic, ShouldEqual, "event-reporter")
				So(cfg.Brokers, ShouldResemble, []string{"localhost:9092"})
				So(cfg.DatasetAPIURL, ShouldEqual, "http://localhost:22000")
				So(cfg.ImportAuthToken, ShouldEqual, "FD0108EA-825D-411C-9B1D-41EF7727F465")
				So(cfg.BindAddress, ShouldEqual, ":22200")
			})
		})

	})
}
