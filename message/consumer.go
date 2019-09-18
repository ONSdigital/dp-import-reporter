package message

import (
	"context"
	"time"

	"github.com/ONSdigital/dp-import-reporter/logging"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
)

//go:generate moq -out ../mocks/message_generated_mocks.go -pkg mocks . KafkaConsumer KafkaMessage Receiver

var logger = logging.Logger{"message.Consumer"}

// Receiver defines a struct that processes a kafka message
type Receiver interface {
	ProcessMessage(event kafka.Message) error
}

type KafkaMessage kafka.Message

type KafkaConsumer interface {
	Incoming() chan kafka.Message
	CommitAndRelease(kafka.Message)
	StopListeningToConsumer(context.Context) error
	Close(context.Context) error
	Errors() chan error
}

// Consumer consumes incoming reportEvent Messages from the event-reporter kafka topic
type Consumer struct {
	closed        chan bool
	ctx           context.Context
	cancel        context.CancelFunc
	consumer      KafkaConsumer
	eventReceiver Receiver
	timeout       time.Duration
}

// NewConsumer Create a new Consumer
func NewConsumer(consumer KafkaConsumer, eventReceiver Receiver, timeout time.Duration) Consumer {
	ctx, cancel := context.WithCancel(context.Background())

	return Consumer{
		closed:        make(chan bool, 1),
		ctx:           ctx,
		cancel:        cancel,
		consumer:      consumer,
		eventReceiver: eventReceiver,
		timeout:       timeout,
	}
}

// Listen polls the kafka topic for incoming messages and dispatches them for processing
func (c *Consumer) Listen() {
	go func() {
		for keepListening := true; keepListening; {
			select {
			case eventMsg := <-c.consumer.Incoming():
				logger.Info("incoming received a message", log.Data{"msg": eventMsg})

				if err := c.eventReceiver.ProcessMessage(eventMsg); err != nil {
					log.ErrorC("error returned from eventReceiver.ProcessMessage event message will not be committed to consumer group", err, nil)
					continue
				}
				c.consumer.CommitAndRelease(eventMsg)
			case <-c.ctx.Done():
				logger.Info("context done, consumer.Listen loop closing", nil)
				keepListening = false
			}
		}
		close(c.closed)
	}()
}

// Close attempts to close (cancel) the consumer Listen loop.
// If ctx is nil or no context deadline is set then the default is applied.
// Close returns when the Listen loop notifies it has exited or the timeout limit is reached.
func (c Consumer) Close(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	if _, ok := ctx.Deadline(); !ok {
		var enforcedCancel context.CancelFunc
		ctx, enforcedCancel = context.WithTimeout(ctx, c.timeout)
		defer enforcedCancel()
	}

	// stops the kafka listener sending to our listener
	c.consumer.StopListeningToConsumer(ctx)

	// cancel triggers exit of the consumer goroutine in Listen()
	c.cancel()

	// Close finalises the consumer exit
	c.consumer.Close(ctx)

	// wait for the consumer to signal that it has exited or the context timeout occurs
	select {
	case <-c.closed:
		logger.Info("gracefully shutdown consumer loop", nil)
	case <-ctx.Done():
		logger.Info("forced shutdown of consumer loop", nil)
	}
}
