package config

import (
	"time"

	"github.com/ian-kent/gofigure"
)

type Config struct {
	NewInstanceTopic        string        `env:"CONSUMER_TOPIC" flag:"event-reporter" flagDesc:"topic name for import file available events"`
	Brokers                 []string      `env:"KAFKA_ADDR" flag:"kafka-addr" flagDesc:"topic name for import file available events"`
	DatasetAPIURL           string        `env:"IMPORT_API_URL" flag:"import-addr" flagDesc:"The address of Import API"`
	ImportAuthToken         string        `env:"IMPORT_AUTH_TOKEN" flag:"import-auth-token" flagDesc:"Authentication token for access to import API"`
	BindAddress             string        `env:"BIND_ADDR" flag:"bind-addr" flagDesc:"The bind port"`
	CacheSize               int           `env:"CACHE_SIZE" flag:"CACHE_SIZE" flagDesc:"The bind port"`
	GracefulShutdownTimeout time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
}

var cfg *Config

func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}
	cfg = &Config{
		NewInstanceTopic:        "event-reporter",
		Brokers:                 []string{"localhost:9092"},
		DatasetAPIURL:           "http://localhost:22000",
		ImportAuthToken:         "FD0108EA-825D-411C-9B1D-41EF7727F465",
		BindAddress:             ":22200",
		CacheSize:               100 * 1024 * 1024,
		GracefulShutdownTimeout: time.Second * 10,
	}
	if err := gofigure.Gofigure(cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
