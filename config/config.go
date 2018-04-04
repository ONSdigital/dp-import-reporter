package config

import (
	"encoding/json"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

// Config struct to hold application configuration.
type Config struct {
	ReportEventTopic        string        `envconfig:"CONSUMER_TOPIC"`
	ReportEventGroup        string        `envconfig:"CONSUMER_GROUP"`
	Brokers                 []string      `envconfig:"KAFKA_ADDR"`
	DatasetAPIURL           string        `envconfig:"DATASET_API_URL"`
	DatasetAPIAuthToken     string        `envconfig:"DATASET_API_AUTH_TOKEN"     json:"-"`
	BindAddress             string        `envconfig:"BIND_ADDR"`
	CacheSize               int           `envconfig:"CACHE_SIZE"`
	CacheExpiry             int           `envconfig:"CACHE_EXPIRY"`
	GracefulShutdownTimeout time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	ServiceAuthToken        string        `envconfig:"SERVICE_AUTH_TOKEN"         json:"-"`
	ZebedeeURL              string        `envconfig:"ZEBEDEE_URL"`
}

var config *Config
var processConfig func(prefix string, spec interface{}) error = envconfig.Process

// Get configures the application and returns the configuration
func Get() (*Config, error) {
	if config != nil {
		return config, nil
	}

	config = &Config{
		ReportEventTopic:        "report-events",
		ReportEventGroup:        "dp-import-reporter",
		Brokers:                 []string{"localhost:9092"},
		DatasetAPIURL:           "http://localhost:22000",
		DatasetAPIAuthToken:     "FD0108EA-825D-411C-9B1D-41EF7727F465",
		BindAddress:             ":22200",
		CacheSize:               100 * 1024 * 1024,
		CacheExpiry:             60,
		GracefulShutdownTimeout: time.Second * 5,
		ServiceAuthToken:        "1D6C47C1-8F42-4F64-9AB4-6E5A16F89607",
		ZebedeeURL:              "http://localhost:8082",
	}

	config.ServiceAuthToken = "Bearer " + config.ServiceAuthToken

	if err := processConfig("", config); err != nil {
		return nil, errors.Wrap(err, "config: error while attempting to load environment config")
	}

	return config, nil
}

// String is implemented to prevent sensitive fields being logged.
// The config is returned as JSON with sensitive fields omitted.
func (config Config) String() string {
	json, _ := json.Marshal(config)
	return string(json)
}
