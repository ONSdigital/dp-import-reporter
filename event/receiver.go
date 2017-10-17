package event

import (
	"github.com/ONSdigital/dp-import-reporter/model"
	"github.com/ONSdigital/dp-import-reporter/schema"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/pkg/errors"
	"github.com/ONSdigital/dp-import-reporter/logging"
)

var receiverLog = logging.Logger{Prefix: "event.Receiver"}

// EventHandler a handler that processes a reported event
type EventHandler interface {
	HandleEvent(e *model.ReportEvent) error
}

// Receiver receives an coming kafka messages, extract the message body and pass to an event handler to process to the event
type Receiver struct {
	Handler EventHandler
}

// ProcessMessage extract the kafka message body and process it
func (r *Receiver) ProcessMessage(msg kafka.Message) error {
	var reportEvent model.ReportEvent
	receiverLog.Info("attempting to unmarshal kafka message", nil)
	if err := schema.ReportEventSchema.Unmarshal(msg.GetData(), &reportEvent); err != nil {
		return errors.Wrap(err, "error while attempting to unmarshal reportEvent from avro")
	}
	receiverLog.Info("successfully unmarhsalled kafka message to reportEvent", nil)
	if err := r.Handler.HandleEvent(&reportEvent); err != nil {
		return errors.Wrap(err, "Handler.HandleEvent returned an error")
	}
	receiverLog.Info("report event handled successfully", nil)
	return nil
}
