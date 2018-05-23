package message

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ONSdigital/dp-import-reporter/mocks"
	"github.com/ONSdigital/dp-import-reporter/model"
	"github.com/ONSdigital/dp-import-reporter/schema"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
	. "github.com/smartystreets/goconvey/convey"
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
			kafkaMsg, incoming, kafkaConsumer, receiver, _ := setUp(onCommit, avro, nil)

			consumer := NewConsumer(kafkaConsumer, receiver, time.Second*10)
			consumer.Listen()
			defer consumer.Close(nil)

			incoming <- kafkaMsg

			select {
			case <-onCommit:
				log.Info("message committed as expected", nil)
			case <-time.After(time.Second * 5):
				log.Info("failing test: expected behaviour failed to happen before timeout", nil)
				t.FailNow()
			}

			Convey("Then eventReceiver.ProcessMessage is called once with the expected parameters", func() {
				params := receiver.ProcessMessageCalls()
				So(len(params), ShouldEqual, 1)
				So(params[0].Event, ShouldResemble, kafkaMsg)
			})

			Convey("And consumer.CommitAndRelease is called once", func() {
				So(len(kafkaConsumer.CommitAndReleaseCalls()), ShouldEqual, 1)
			})
		})

		Convey("When the eventReceiver.ProcessMessage returns an error", func() {
			onCommit := make(chan bool, 1)
			avro, _ := schema.ReportEventSchema.Marshal(reportEvent)
			handlerErr := errors.New("Flubba Wubba Dub Dub")
			kafkaMsg, incoming, kafkaConsumer, _, _ := setUp(onCommit, avro, nil)

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

			consumer := NewConsumer(kafkaConsumer, receiverMock, time.Second*10)
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
				So(len(kafkaConsumer.CommitAndReleaseCalls()), ShouldEqual, 0)
			})
		})

	})
}

func setUp(onCommit chan bool, avroBytes []byte, handlerErr error) (*mocks.KafkaMessageMock, chan kafka.Message, *mocks.KafkaConsumerMock, *mocks.ReceiverMock, chan error) {
	kafkaMsg := &mocks.KafkaMessageMock{
		GetDataFunc: func() []byte {
			return avroBytes
		},
	}

	incomingChan := make(chan kafka.Message, 1)
	errorsChan := make(chan error, 1)

	consumerMock := &mocks.KafkaConsumerMock{
		IncomingFunc: func() chan kafka.Message {
			return incomingChan
		},
		CommitAndReleaseFunc: func(event kafka.Message) {
			onCommit <- true
		},
		ErrorsFunc: func() chan error {
			return errorsChan
		},
		StopListeningToConsumerFunc: func(ctx context.Context) error {
			return nil
		},
		CloseFunc: func(ctx context.Context) error {
			return nil
		},
	}

	eventHandler := &mocks.ReceiverMock{
		ProcessMessageFunc: func(event kafka.Message) error {
			return handlerErr
		},
	}
	return kafkaMsg, incomingChan, consumerMock, eventHandler, errorsChan
}
