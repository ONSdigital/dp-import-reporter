package message

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ONSdigital/dp-import-reporter/mocks"
	"github.com/ONSdigital/dp-import-reporter/model"
	"github.com/ONSdigital/dp-import-reporter/schema"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-kafka/v2/kafkatest"
	"github.com/ONSdigital/log.go/log"
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

var ctx = context.Background()

func TestMessageConsumerListen(t *testing.T) {
	Convey("Given the consumer is configured correctly", t, func() {

		Convey("When an incoming message is received", func() {
			avro, _ := schema.ReportEventSchema.Marshal(reportEvent)
			kafkaMsg, kafkaConsumer, receiver := setUp(avro, nil)

			consumer := NewConsumer(kafkaConsumer, receiver, time.Second*10)
			consumer.Listen(ctx)
			defer consumer.Close(nil)

			kafkaConsumer.Channels().Upstream <- kafkaMsg

			select {
			case <-kafkaMsg.UpstreamDone():
				log.Event(ctx, "message committed as expected", log.INFO)
			case <-time.After(time.Second * 5):
				log.Event(ctx, "failing test: expected behaviour failed to happen before timeout", log.INFO)
				t.FailNow()
			}

			Convey("Then eventReceiver.ProcessMessage is called once with the expected parameters", func() {
				params := receiver.ProcessMessageCalls()
				So(len(params), ShouldEqual, 1)
				So(params[0].Event, ShouldResemble, kafkaMsg)
			})

			Convey("And consumer.CommitAndRelease is called once", func() {
				So(len(kafkaMsg.CommitAndReleaseCalls()), ShouldEqual, 1)
			})
		})

		Convey("When the eventReceiver.ProcessMessage returns an error", func() {
			avro, _ := schema.ReportEventSchema.Marshal(reportEvent)
			handlerErr := errors.New("Flubba Wubba Dub Dub")
			kafkaMsg, kafkaConsumer, _ := setUp(avro, nil)

			onHandle := make(chan bool)

			receiverMock := &mocks.ReceiverMock{
				ProcessMessageFunc: func(ctx context.Context, event kafka.Message) error {
					go func() {
						<-time.After(time.Second * 2)
						onHandle <- true
					}()
					return handlerErr
				},
			}

			consumer := NewConsumer(kafkaConsumer, receiverMock, time.Second*10)
			consumer.Listen(ctx)

			kafkaConsumer.Channels().Upstream <- kafkaMsg

			select {
			case <-onHandle: // wait for onHandle to receive before performing test assertions.
				log.Event(ctx, "message handled as expected", log.INFO)
			case <-time.After(time.Second * 5):
				log.Event(ctx, "failing test: expected behaviour did not happen before timeout", log.INFO)
				t.FailNow()
			}

			consumer.Close(nil)

			Convey("Then eventReceiver.ProcessMessage is called once with the expected parameters", func() {
				params := receiverMock.ProcessMessageCalls()
				So(len(params), ShouldEqual, 1)
				So(params[0].Event, ShouldResemble, kafkaMsg)
			})

			Convey("And eventMsg.Commit is never called", func() {
				So(len(kafkaMsg.CommitAndReleaseCalls()), ShouldEqual, 0)
			})
		})

	})
}

func setUp(avroBytes []byte, handlerErr error) (*kafkatest.Message, *kafkatest.MessageConsumer, *mocks.ReceiverMock) {

	kafkaMsg := kafkatest.NewMessage(avroBytes, 0)

	consumerMock := kafkatest.NewMessageConsumer(true)

	eventHandler := &mocks.ReceiverMock{
		ProcessMessageFunc: func(ctx context.Context, event kafka.Message) error {
			return handlerErr
		},
	}
	return kafkaMsg, consumerMock, eventHandler
}
