package main

import (
	"net/http"
	"testing"

	"github.com/ONSdigital/dp-import-reporter/config"
	"github.com/ONSdigital/dp-import-reporter/handler"
	"github.com/ONSdigital/dp-import-reporter/schema"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
	. "github.com/smartystreets/goconvey/convey"
)

type kafkaMsg struct {
	myBytes []byte
	commitI int
}

func (k kafkaMsg) GetData() []byte {

	return k.myBytes
}

func (k kafkaMsg) Commit() {
	k.commitI++
}

func TestConsumer(t *testing.T) {
	// t.Parallel()

	Convey("Set up the variables for test enviroment", t, func() {
		cfg, _ := config.Get()
		log.Namespace = "dp-event-reporter"
		msg, _ := schema.ReportedEventSchema.Marshal(&handler.EventReport{
			InstanceID: "a4695fee-f0a2-49c4-b136-e3ca8dd40476",
			EventType:  "error",
			EventMsg:   "Broken on something.",
		})

		newInstanceEventConsumer, err := kafka.NewConsumerGroup(cfg.Brokers, cfg.NewInstanceTopic, log.Namespace, kafka.OffsetNewest)
		Convey("Check error for kafka consumer group", func() {
			So(err, ShouldEqual, nil)
		})
		Convey("Check that the newInstanceEvent matches", func() {
			So(newInstanceEventConsumer, ShouldEqual, newInstanceEventConsumer)
		})
		client := &http.Client{}

		c := cacheSetup(cfg)

		kafkaMessege := &kafkaMsg{
			myBytes: msg,
			commitI: 0,
		}

		go loop(newInstanceEventConsumer, cfg, client, c)
		newInstanceEventConsumer.Incoming() <- *kafkaMessege

		//incorrect instance id will throw error.
		msg1, err1 := schema.ReportedEventSchema.Marshal(&handler.EventReport{
			InstanceID: "a4695fee-f0a2-49c4-b136-e3ca822dd40476",
			EventType:  "error",
			EventMsg:   "Broken on something.",
		})
		Convey("Check error when marshalling", func() {
			So(err1, ShouldBeNil)
		})
		kafkaMessege1 := &kafkaMsg{
			myBytes: msg1,
			commitI: 0,
		}
		newInstanceEventConsumer.Incoming() <- *kafkaMessege1

	})
}
func TestInit(t *testing.T) {
	Convey("Should init the parameters needed for the running of the consumer", t, func() {
		cfg, _ := config.Get()
		newKakfaConsumer, err := consumerInit(cfg)
		So(newKakfaConsumer, ShouldNotBeNil)
		So(err, ShouldBeNil)

	})
}
