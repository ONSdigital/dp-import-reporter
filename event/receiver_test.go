package event

import (
	"errors"
	"github.com/ONSdigital/dp-import-reporter/mocks"
	"github.com/ONSdigital/dp-import-reporter/model"
	"github.com/ONSdigital/dp-import-reporter/schema"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
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

		kafkaMsg := &mocks.KafkaMessageMock{
			GetDataFunc: func() []byte {
				avro, _ := schema.ReportEventSchema.Marshal(e)
				return avro
			},
		}

		handler := &mocks.EventHandlerMock{
			HandleEventFunc: func(e *model.ReportEvent) error {
				return nil
			},
		}

		receiver := Receiver{
			Handler: handler,
		}

		Convey("When ProcessMessage is invoked with a valid message", func() {
			err := receiver.ProcessMessage(kafkaMsg)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And handler.HandleEvent is called once with the expected parameters", func() {
				params := handler.HandleEventCalls()
				So(len(params), ShouldEqual, 1)
				So(params[0].E, ShouldResemble, e)
			})
		})

		Convey("When an invalid message is received", func() {
			kafkaMsg := &mocks.KafkaMessageMock{
				GetDataFunc: func() []byte {
					return []byte("This is not a valid message")
				},
			}

			err := receiver.ProcessMessage(kafkaMsg)

			Convey("Then the expected error is returned", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("And handler.HandleEvent is never called", func() {
				So(len(handler.HandleEventCalls()), ShouldEqual, 0)
			})
		})

		Convey("When Handler.HandleEvent returns an error", func() {
			handler := &mocks.EventHandlerMock{
				HandleEventFunc: func(e *model.ReportEvent) error {
					return errMock
				},
			}

			receiver := Receiver{
				Handler: handler,
			}

			err := receiver.ProcessMessage(kafkaMsg)

			Convey("Then the expected error is returned", func() {
				So(err, ShouldResemble, errMock)
			})

			Convey("And handler.HandleEvent is called once with the expected parameters", func() {
				params := handler.HandleEventCalls()
				So(len(params), ShouldEqual, 1)
				So(params[0].E, ShouldResemble, e)
			})
		})
	})
}
