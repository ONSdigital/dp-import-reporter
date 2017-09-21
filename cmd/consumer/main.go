package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/ONSdigital/dp-import-reporter/config"
	"github.com/gorilla/mux"

	"github.com/ONSdigital/dp-import-reporter/handler"
	"github.com/ONSdigital/dp-import-reporter/schema"
	"github.com/ONSdigital/go-ns/handlers/healthcheck"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/coocood/freecache"
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
		logFatal("Config was not configured correctly: ", err, nil)
	}

	newInstanceEventConsumer, err := consumerInit(cfg)
	if err != nil {
		logFatal("error initiating", err, nil)
	}

	consume(newInstanceEventConsumer, cfg)

}
func serverRunner(cfg *config.Config, errorChannel chan bool) *server.Server {
	router := mux.NewRouter()
	router.Path("/healthcheck").HandlerFunc(healthcheck.Handler)
	httpServer := server.New(cfg.BindAddress, router)

	//disable autohandling of os siginals by server
	httpServer.HandleOSSignals = false
	go func() {
		log.Debug("Starting http server", log.Data{"bind_addr": cfg.BindAddress})
		if err := httpServer.ListenAndServe(); err != nil {
			errorChannel <- true
		}
	}()

	return httpServer
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
	errorChannel := make(chan bool)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	//cache init
	c := cacheSetup(cfg)
	httpServer := serverRunner(cfg, errorChannel)
	go func() {
		for running {
			select {
			case newInstanceMessage := <-newInstanceEventConsumer.Incoming():
				var msg handler.EventReport
				if err := schema.ReportedEventSchema.Unmarshal(newInstanceMessage.GetData(), &msg); err != nil {
					log.ErrorC("failed to unmarshal message", err, log.Data{"topic": cfg.NewInstanceTopic})
					// Fatal error reading message, should never fall in here
					errorChannel <- true
					continue
				}
				if err := msg.HandleEvent(c, cfg); err != nil {
					log.ErrorC("Failure updating events", err, log.Data{"topic": cfg.NewInstanceTopic})
					continue
				}
				newInstanceMessage.Commit()
			case newImportConsumerErrorMessage := <-newInstanceEventConsumer.Errors():
				log.Error(errors.New("consumer recieved error: "), log.Data{"error": newImportConsumerErrorMessage, "topic": cfg.NewInstanceTopic})
				errorChannel <- true
			case <-errorChannel:
				log.ErrorC("Error channel..", errors.New("Errors"), nil)
				shutdownGracefully(newInstanceEventConsumer, httpServer, cfg)
			case <-signals:
				log.ErrorC("Signal was sent to application", errors.New("signal passed to application"), nil)
				shutdownGracefully(newInstanceEventConsumer, httpServer, cfg)
			}
		}
	}()
	<-errorChannel
	// running = false

	shutdownGracefully(newInstanceEventConsumer, httpServer, cfg)

	// assert: only get here when we have an error, which has been logged
	// newInstanceEventConsumer.Closer() <- true
	// logFatal("gracefully shutting down application...", errors.New("Aborting application, gracfully shutting down"), nil)

}

func shutdownGracefully(consumer *kafka.ConsumerGroup, httpServer *server.Server, cfg *config.Config) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.GracefulShutdownTimeout)
	err := consumer.Close(ctx)
	if err != nil {
		log.Error(err, nil)
	}

	err = httpServer.Shutdown(ctx)
	if err != nil {
		log.Error(err, nil)
	}

	cancel()
	log.Info("Gracefully shut down application...", nil)

	os.Exit(1)
}
