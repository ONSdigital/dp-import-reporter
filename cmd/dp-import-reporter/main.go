package main

import (
	"context"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"io"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/dp-import-reporter/client"
	"github.com/ONSdigital/dp-import-reporter/config"
	"github.com/ONSdigital/dp-import-reporter/event"
	"github.com/ONSdigital/dp-import-reporter/logging"
	"github.com/ONSdigital/dp-import-reporter/message"
	"github.com/ONSdigital/dp-import-reporter/server"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
	"github.com/coocood/freecache"
)

var logger = logging.Logger{Prefix: "main"}

type ResponseBodyReader struct{}

func (r ResponseBodyReader) Read(reader io.Reader) ([]byte, error) {
	return ioutil.ReadAll(reader)
}

func main() {
	log.Namespace = "dp-import-reporter"

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	errorChannel := make(chan error, 1)

	cfg, err := config.Get()
	if err != nil {
		log.ErrorC("config.get returned error", err, nil)
		os.Exit(1)
	}

	logger.Info("successfully loaded dp-import-reporter configuration", log.Data{
		"config": cfg,
	})

	datasetAPIClient, err := client.NewDatasetAPIClient(cfg.ServiceAuthToken, cfg.DatasetAPIURL, cfg.DatasetAPIAuthToken, &http.Client{}, ResponseBodyReader{})
	if err != nil {
		log.ErrorC("error creating new dataset api client", err, nil)
		os.Exit(1)
	}

	cache := freecache.NewCache(cfg.CacheSize)
	// TODO why is this set?
	debug.SetGCPercent(20)

	server.Start(cache, cfg.BindAddress, errorChannel)

	reportEventHandler := event.Handler{
		ExpireSeconds: cfg.CacheExpiry,
		Cache:         cache,
		DatasetAPI:    datasetAPIClient,
	}

	eventReceiver := &event.Receiver{Handler: reportEventHandler}

	// create the report event kafka kafkaConsumer.
	kafkaConsumer, err := kafka.NewConsumerGroup(cfg.Brokers, cfg.ReportEventTopic, cfg.ReportEventGroup, kafka.OffsetNewest)
	if err != nil {
		log.ErrorC("error while attempting to create kafka kafkaConsumer", err, nil)
		os.Exit(1)
	}

	// create event kafkaConsumer.
	reportEventConsumer := message.NewConsumer(kafkaConsumer, eventReceiver, cfg.GracefulShutdownTimeout)

	reportEventConsumer.Listen()

	// block until a shutdown event happens
	select {
	case sig := <-signals:
		logger.Info("os signal received commencing graceful shutdown", log.Data{"signal": sig.String()})
	case err := <-kafkaConsumer.Errors():
		logger.ErrorC("kafkaConsumer errors chan received an error commencing graceful shutdown", err, nil)
	case err := <-errorChannel:
		logger.ErrorC("errors channel received an error commencing graceful shutdown", err, nil)
	}

	logger.Info("attempting graceful shutdown of service", nil)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.GracefulShutdownTimeout)
	defer cancel()

	reportEventConsumer.Close(ctx)
	kafkaConsumer.Close(ctx)
	server.Shutdown(ctx)

	logger.Info("shutdown complete", nil)
	os.Exit(1)
}
