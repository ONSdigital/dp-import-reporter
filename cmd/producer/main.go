package main

import (
	"flag"
	"time"

	"github.com/ONSdigital/dp-import-reporter/handler"
	"github.com/ONSdigital/dp-import-reporter/schema"
	"github.com/ONSdigital/go-ns/kafka"
)

// InputFileAvailableSchema schema
var InputFileAvailableSchema = `{
	"type": "record",
	"name": "input-file-available",
	"fields": [
		{"name": "file_url", "type": "string"},
		{"name": "instance_id", "type": "string"}
	]
}`

type inputFileAvailable struct {
	fileURL    string `avro:"file_url"`
	InstanceID string `avro:"instance_id"`
}

// from dp-observation-importer
var observationsInsertedSchema = `{
  "type": "record",
  "name": "import-observations-inserted",
  "fields": [
    {"name": "instance_id", "type": "string"},
    {"name": "observations_inserted", "type": "int"}
  ]
}`

// ObservationsInsertedEvent is the Avro schema for each observation batch inserted.
// var ObservationsInsertedEvent avro.Schema = avro.Schema{Definition: observationsInsertedSchema}

type insertedObservationsMessage struct {
	InstanceID           string `avro:"instance_id"`
	ObservationsInserted int32  `avro:"observations_inserted"`
}

/*
var TestNestedArraySchema = `{
	"type": "record",
	"name": "publish-dataset",
	"fields": [
		{
			"name": "instance_ids",
			"type": {
				"type": "array",
				"items": "string"
			}
		},
		{
			"name": "files",
			"type": {
				"type": "array",
				"items": {
					"name": "file",
					"type": "record",
					"fields": [
						{
							"name": "alias-name",
							"type": "string"
						},
						{
							"name": "url",
							"type": "string"
						}
					]
				}
			}
		}
	]
}`

type testNestedArray struct {
	InstanceIDs []string `avro:"instance_ids"`
	Files       []File   `avro:"files"`
}

type File struct {
	AliasName string `avro:"alias-name"`
	URL       string `avro:"url"`
}
*/

// type EventReport struct {
// 	InstanceID string `avro:"instance_id"`
// 	EventType  string `avro:"event_type"`
// 	EventMsg   string `avro:"event_message"`
// 	// Number     int    `avro:"number"`
// }

func main() {
	// instance_id := flag.String("id", "21", "instance_id")
	producerTopic := flag.String("topic", "event-reporter", "producer topic")
	// fileURL := flag.String("s3", "s3://dp-dimension-extractor/OCIGrowth.csv", "s3 file")
	// insertedObservations := flag.Int("inserts", 2000, "inserted observations")
	flag.Parse()

	brokers := []string{"localhost:9092"}
	// inputFileAvailableProducer, err := kafka.NewProducer(brokers, "dimensions-extracted", int(2000000))
	// inputFileAvailableProducer, err := kafka.NewProducer(brokers, "input-test-file", int(2000000))
	producer, err := kafka.NewProducer(brokers, *producerTopic, int(2000000))
	if err != nil {
		panic(err)
	}

	avroBytes, _ := schema.ReportedEventSchema.Marshal(&handler.EventReport{
		InstanceID: "a4695fee-f0a2-49c4-b136-e3ca8dd40476",
		EventType:  "error",
		EventMsg:   "Broken on something.",
		// Number:     123,
	})

	producer.Output() <- avroBytes
	//  producer.Output() <- []byte("{}")
	time.Sleep(time.Duration(1000 * time.Millisecond))
	// producer.Output() <- avroBytes
	// time.Sleep(time.Duration(1000 * time.Millisecond))
	//
	// producer.Output() <- avroBytes
	// time.Sleep(time.Duration(1000 * time.Millisecond))

	producer.Closer() <- true
}
