package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/ONSdigital/dp-import-reporter/config"
	"github.com/ONSdigital/dp-import-reporter/handler"
	"github.com/ONSdigital/go-ns/errorhandler/models"
	"github.com/ONSdigital/go-ns/errorhandler/schema"

	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
	"github.com/coocood/freecache"
	"github.com/ONSdigital/dp-import-reporter/healthcheck"
)

// logFatal is a utility method for a common failure pattern in main()
func logFatal(context string, err error, data log.Data) {
	log.ErrorC(context, err, data)
	os.Exit(1)
}

func main() {
	log.Namespace = "dp-event-reporter"

	cfg, err := config.Get()
	if err != nil {
		os.Exit(1)
	}

	newInstanceEventConsumer, err := consumerInit(cfg)
	if err != nil {
		logFatal("error initiating", err, nil)
	}

	consume(newInstanceEventConsumer, cfg)

}

func consumerInit(cfg *config.Config) (*kafka.ConsumerGroup, error) {

	log.Info("starting", log.Data{
		"new-import-topic": cfg.NewInstanceTopic,
		"import-api":       cfg.DatasetAPIURL,
	})
	consumer, err := kafka.NewConsumerGroup(cfg.Brokers, cfg.NewInstanceTopic, log.Namespace, kafka.OffsetNewest)
	if err != nil {
		logFatal("could not obtain consumer", err, nil)
	}
	return consumer, err
}

func cacheSetup(cfg *config.Config) *freecache.Cache {
	c := freecache.NewCache(cfg.CacheSize)
	debug.SetGCPercent(20)
	return c
}

func consume(newInstanceEventConsumer *kafka.ConsumerGroup, cfg *config.Config) {
	running := true
	errorChannel := make(chan error)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	//cache init
	c := cacheSetup(cfg)

	healthcheck.NewHandler(cfg.BindAddress, errorChannel)

	go func() {
		for running {
			select {
			case newInstanceMessage := <-newInstanceEventConsumer.Incoming():
				var msg errorModel.EventReport
				if err := errorschema.ReportedEventSchema.Unmarshal(newInstanceMessage.GetData(), &msg); err != nil {
					log.ErrorC("failed to unmarshal message", err, log.Data{"topic": cfg.NewInstanceTopic})
					// Fatal error reading message, should never fall in here
					continue
				}

				if err := handler.HandleEvent(c, cfg, &msg); err != nil {
					log.ErrorC("Failure updating events", err, log.Data{"topic": cfg.NewInstanceTopic})
					continue
				}
				newInstanceMessage.Commit()

			case newImportConsumerErrorMessage := <-newInstanceEventConsumer.Errors():
				log.Error(errors.New("consumer recieved error: "), log.Data{"error": newImportConsumerErrorMessage, "topic": cfg.NewInstanceTopic})
				errorChannel <- newImportConsumerErrorMessage
			case <-errorChannel:
				log.ErrorC("Error channel..", errors.New("Errors"), nil)
				shutdownGracefully(newInstanceEventConsumer, cfg)
			case <-signals:
				log.ErrorC("Signal was sent to application", errors.New("signal passed to application"), nil)
				shutdownGracefully(newInstanceEventConsumer, cfg)
			}
		}
	}()
	<-errorChannel

	shutdownGracefully(newInstanceEventConsumer, cfg)

}

func shutdownGracefully(consumer *kafka.ConsumerGroup, cfg *config.Config) {
	log.Info("Attempting graceful shut down...", nil)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.GracefulShutdownTimeout)
	err := consumer.Close(ctx)
	if err != nil {
		log.Error(err, nil)
	}

	healthcheck.Shutdown(ctx)

	cancel()
	log.Info("Gracefully shut down application", nil)

	os.Exit(1)
}
