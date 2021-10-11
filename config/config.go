package config

import (
	"encoding/json"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

// KafkaSecProtocolTLS informs service to use TLS protocol for kafka
const KafkaSecProtocolTLS = "TLS"

// Config struct to hold application configuration.
type Config struct {
	BindAddress             string        `envconfig:"BIND_ADDR"`
	DatasetAPIURL           string        `envconfig:"DATASET_API_URL"`
	DatasetAPIAuthToken     string        `envconfig:"DATASET_API_AUTH_TOKEN"     json:"-"`
	CacheSize               int           `envconfig:"CACHE_SIZE"`
	CacheExpiry             int           `envconfig:"CACHE_EXPIRY"`
	GracefulShutdownTimeout time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	ServiceAuthToken        string        `envconfig:"SERVICE_AUTH_TOKEN"         json:"-"`
	ZebedeeURL              string        `envconfig:"ZEBEDEE_URL"`
	ReportEventTopic        string        `envconfig:"CONSUMER_TOPIC"`
	ReportEventGroup        string        `envconfig:"CONSUMER_GROUP"`
	KafkaBrokers            []string      `envconfig:"KAFKA_ADDR"`
	KafkaVersion            string        `envconfig:"KAFKA_VERSION"`
	KafkaSecProtocol        string        `envconfig:"KAFKA_SEC_PROTO"`
	KafkaSecCACerts         string        `envconfig:"KAFKA_SEC_CA_CERTS"`
	KafkaSecClientCert      string        `envconfig:"KAFKA_SEC_CLIENT_CERT"`
	KafkaSecClientKey       string        `envconfig:"KAFKA_SEC_CLIENT_KEY"       json:"-"`
	KafkaSecSkipVerify      bool          `envconfig:"KAFKA_SEC_SKIP_VERIFY"`
	KafkaOffsetOldest       bool          `envconfig:"KAFKA_OFFSET_OLDEST"`
}

var config *Config
var processConfig func(prefix string, spec interface{}) error = envconfig.Process

// Get configures the application and returns the configuration
func Get() (*Config, error) {
	if config != nil {
		return config, nil
	}

	config = &Config{
		BindAddress:             ":22200",
		DatasetAPIURL:           "http://localhost:22000",
		DatasetAPIAuthToken:     "FD0108EA-825D-411C-9B1D-41EF7727F465",
		CacheSize:               100 * 1024 * 1024,
		CacheExpiry:             60,
		GracefulShutdownTimeout: time.Second * 5,
		ServiceAuthToken:        "1D6C47C1-8F42-4F64-9AB4-6E5A16F89607",
		ZebedeeURL:              "http://localhost:8082",
		ReportEventTopic:        "report-events",
		ReportEventGroup:        "dp-import-reporter",
		KafkaBrokers:            []string{"localhost:9092"},
		KafkaVersion:            "1.0.2",
		KafkaOffsetOldest:       true,
	}

	if err := processConfig("", config); err != nil {
		return nil, errors.Wrap(err, "config: error while attempting to load environment config")
	}

	if config.KafkaSecProtocol != "" && config.KafkaSecProtocol != KafkaSecProtocolTLS {
		return nil, errors.New("KAFKA_SEC_PROTO has invalid value")
	}

	config.ServiceAuthToken = "Bearer " + config.ServiceAuthToken

	return config, nil
}

// String is implemented to prevent sensitive fields being logged.
// The config is returned as JSON with sensitive fields omitted.
func (config Config) String() string {
	json, _ := json.Marshal(config)
	return string(json)
}
