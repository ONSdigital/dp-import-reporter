package message

import (
	"context"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
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

// Listen poll the kafka topic for incoming messages and process them
func (c *MessageConsumer) Listen() {
	defer func() {
		c.closed <- true
	}()

	go func() {
		for {
			select {
			case eventMsg := <-c.consumer.Incoming():
				if err := c.eventReceiver.ProcessMessage(eventMsg); err != nil {
					log.ErrorC("unexpected error returned from eventReceiver.ProcessMessage, event message will not be committed to consumer group", err, nil)
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

// Close attempt to close the consumer Listen loop. If context is nil or no context deadline is set then the default is
// applied. Close will return when the Listen loop notifies it has exited or the timeout limit is reached.
func (c MessageConsumer) Close(ctx context.Context) {
	if ctx == nil {
		ctx, _ = context.WithTimeout(context.Background(), time.Second*10)
	}
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
