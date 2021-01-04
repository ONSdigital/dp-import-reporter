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

			kafkaConsumer.Channels().Upstream <- kafkaMsg

			select {
			case <-kafkaMsg.UpstreamDone():
				log.Event(ctx, "message committed as expected", log.INFO)
			case <-time.After(time.Second * 5):
				log.Event(ctx, "failing test: expected behaviour failed to happen before timeout", log.INFO)
				t.FailNow()
			}

			consumer.Close(nil)
			<-consumer.closed

			Convey("Then eventReceiver.ProcessMessage is called once with the expected parameters", func() {
				calls := receiver.ProcessMessageCalls()
				So(calls, ShouldHaveLength, 1)
				So(calls[0].Event, ShouldResemble, kafkaMsg)
			})

			Convey("And consumer.CommitAndRelease is called once", func() {
				So(len(kafkaMsg.CommitAndReleaseCalls()), ShouldEqual, 1)
			})
		})

		Convey("When the eventReceiver.ProcessMessage returns an error", func() {
			avro, _ := schema.ReportEventSchema.Marshal(reportEvent)
			handlerErr := errors.New("Flubba Wubba Dub Dub")
			kafkaMsg, kafkaConsumer, receiver := setUp(avro, handlerErr)

			consumer := NewConsumer(kafkaConsumer, receiver, time.Second*10)
			consumer.Listen(ctx)

			kafkaConsumer.Channels().Upstream <- kafkaMsg

			select {
			case <-kafkaMsg.UpstreamDone():
				log.Event(ctx, "message released", log.INFO)
			case <-time.After(time.Second * 5):
				log.Event(ctx, "failing test: expected behaviour did not happen before timeout", log.INFO)
				t.FailNow()
			}

			consumer.Close(nil)
			<-consumer.closed

			Convey("Then eventReceiver.ProcessMessage is called once with the expected parameters", func() {
				calls := receiver.ProcessMessageCalls()
				So(calls, ShouldHaveLength, 1)
				So(calls[0].Event, ShouldResemble, kafkaMsg)
			})

			Convey("And eventMsg.Commit is called", func() {
				So(len(kafkaMsg.CommitAndReleaseCalls()), ShouldEqual, 1)
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
