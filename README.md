dp-import-reporter
================

### Getting started

Service is authenticated against `zebedee`, one can run [dp-auth-api-stub](https://github.com/ONSdigital/dp-auth-api-stub) to mimic
service identity check in zebedee.

Run `make debug`

### Configuration

| Environment variable      | Default                              | Description
| ------------------------- | -------------------------------------| ------------------------------
| BIND_ADDR                 | :22200                               | The port to bind the application healhcheck endpoint to
| KAFKA_ADDR                | http://localhost:9092                | The address of the kafka Instance
| CONSUMER_GROUP            | dp-event-reporter                    | The kafka consumer group
| CONSUMER_TOPIC            | report-events                        | The kafka consumer topic
| DATASET_API_URL           | http://localhost:22000               | The URL of the import API
| DATASET_API_AUTH_TOKEN    | D0108EA-825D-411C-9B1D-41EF7727F465  | The Auth token for the Dataset API
| CACHE_SIZE                | 100 * 1024 * 1024                    | The size of the in memory cache
| CACHE_EXPIRY              | 60                                   | The time to live (in seconds) of the cache
| GRACEFUL_SHUTDOWN_TIMEOUT | 5s                                   | The shutdown timeout in seconds (time.Duration)
| SERVICE_AUTH_TOKEN        | AB0A5CFA-3C55-4FA8-AACC-F98039BED0AC | The service authorization token
| ZEBEDEE_URL               | http://localhost:8082                | The host name for Zebedee



### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2016-2017, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
