package main

import (
	"errors"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/ONSdigital/dp-import-reporter/config"

	"github.com/ONSdigital/dp-import-reporter/handler"
	"github.com/ONSdigital/dp-import-reporter/schema"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
	"github.com/coocood/freecache"
)

// logFatal is a utility method for a common failure pattern in main()
func logFatal(context string, err error, data log.Data) {
	log.ErrorC(context, err, data)
	os.Exit(1)
}

func main() {

	log.Namespace = "dp-import-api"

	cfg, err := config.Get()
	if err != nil {
		logFatal("Config was not configured correctly: ", err, nil)
	}

	newInstanceEventConsumer, err := consumerInit(cfg)
	if err != nil {
		logFatal("error initiating", err, nil)
	}
	//cache init
	c := cacheSetup(cfg)
	client := &http.Client{}

	consume(newInstanceEventConsumer, cfg, client, c)
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

func consume(newInstanceEventConsumer *kafka.ConsumerGroup, cfg *config.Config, client *http.Client, c *freecache.Cache) {
	running := true
	errorChannel := make(chan bool)
	go func() {
		for running {
			select {
			case newInstanceMessage := <-newInstanceEventConsumer.Incoming():
				var msg handler.EventReport
				if err := schema.ReportedEventSchema.Unmarshal(newInstanceMessage.GetData(), &msg); err != nil {
					log.ErrorC("failed to unmarshal message", err, log.Data{"topic": cfg.NewInstanceTopic})
					// Fatal error reading message, should never fall in here
					continue
				}
				if err := msg.HandleEvent(client, c, cfg); err != nil {
					log.ErrorC("Failure updating events", err, log.Data{"topic": cfg.NewInstanceTopic})
					continue
				}
				newInstanceMessage.Commit()
			case newImportConsumerErrorMessage := <-newInstanceEventConsumer.Errors():
				log.Error(errors.New("consumer recieved error: "), log.Data{"error": newImportConsumerErrorMessage, "topic": cfg.NewInstanceTopic})
				running = false
				errorChannel <- true
			}
		}
	}()
	<-errorChannel
	// assert: only get here when we have an error, which has been logged
	newInstanceEventConsumer.Closer() <- true
	logFatal("gracefully shutting down application...", errors.New("Aborting application, gracfully shutting down"), nil)

}
