package main

import (
	"context"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/ONSdigital/dp-import-reporter/client"
	"github.com/ONSdigital/dp-import-reporter/config"
	"github.com/ONSdigital/dp-import-reporter/event"
	"github.com/ONSdigital/dp-import-reporter/message"
	"github.com/ONSdigital/dp-import-reporter/server"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
	"github.com/coocood/freecache"
	"io"
	"io/ioutil"
	"net/http"
)

func main() {
	log.Namespace = "dp-event-reporter"

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	errorChannel := make(chan error, 1)

	cfg, err := config.Get()
	if err != nil {
		log.ErrorC("config.get retruned error", err, nil)
		os.Exit(1)
	}

	log.Info("dp-import-reporter config", log.Data{
		"config": cfg,
	})

	datasetAPIClient, err := client.NewDatasetAPIClient(cfg.DatasetAPIURL, cfg.DatasetAPIAuthToken, &http.Client{}, ResponseBodyReader{})
	if err != nil {
		log.ErrorC("error creating new dataset api client", err, nil)
		os.Exit(1)
	}

	cache := freecache.NewCache(cfg.CacheSize)
	// TODO why is this set?
	debug.SetGCPercent(20)

	server.Start(cache, cfg.BindAddress, errorChannel)

	reportEventHandler := event.Handler{
		Cache:         cache,
		DatasetAPI:    datasetAPIClient,
		ExpireSeconds: cfg.CacheExpiry,
	}

	eventReceiver := &event.Receiver{
		Handler: reportEventHandler,
	}

	// create the report event kafka kafkaConsumer.
	kafkaConsumer, err := kafka.NewConsumerGroup(cfg.Brokers, cfg.ReportEventTopic, log.Namespace, kafka.OffsetNewest)
	if err != nil {
		log.ErrorC("error while attempting to create kafka kafkaConsumer", err, nil)
		os.Exit(1)
	}

	// create event kafkaConsumer.
	reportEventConsumer := message.NewMessageConsumer(kafkaConsumer, eventReceiver)
	reportEventConsumer.Listen()

	// shutdown all the things.
	gracefulShutdown := func() {
		log.Info("attempting graceful shutdown of service", nil)

		ctx, cancel := context.WithTimeout(context.Background(), cfg.GracefulShutdownTimeout)
		defer cancel()

		reportEventConsumer.Close(ctx)
		kafkaConsumer.Close(ctx)
		server.Shutdown(ctx)
	}

	// block until shutdown event happens...
	select {
	case <-signals:
		gracefulShutdown()
	case err := <-kafkaConsumer.Errors():
		log.ErrorC("kafkaConsumer error chan received error commencing graceful shutdown", err, nil)
		gracefulShutdown()
	case err := <-errorChannel:
		log.ErrorC("errors channel received error commencing graceful shutdown", err, nil)
		gracefulShutdown()
	}
}

type ResponseBodyReader struct{}

func (r ResponseBodyReader) Read(reader io.Reader) ([]byte, error) {
	return ioutil.ReadAll(reader)
}
