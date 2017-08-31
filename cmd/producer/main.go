package main

import (
	"flag"
	"time"

	"github.com/ONSdigital/dp-import-reporter/handler"
	"github.com/ONSdigital/dp-import-reporter/schema"
	"github.com/ONSdigital/go-ns/kafka"
)

// this is a test file to reproduce events being pushed to the kafka consumer

type inputFileAvailable struct {
	fileURL    string `avro:"file_url"`
	InstanceID string `avro:"instance_id"`
}

func main() {
	producerTopic := flag.String("topic", "event-reporter", "producer topic")

	flag.Parse()

	brokers := []string{"localhost:9092"}

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
	time.Sleep(time.Duration(1000 * time.Millisecond))

	producer.Closer() <- true
}
