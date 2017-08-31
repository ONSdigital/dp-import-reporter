package main

import (
	"errors"
	"net/http"
	"runtime/debug"

	"github.com/ONSdigital/dp-import-reporter/handler"
	"github.com/ONSdigital/dp-import-reporter/schema"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
	"github.com/coocood/freecache"
	"github.com/ian-kent/gofigure"
)

type config struct {
	NewInstanceTopic string   `env:"INPUT_FILE_AVAILABLE_TOPIC" flag:"event-reporter" flagDesc:"topic name for import file available events"`
	Brokers          []string `env:"KAFKA_ADDR" flag:"kafka-addr" flagDesc:"topic name for import file available events"`
	ImportAddr       string   `env:"IMPORT_ADDR" flag:"import-addr" flagDesc:"The address of Import API"`
	ImportAuthToken  string   `env:"IMPORT_AUTH_TOKEN" flag:"import-auth-token" flagDesc:"Authentication token for access to import API"`
}

// logFatal is a utility method for a common failure pattern in main()
func logFatal(context string, err error, data log.Data) {
	log.ErrorC(context, err, data)
	panic(err)
}

func main() {
	newInstanceEventConsumer, cfg, err := consumerInit()
	if err != nil {
		logFatal("error initiating", err, nil)
	}
	//cache init
	c := cacheSetup()
	client := &http.Client{}

	loop(newInstanceEventConsumer, &cfg, client, c)
}

func consumerInit() (*kafka.ConsumerGroup, config, error) {
	log.Namespace = "dp-event-reporter"

	cfg := config{
		NewInstanceTopic: "event-reporter",
		Brokers:          []string{"localhost:9092"},
		ImportAddr:       "http://localhost:21800",
		ImportAuthToken:  "FD0108EA-825D-411C-9B1D-41EF7727F465",
	}

	if err := gofigure.Gofigure(&cfg); err != nil {
		logFatal("gofigure failed", err, nil)
	}

	log.Info("starting", log.Data{
		"new-import-topic": cfg.NewInstanceTopic,
		"import-api":       cfg.ImportAddr,
	})
	newInstanceEventConsumer, err := kafka.NewConsumerGroup(cfg.Brokers, cfg.NewInstanceTopic, log.Namespace, kafka.OffsetNewest)
	if err != nil {
		logFatal("could not obtain consumer", err, nil)
	}
	return newInstanceEventConsumer, cfg, err
}
func cacheSetup() *freecache.Cache {
	cacheSize := 100 * 1024 * 1024
	c := freecache.NewCache(cacheSize)
	debug.SetGCPercent(20)
	return c
}

func loop(newInstanceEventConsumer *kafka.ConsumerGroup, cfg *config, client *http.Client, c *freecache.Cache) {

SUCCESS:
	for {
		select {
		case newInstanceMessage := <-newInstanceEventConsumer.Incoming():
			var msg handler.EventReport
			if err := schema.ReportedEventSchema.Unmarshal(newInstanceMessage.GetData(), &msg); err != nil {
				log.ErrorC("failed to unmarshal message", err, log.Data{"topic": cfg.NewInstanceTopic})
				// Fatal error reading message, should never fall in here
				continue
			}

			if err := msg.HandleEvent(client, cfg.ImportAuthToken, c); err != nil {
				log.ErrorC("Failure updating events", err, log.Data{"topic": cfg.NewInstanceTopic})
			}

			newInstanceMessage.Commit()

		case newImportConsumerErrorMessage := <-newInstanceEventConsumer.Errors():
			log.Error(errors.New("Event handler has stopped working."), log.Data{"error": newImportConsumerErrorMessage, "topic": cfg.NewInstanceTopic})
			break SUCCESS
		}
	}

	// assert: only get here when we have an error, which has been logged
	newInstanceEventConsumer.Closer() <- true
	logFatal("", errors.New("aborting after error"), nil)
}
