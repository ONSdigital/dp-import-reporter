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
	"github.com/ONSdigital/dp-import-reporter/message"
	"github.com/ONSdigital/dp-import-reporter/server"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/log.go/log"
	"github.com/coocood/freecache"
)

var bufferSize = 1

type ResponseBodyReader struct{}

func (r ResponseBodyReader) Read(reader io.Reader) ([]byte, error) {
	return ioutil.ReadAll(reader)
}

func main() {
	log.Namespace = "dp-import-reporter"

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	errorChannel := make(chan error, 1)

	ctx := context.Background()

	cfg, err := config.Get()
	if err != nil {
		log.Event(ctx, "config.get returned error", log.ERROR, log.Error(err))
		os.Exit(1)
	}

	log.Event(ctx, "successfully loaded dp-import-reporter configuration", log.INFO, log.Data{
		"config": cfg,
	})

	datasetAPIClient, err := client.NewDatasetAPIClient(cfg.ServiceAuthToken, cfg.DatasetAPIURL, cfg.DatasetAPIAuthToken, &http.Client{}, ResponseBodyReader{})
	if err != nil {
		log.Event(ctx, "error creating new dataset api client", log.ERROR, log.Error(err))
		os.Exit(1)
	}

	cache := freecache.NewCache(cfg.CacheSize)
	// TODO why is this set?
	debug.SetGCPercent(20)

	server.Start(ctx, cache, cfg.BindAddress, errorChannel)

	reportEventHandler := event.Handler{
		ExpireSeconds: cfg.CacheExpiry,
		Cache:         cache,
		DatasetAPI:    datasetAPIClient,
	}

	eventReceiver := &event.Receiver{Handler: reportEventHandler}

	kafkaOffset := kafka.OffsetNewest

	if cfg.KafkaOffsetOldest {
		kafkaOffset = kafka.OffsetOldest
	}

	cgConfig := &kafka.ConsumerGroupConfig{
		Offset:       &kafkaOffset,
		KafkaVersion: &cfg.KafkaVersion,
	}

	cgChannels := kafka.CreateConsumerGroupChannels(bufferSize)

	// Create InstanceEvent kafka consumer - exit on channel validation error. Non-initialised consumers will not error at creation time.
	kafkaConsumer, err := kafka.NewConsumerGroup(
		ctx,
		cfg.Brokers,
		cfg.ReportEventTopic,
		cfg.ReportEventGroup,
		cgChannels,
		cgConfig,
	)

	// create the report event kafka kafkaConsumer.
	if err != nil {
		log.Event(ctx, "error while attempting to create kafka kafkaConsumer", log.ERROR, log.Error(err))
		os.Exit(1)
	}

	// create event kafkaConsumer.
	reportEventConsumer := message.NewConsumer(kafkaConsumer, eventReceiver, cfg.GracefulShutdownTimeout)

	reportEventConsumer.Listen(ctx)

	// block until a fatal event happens
	select {
	case sig := <-signals:
		log.Event(ctx, "os signal received commencing graceful shutdown", log.INFO, log.Data{"signal": sig.String()})
	case err = <-kafkaConsumer.Channels().Errors:
		log.Event(ctx, "kafkaConsumer errors chan received an error commencing graceful shutdown", log.ERROR, log.Error(err))
	case err = <-errorChannel:
		log.Event(ctx, "errors channel received an error commencing graceful shutdown", log.ERROR, log.Error(err))
	}

	log.Event(ctx, "attempting graceful shutdown of service", log.INFO)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.GracefulShutdownTimeout)
	go func() {
		defer cancel()

		reportEventConsumer.Close(ctx)
		server.Shutdown(ctx)
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()
	if err == nil && ctx.Err() != context.Canceled {
		err = ctx.Err()
	}
	log.Event(ctx, "shutdown complete", log.INFO)
	if err != nil {
		os.Exit(1)
	}
}
