package config

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
	"errors"
	"time"
)

type processConfigMock struct {
	prefixCalls []string
	specCalls   []interface{}
	err         error
}

func (m *processConfigMock) Process(prefix string, spec interface{}) error {
	m.prefixCalls = append(m.prefixCalls, prefix)
	m.specCalls = append(m.specCalls, spec)
	return m.err
}

func newProcessCgfMock() *processConfigMock {
	return &processConfigMock{
		prefixCalls: make([]string, 0),
		specCalls:   make([]interface{}, 0),
		err:         nil,
	}
}

func TestConfig_configNotNil(t *testing.T) {
	Convey("Given that config is not nil", t, func() {
		cfg = &Config{}

		mockProcessCgf := newProcessCgfMock()
		processConfig = mockProcessCgf.Process

		Convey("When Get is called", func() {

			actual, err := Get()

			Convey("Then no error is returned along with the expected config", func() {
				So(actual, ShouldResemble, cfg)
				So(err, ShouldBeNil)
			})

			Convey("And processConfig is never called", func() {
				So(len(mockProcessCgf.prefixCalls), ShouldEqual, 0)
				So(len(mockProcessCgf.specCalls), ShouldEqual, 0)
			})
		})

	})
}

func TestConfig_configNilErr(t *testing.T) {
	Convey("Given that config is nil", t, func() {
		cfg = nil
		mockProcessCgf := newProcessCgfMock()
		processConfig = mockProcessCgf.Process

		Convey("When processConfig returns an error", func() {
			mockProcessCgf := newProcessCgfMock()
			mockProcessCgf.err = errors.New("Boom!")
			processConfig = mockProcessCgf.Process

			actual, err := Get()

			Convey("Then the nil and the expected error are returned", func() {
				So(actual, ShouldBeNil)
				So(err, ShouldResemble, errors.New("Boom!"))
			})

			Convey("And processConfig is called 1 time with expected parameters", func() {
				So(len(mockProcessCgf.prefixCalls), ShouldEqual, 1)
				So(mockProcessCgf.prefixCalls[0], ShouldEqual, "")

				expected := &Config{
					NewInstanceTopic:        "event-reporter",
					Brokers:                 []string{"localhost:9092"},
					DatasetAPIURL:           "http://localhost:22000",
					ImportAuthToken:         "FD0108EA-825D-411C-9B1D-41EF7727F465",
					BindAddress:             ":22200",
					CacheSize:               100 * 1024 * 1024,
					GracefulShutdownTimeout: time.Second * 5,
				}

				So(len(mockProcessCgf.specCalls), ShouldEqual, 1)
				So(mockProcessCgf.specCalls[0], ShouldResemble, expected)
			})
		})
	})
}

func TestConfig_configNilSuccess(t *testing.T) {
	Convey("Given that config is nil", t, func() {
		cfg = nil
		mockProcessCgf := newProcessCgfMock()
		processConfig = mockProcessCgf.Process

		Convey("When processConfig is called", func() {
			mockProcessCgf := newProcessCgfMock()
			mockProcessCgf.err = nil
			processConfig = mockProcessCgf.Process

			expected := &Config{
				NewInstanceTopic:        "event-reporter",
				Brokers:                 []string{"localhost:9092"},
				DatasetAPIURL:           "http://localhost:22000",
				ImportAuthToken:         "FD0108EA-825D-411C-9B1D-41EF7727F465",
				BindAddress:             ":22200",
				CacheSize:               100 * 1024 * 1024,
				GracefulShutdownTimeout: time.Second * 5,
			}

			actual, err := Get()

			Convey("Then the expected config is returned and error is nil", func() {
				So(actual, ShouldResemble, expected)
				So(err, ShouldBeNil)
			})

			Convey("And processConfig is called 1 time with expected parameters", func() {
				So(len(mockProcessCgf.prefixCalls), ShouldEqual, 1)
				So(mockProcessCgf.prefixCalls[0], ShouldEqual, "")

				So(len(mockProcessCgf.specCalls), ShouldEqual, 1)
				So(mockProcessCgf.specCalls[0], ShouldResemble, expected)
			})
		})
	})
}
