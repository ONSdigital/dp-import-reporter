package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/ONSdigital/go-ns/log"
)

// Config struct to hold application configuration.
type Config struct {
	ReportEventTopic        string        `envconfig:"CONSUMER_TOPIC"`
	Brokers                 []string      `envconfig:"KAFKA_ADDR"`
	DatasetAPIURL           string        `envconfig:"DATASET_API_URL"`
	DatasetAPIAuthToken     string        `envconfig:"DATASET_AUTH_TOKEN"`
	BindAddress             string        `envconfig:"BIND_ADDR"`
	CacheSize               int           `envconfig:"CACHE_SIZE"`
	CacheExpiry             int           `envconfig:"CACHE_EXPIRY"`
	GracefulShutdownTimeout time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
}

var config *Config
var processConfig func(prefix string, spec interface{}) error = envconfig.Process

func Get() (*Config, error) {
	if config != nil {
		return config, nil
	}

	config = &Config{
		ReportEventTopic:        "event-reporter",
		Brokers:                 []string{"localhost:9092"},
		DatasetAPIURL:           "http://localhost:21800",
		DatasetAPIAuthToken:     "FD0108EA-825D-411C-9B1D-41EF7727F465",
		BindAddress:             ":22200",
		CacheSize:               100 * 1024 * 1024,
		CacheExpiry:             60,
		GracefulShutdownTimeout: time.Second * 5,
	}

	if err := processConfig("", config); err != nil {
		log.ErrorC("error while attempting to load env config", err, nil)
		return nil, err
	}

	return config, nil
}
