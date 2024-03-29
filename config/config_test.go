package config

import (
	"os"
	"testing"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	expectedConfig = &Config{
		ReportEventGroup:        "dp-import-reporter",
		ReportEventTopic:        "report-events",
		KafkaBrokers:            []string{"localhost:9092", "localhost:9093", "localhost:9094"},
		KafkaVersion:            "1.0.2",
		KafkaSecProtocol:        "",
		KafkaOffsetOldest:       true,
		DatasetAPIURL:           "http://localhost:22000",
		DatasetAPIAuthToken:     "FD0108EA-825D-411C-9B1D-41EF7727F465",
		BindAddress:             ":22200",
		CacheSize:               100 * 1024 * 1024,
		CacheExpiry:             60,
		GracefulShutdownTimeout: time.Second * 5,
		ServiceAuthToken:        "1D6C47C1-8F42-4F64-9AB4-6E5A16F89607",
		ZebedeeURL:              "http://localhost:8082",
	}

	errMock = errors.New("boom")
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

func newProcessConfigMock() *processConfigMock {
	return &processConfigMock{
		prefixCalls: make([]string, 0),
		specCalls:   make([]interface{}, 0),
		err:         nil,
	}
}

func TestConfigNotNil(t *testing.T) {
	Convey("Given that config is not nil", t, func() {
		config = &Config{}

		mockProcessCgf := newProcessConfigMock()
		processConfig = mockProcessCgf.Process

		Convey("When Get is called", func() {

			actual, err := Get()

			Convey("Then no error is returned along with the expected config", func() {
				So(actual, ShouldResemble, config)
				So(err, ShouldBeNil)
			})

			Convey("And processConfig is never called", func() {
				So(len(mockProcessCgf.prefixCalls), ShouldEqual, 0)
				So(len(mockProcessCgf.specCalls), ShouldEqual, 0)
			})
		})

	})
}

func TestConfigNilErr(t *testing.T) {
	Convey("Given that config is nil", t, func() {
		config = nil
		mockProcessCgf := newProcessConfigMock()
		processConfig = mockProcessCgf.Process

		Convey("When processConfig returns an error", func() {
			mockProcessCgf := newProcessConfigMock()
			mockProcessCgf.err = errMock
			processConfig = mockProcessCgf.Process

			actual, err := Get()

			Convey("Then the nil and the expected error are returned", func() {
				So(actual, ShouldBeNil)
				So(err.Error(), ShouldEqual, errors.Wrap(errMock, "config: error while attempting to load environment config").Error())
			})

			Convey("And processConfig is called 1 time with expected parameters", func() {
				So(len(mockProcessCgf.prefixCalls), ShouldEqual, 1)
				So(mockProcessCgf.prefixCalls[0], ShouldEqual, "")
				So(len(mockProcessCgf.specCalls), ShouldEqual, 1)
				So(mockProcessCgf.specCalls[0], ShouldResemble, expectedConfig)
			})
		})
	})
}

func TestConfigNilSuccess(t *testing.T) {
	Convey("Given that config is nil", t, func() {
		config = nil
		mockProcessCgf := newProcessConfigMock()
		processConfig = mockProcessCgf.Process

		Convey("When processConfig is called", func() {
			mockProcessCgf := newProcessConfigMock()
			mockProcessCgf.err = nil
			processConfig = mockProcessCgf.Process

			actual, err := Get()

			Convey("Then the expected config is returned and error is nil", func() {
				expectedConfig.ServiceAuthToken = "Bearer " + expectedConfig.ServiceAuthToken
				So(actual, ShouldResemble, expectedConfig)
				So(err, ShouldBeNil)
			})

			Convey("And processConfig is called 1 time with expected parameters", func() {
				So(len(mockProcessCgf.prefixCalls), ShouldEqual, 1)
				So(mockProcessCgf.prefixCalls[0], ShouldEqual, "")

				So(len(mockProcessCgf.specCalls), ShouldEqual, 1)
				So(mockProcessCgf.specCalls[0], ShouldResemble, expectedConfig)
			})
		})
	})
}

func TestGetReturnsCorrectValues(t *testing.T) {
	Convey("Given a clean environment", t, func() {
		os.Clearenv()
		config = nil
		processConfig = envconfig.Process

		Convey("When default configuration is obtained", func() {
			configuration, err := Get()

			Convey("Then expected configuration is returned", func() {
				So(err, ShouldBeNil)
				So(configuration, ShouldResemble, expectedConfig)
			})
		})

		Convey("When configuration is called with an invalid security setting", func() {
			os.Setenv("KAFKA_SEC_PROTO", "ssl")
			configuration, err := Get()

			Convey("Then an error is returned", func() {
				So(configuration, ShouldBeNil)
				So(err.Error(), ShouldEqual, "KAFKA_SEC_PROTO has invalid value")
			})
		})

	})
}
