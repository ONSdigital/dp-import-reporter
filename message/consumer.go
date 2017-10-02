package message

import (
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
	"context"
	"time"
)

//go:generate moq -out ../mocks/message_generated_mocks.go -pkg mocks . KafkaConsumer KafkaMessage Receiver

// MessageConsumer consumes incoming reportEvent Messages from the event-reporter kafka topic
type MessageConsumer struct {
	closed        chan bool
	ctx           context.Context
	cancel        context.CancelFunc
	consumer      KafkaConsumer
	eventReceiver Receiver
}

// NewMessageConsumer Create a new MessageConsumer
func NewMessageConsumer(consumer KafkaConsumer, eventHandler Receiver) MessageConsumer {
	ctx, cancel := context.WithCancel(context.Background())

	return MessageConsumer{
		closed:        make(chan bool, 1),
		ctx:           ctx,
		cancel:        cancel,
		consumer:      consumer,
		eventReceiver: eventHandler,
	}
}

// Listen poll the kafka topic for incoming messages
func (c *MessageConsumer) Listen() {
	defer func() {
		c.closed <- true
	}()

	go func() {
		for {
			select {
			case eventMsg := <-c.consumer.Incoming():
				if err := c.eventReceiver.ProcessMessage(eventMsg); err != nil {
					log.ErrorC("unexpected error returned from handleEvent, kafka offset will not be updated", err, nil)
					//eventMsg.Commit()
					continue
				}
				eventMsg.Commit()
			case <-c.ctx.Done():
				log.Info("attempting to close down consumer", nil)
				return
			}
		}
	}()
}

// Close shutdown Consumer.Listen loop.
func (c MessageConsumer) Close(ctx context.Context) {
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

// Receiver defines a struct that processes a kafka message
type Receiver interface {
	ProcessMessage(event kafka.Message) error
}

type KafkaMessage kafka.Message

type KafkaConsumer interface {
	Incoming() chan kafka.Message
}
