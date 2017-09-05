package config

import "github.com/ian-kent/gofigure"

type Config struct {
	NewInstanceTopic string   `env:"CONSUMER_TOPIC" flag:"event-reporter" flagDesc:"topic name for import file available events"`
	Brokers          []string `env:"KAFKA_ADDR" flag:"kafka-addr" flagDesc:"topic name for import file available events"`
	ImportAPIURL     string   `env:"IMPORT_API_URL" flag:"import-addr" flagDesc:"The address of Import API"`
	ImportAuthToken  string   `env:"IMPORT_AUTH_TOKEN" flag:"import-auth-token" flagDesc:"Authentication token for access to import API"`
	BindAddress      string   `env:"BIND_ADDR" flag:"bind-addr" flagDesc:"The bind port"`
	CacheSize        int      `env:"CACHE_SIZE" flag:"CACHE_SIZE" flagDesc:"The bind port"`
}

var cfg *Config

func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}
	cfg = &Config{
		NewInstanceTopic: "event-reporter",
		Brokers:          []string{"localhost:9092"},
		ImportAPIURL:     "http://localhost:21800",
		ImportAuthToken:  "FD0108EA-825D-411C-9B1D-41EF7727F465",
		BindAddress:      ":22200",
		CacheSize:        100 * 1024 * 1024,
	}
	if err := gofigure.Gofigure(cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
