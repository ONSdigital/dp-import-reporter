package main

import (
	"testing"

	"github.com/ONSdigital/dp-import-reporter/config"
	"github.com/ONSdigital/go-ns/errorhandler/models"

	"github.com/ONSdigital/go-ns/errorhandler/schema"

	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"

	. "github.com/smartystreets/goconvey/convey"
)

//TODO all the commented out sections need to be mocked/ part of the mocking process

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
		// msg, _ := errorschema.ReportedEventSchema.Marshal(&errorModel.EventReport{
		// 	InstanceID: "a4695fee-f0a2-49c4-b136-e3ca8dd40476",
		// 	EventType:  "error",
		// 	EventMsg:   "Broken on something.",
		// })

		newInstanceEventConsumer, err := kafka.NewConsumerGroup(cfg.Brokers, cfg.NewInstanceTopic, log.Namespace, kafka.OffsetNewest)
		Convey("Check error for kafka consumer group", func() {
			So(err, ShouldEqual, nil)
		})
		Convey("Check that the newInstanceEvent matches", func() {
			So(newInstanceEventConsumer, ShouldEqual, newInstanceEventConsumer)
		})

		// kafkaMessege := &kafkaMsg{
		// 	myBytes: msg,
		// 	commitI: 0,
		// }

		// go consume(newInstanceEventConsumer, cfg)
		// newInstanceEventConsumer.Incoming() <- *kafkaMessege

		//incorrect instance id will throw error.
		_, err1 := errorschema.ReportedEventSchema.Marshal(&errorModel.EventReport{
			InstanceID: "a4695fee-f0a2-49c4-b136-e3ca822dd40476",
			EventType:  "error",
			EventMsg:   "Broken on something.",
		})
		Convey("Check error when marshalling", func() {
			So(err1, ShouldBeNil)
		})

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

// func TestConsumerError(t *testing.T) {
// 	Convey("When an error message is sent to the error collector", t, func() {
// 		cfg, _ := config.Get()
//
// 		newInstanceEventConsumer, err := kafka.NewConsumerGroup(cfg.Brokers, cfg.NewInstanceTopic, log.Namespace, kafka.OffsetNewest)
// 		Convey("Check error for kafka consumer group", func() {
// 			So(err, ShouldEqual, nil)
// 		})
// 		Convey("Check that the newInstanceEvent matches", func() {
// 			So(newInstanceEventConsumer, ShouldEqual, newInstanceEventConsumer)
// 		})
// 		go consume(newInstanceEventConsumer, cfg)
// 		newInstanceEventConsumer.Errors() <- errors.New("AN ERROR")
//
// 	})
// }
