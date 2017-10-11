package message

import (
	"errors"
	"github.com/ONSdigital/dp-import-reporter/mocks"
	"github.com/ONSdigital/dp-import-reporter/model"
	"github.com/ONSdigital/dp-import-reporter/schema"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

var (
	reportEvent = &model.ReportEvent{
		InstanceID:  "666",
		ServiceName: "super-cool-service",
		EventType:   "error",
		EventMsg:    "Explosions!",
	}
)

func TestMessageConsumerListen(t *testing.T) {
	Convey("Given the consumer is configured correctly", t, func() {

		Convey("When an incoming message is received", func() {
			onCommit := make(chan bool, 1)
			avro, _ := schema.ReportEventSchema.Marshal(reportEvent)
			kafkaMsg, incoming, kafkaConsumer, receiver := setUp(onCommit, avro, nil)

			consumer := NewMessageConsumer(kafkaConsumer, receiver)
			consumer.Listen()
			defer consumer.Close(nil)

			incoming <- kafkaMsg

			select {
			case <-onCommit:
				log.Info("message committed as expected", nil)
			case <-time.After(time.Second * 5):
				log.Info("failing test: expected behaviour did not happen before timeout", nil)
				t.FailNow()
			}

			Convey("Then eventReceiver.ProcessMessage is called once with the expected parameters", func() {
				params := receiver.ProcessMessageCalls()
				So(len(params), ShouldEqual, 1)
				So(params[0].Event, ShouldResemble, kafkaMsg)
			})

			Convey("And eventMsg.Commit is called once", func() {
				So(len(kafkaMsg.CommitCalls()), ShouldEqual, 1)
			})
		})

		Convey("When the eventReceiver.ProcessMessage returns an error", func() {
			onCommit := make(chan bool, 1)
			avro, _ := schema.ReportEventSchema.Marshal(reportEvent)
			handlerErr := errors.New("Flubba Wubba Dub Dub")
			kafkaMsg, incoming, kafkaConsumer, _ := setUp(onCommit, avro, nil)

			onHandle := make(chan bool)

			receiverMock := &mocks.ReceiverMock{
				ProcessMessageFunc: func(event kafka.Message) error {
					go func() {
						<-time.After(time.Second * 2)
						onHandle <- true
					}()
					return handlerErr
				},
			}

			consumer := NewMessageConsumer(kafkaConsumer, receiverMock)
			consumer.Listen()

			incoming <- kafkaMsg

			select {
			case <-onHandle: // wait for onHandle to receive before performing test assertions.
				log.Debug("message handled as expected", nil)
			case <-time.After(time.Second * 5):
				log.Info("failing test: expected behaviour did not happen before timeout", nil)
				t.FailNow()
			}

			consumer.Close(nil)

			Convey("Then eventReceiver.ProcessMessage is called once with the expected parameters", func() {
				params := receiverMock.ProcessMessageCalls()
				So(len(params), ShouldEqual, 1)
				So(params[0].Event, ShouldResemble, kafkaMsg)
			})

			Convey("And eventMsg.Commit is never called", func() {
				So(len(kafkaMsg.CommitCalls()), ShouldEqual, 0)
			})
		})

	})
}

func setUp(onCommit chan bool, avroBytes []byte, handlerErr error) (*mocks.KafkaMessageMock, chan kafka.Message, *mocks.KafkaConsumerMock, *mocks.ReceiverMock) {
	kafkaMsg := &mocks.KafkaMessageMock{
		GetDataFunc: func() []byte {
			return avroBytes
		},
		CommitFunc: func() {
			onCommit <- true
		},
	}

	incomingChan := make(chan kafka.Message, 1)

	consumerMock := &mocks.KafkaConsumerMock{
		IncomingFunc: func() chan kafka.Message {
			return incomingChan
		},
	}

	eventHandler := &mocks.ReceiverMock{
		ProcessMessageFunc: func(event kafka.Message) error {
			return handlerErr
		},
	}
	return kafkaMsg, incomingChan, consumerMock, eventHandler
}