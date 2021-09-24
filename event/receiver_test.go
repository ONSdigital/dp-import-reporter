package event

import (
	"context"
	"testing"

	"github.com/ONSdigital/dp-import-reporter/model"
	"github.com/ONSdigital/dp-import-reporter/schema"
	"github.com/ONSdigital/dp-kafka/v2/kafkatest"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

var errMock = errors.New("broken")

func TestReceiverProcessMessage(t *testing.T) {

	Convey("Given a correctly configured Receiver", t, func() {
		e := &model.ReportEvent{
			EventMsg:    "message",
			EventType:   "error",
			ServiceName: "serviceName",
			InstanceID:  "666",
		}

		avro, _ := schema.ReportEventSchema.Marshal(e)
		kafkaMsg := kafkatest.NewMessage(avro, 0)

		handler := &EventHandlerMock{
			HandleEventFunc: func(ctx context.Context, e *model.ReportEvent) error {
				return nil
			},
		}

		receiver := Receiver{
			Handler: handler,
		}

		Convey("When ProcessMessage is invoked with a valid message", func() {
			err := receiver.ProcessMessage(ctx, kafkaMsg)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And Handler.HandleEvent is called once with the expected parameters", func() {
				params := handler.HandleEventCalls()
				So(len(params), ShouldEqual, 1)
				So(params[0].E, ShouldResemble, e)
			})
		})

		Convey("When an invalid message is received", func() {

			kafkaMsg := kafkatest.NewMessage([]byte("This is not a valid message"), 0)

			err := receiver.ProcessMessage(ctx, kafkaMsg)

			Convey("Then the expected error is returned", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("And Handler.HandleEvent is never called", func() {
				So(len(handler.HandleEventCalls()), ShouldEqual, 0)
			})
		})

		Convey("When Handler.HandleEvent returns an error", func() {
			handler := &EventHandlerMock{
				HandleEventFunc: func(ctx context.Context, e *model.ReportEvent) error {
					return errMock
				},
			}

			receiver := Receiver{
				Handler: handler,
			}

			err := receiver.ProcessMessage(ctx, kafkaMsg)

			Convey("Then the expected error is returned", func() {
				So(err.Error(), ShouldEqual, errors.Wrap(errMock, "Handler.HandleEvent returned an error").Error())
			})

			Convey("And Handler.HandleEvent is called once with the expected parameters", func() {
				params := handler.HandleEventCalls()
				So(len(params), ShouldEqual, 1)
				So(params[0].E, ShouldResemble, e)
			})
		})
	})
}
