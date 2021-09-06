# dp-import-reporter

## Getting started

Service is authenticated against `zebedee`, one can run [dp-auth-api-stub](https://github.com/ONSdigital/dp-auth-api-stub) to mimic
service identity check in zebedee.

Run `make debug`

## Kafka scripts

Scripts for updating and debugging Kafka can be found [here](https://github.com/ONSdigital/dp-data-tools)(dp-data-tools)

### Configuration

| Environment variable      | Default                              | Description
| ------------------------- | -------------------------------------| ------------------------------
| BIND_ADDR                 | :22200                               | The port to bind the application healhcheck endpoint to
| KAFKA_LEGACY_ADDR         | `localhost:9092`                     | The addresses of the kafka brokers (CSV) - non-TLS
| KAFKA_LEGACY_VERSION      | `1.0.2`                              | The version of Kafka - non-TLS
| KAFKA_ADDR                | `localhost:9092`                     | The addresses of the kafka brokers (CSV)
| KAFKA_VERSION             | `1.0.2`                              | The version of Kafka
| KAFKA_SEC_PROTO           | _unset_                              | if set to `TLS`, kafka connections will use TLS [1]
| KAFKA_SEC_CLIENT_KEY      | _unset_                              | PEM for the client key [1]
| KAFKA_SEC_CLIENT_CERT     | _unset_                              | PEM for the client certificate [1]
| KAFKA_SEC_CA_CERTS        | _unset_                              | CA cert chain for the server cert [1]
| KAFKA_SEC_SKIP_VERIFY     | false                                | ignores server certificate issues if `true` [1]
| KAFKA_OFFSET_OLDEST       | true                                 | start consuming kafka topics at oldest message (if false, starts at newest)
| CONSUMER_GROUP            | dp-event-reporter                    | The kafka consumer group
| CONSUMER_TOPIC            | report-events                        | The kafka consumer topic
| DATASET_API_URL           | http://localhost:22000               | The URL of the dataset API
| DATASET_API_AUTH_TOKEN    | D0108EA-825D-411C-9B1D-41EF7727F465  | The Auth token for the Dataset API
| CACHE_SIZE                | 100 * 1024 * 1024                    | The size of the in memory cache
| CACHE_EXPIRY              | 60                                   | The time to live (in seconds) of the cache
| GRACEFUL_SHUTDOWN_TIMEOUT | 5s                                   | The shutdown timeout in seconds (time.Duration)
| SERVICE_AUTH_TOKEN        | AB0A5CFA-3C55-4FA8-AACC-F98039BED0AC | The service authorization token
| ZEBEDEE_URL               | http://localhost:8082                | The host name for Zebedee

**Notes:**

1. For more info, see the [kafka TLS examples documentation](https://github.com/ONSdigital/dp-kafka/tree/main/examples#tls)

## Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

## License

Copyright © 2016-2021, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
