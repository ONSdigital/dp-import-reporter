package main

import (
	"context"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/ONSdigital/dp-import-reporter/config"
	"github.com/ONSdigital/dp-import-reporter/handler"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
	"github.com/coocood/freecache"
	"github.com/ONSdigital/dp-import-reporter/healthcheck"
	"github.com/ONSdigital/dp-import-reporter/event"
	"github.com/ONSdigital/dp-import-reporter/client"
	"github.com/ONSdigital/dp-import-reporter/model"
	"net/http"
	"io/ioutil"
	"io"
	"github.com/ONSdigital/dp-import-reporter/schema"
)

// logFatal is a utility method for a common failure pattern in main()
func logFatal(context string, err error, data log.Data) {
	log.ErrorC(context, err, data)
	os.Exit(1)
}

type ResponseBodyReader struct{}

func (r ResponseBodyReader) Read(reader io.Reader) ([]byte, error) {
	return ioutil.ReadAll(reader)
}

func main() {
	log.Namespace = "dp-event-reporter"

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	errorChannel := make(chan error, 1)

	cfg, err := config.Get()
	if err != nil {
		os.Exit(1)
	}

	cache := freecache.NewCache(cfg.CacheSize)
	// TODO why is this set?
	debug.SetGCPercent(20)

	// run the health check http server.
	healthcheck.NewHandler(cfg.BindAddress, errorChannel)

	datasetAPIClient, err := client.NewDatasetAPIClient(cfg.DatasetAPIURL, cfg.ImportAuthToken, &http.Client{}, ResponseBodyReader{})
	if err != nil {
		// TODO
		os.Exit(1)
	}

	handler.DatasetAPI = datasetAPIClient

	// create the report event kafka consumer.
	consumer, err := kafka.NewConsumerGroup(cfg.Brokers, cfg.NewInstanceTopic, log.Namespace, kafka.OffsetNewest)
	if err != nil {
		logFatal("could not obtain consumer", err, nil)
	}

	// what to do with a report event.
	handleEvent := func(eventMsg kafka.Message) {
		var reportEvent model.EventReport
		if err := schema.ReportEventSchema.Unmarshal(eventMsg.GetData(), &reportEvent); err != nil {
			log.ErrorC("failed to unmarshal message", err, log.Data{"topic": cfg.NewInstanceTopic})
			return
		}

		if err := handler.HandleEvent(cache, cfg, &reportEvent); err != nil {
			log.ErrorC("Failure updating events", err, log.Data{"topic": cfg.NewInstanceTopic})
			return
		}
	}

	// create event consumer.
	reportEventConsumer := event.NewEventConsumer(consumer, handleEvent)
	reportEventConsumer.Listen()

	// shutdown all the things.
	gracefulShutdown := func() {
		log.Info("Attempting graceful shutdown...", nil)

		ctx, cancel := context.WithTimeout(context.Background(), cfg.GracefulShutdownTimeout)
		defer cancel()

		healthcheck.Shutdown(ctx)
		reportEventConsumer.Close(ctx)
		consumer.Close(ctx)
	}

	// block until shutdown event happens...
	select {
	case <-signals:
		gracefulShutdown()
	case err := <-consumer.Errors():
		log.ErrorC("consumer error chan received error commencing graceful shutdown", err, nil)
		gracefulShutdown()
	case err := <-errorChannel:
		log.ErrorC("errors channel received error commencing graceful shutdown", err, nil)
		gracefulShutdown()
	}
}
