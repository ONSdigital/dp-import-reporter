package event

import (
	"github.com/ONSdigital/dp-import-reporter/model"
	"github.com/ONSdigital/dp-import-reporter/schema"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
)

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
	if err := schema.ReportEventSchema.Unmarshal(msg.GetData(), &reportEvent); err != nil {
		log.ErrorC("unexpected error while attempting to unmarshal e kafka message.", err, nil)
		return err
	}

	if err := r.Handler.HandleEvent(&reportEvent); err != nil {
		log.ErrorC("unexpected error returned from e Handler", err, nil)
		return err
	}
	return nil
}
