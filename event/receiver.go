package event

import (
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/dp-import-reporter/model"
	"github.com/ONSdigital/dp-import-reporter/schema"
	"github.com/ONSdigital/go-ns/log"
)

type EventHandler interface {
	HandleEvent(e *model.ReportEvent) error
}

type Receiver struct {
	Handler EventHandler
}

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
