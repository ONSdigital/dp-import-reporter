package event

import (
	"context"

	"github.com/ONSdigital/dp-import-reporter/model"
	"github.com/ONSdigital/dp-import-reporter/schema"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/pkg/errors"
)

// EventHandler a handler that processes a reported event
type EventHandler interface {
	HandleEvent(ctx context.Context, e *model.ReportEvent) error
}

// Receiver receives an coming kafka messages, extract the message body and pass to an event handler to process to the event
type Receiver struct {
	Handler EventHandler
}

// ProcessMessage extract the kafka message body and process it
func (r *Receiver) ProcessMessage(ctx context.Context, msg kafka.Message) error {
	var reportEvent model.ReportEvent
	log.Info(ctx, "attempting to unmarshal kafka message")
	if err := schema.ReportEventSchema.Unmarshal(msg.GetData(), &reportEvent); err != nil {
		return errors.Wrap(err, "error while attempting to unmarshal reportEvent from avro")
	}
	log.Info(ctx, "successfully unmarhsalled kafka message to reportEvent")
	if err := r.Handler.HandleEvent(ctx, &reportEvent); err != nil {
		return errors.Wrap(err, "Handler.HandleEvent returned an error")
	}
	log.Info(ctx, "report event handled successfully")
	return nil
}
