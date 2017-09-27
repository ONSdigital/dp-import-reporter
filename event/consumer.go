package event

import (
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
	"context"
	"time"
)

type HandleEvent func(event kafka.Message)

type EventConsumer struct {
	closed      chan bool
	ctx         context.Context
	cancel      context.CancelFunc
	consumer    *kafka.ConsumerGroup
	handleEvent HandleEvent
}

func NewEventConsumer(consumer *kafka.ConsumerGroup, handleEvent HandleEvent) EventConsumer {
	ctx, cancel := context.WithCancel(context.Background())

	return EventConsumer{
		closed:      make(chan bool, 1),
		ctx:         ctx,
		cancel:      cancel,
		consumer:    consumer,
		handleEvent: handleEvent,
	}
}

func (c *EventConsumer) Listen() {
	defer func() {
		c.closed <- true
	}()

	go func() {
		for {
			select {
			case eventMsg := <-c.consumer.Incoming():
				c.handleEvent(eventMsg)
				eventMsg.Commit()
			case <-c.ctx.Done():
				log.Info("attempting to close down consumer", nil)
				return
			}
		}
	}()
}

// Close shutdown Consumer.Listen loop.
func (c EventConsumer) Close(ctx context.Context) {
	// if nil use a default context with a timeout
	if ctx == nil {
		ctx, _ = context.WithTimeout(context.Background(), time.Second*10)
	}

	// if the context is not nil but no deadline is set the apply the default
	if _, ok := ctx.Deadline(); !ok {
		ctx, _ = context.WithTimeout(context.Background(), time.Second*10)
	}

	// Call cancel to attempt to exit the consumer loop.
	c.cancel()

	// Wait for the consumer to tell is has exited or the context timeout occurs.
	select {
	case <-c.closed:
		log.Info("gracefully shutdown consumer loop", nil)
	case <-ctx.Done():
		log.Info("forced shutdown of consumer loop", nil)
	}
}
