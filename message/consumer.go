package message

import (
	"context"
	"github.com/ONSdigital/dp-import-reporter/logging"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
	"time"
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

// Listen poll the kafka topic for incoming messages and process them
func (c *Consumer) Listen() {
	defer func() {
		c.closed <- true
	}()

	go func() {
		for {
			select {
			case eventMsg := <-c.consumer.Incoming():
				logger.Info("incoming received a message", nil)

				if err := c.eventReceiver.ProcessMessage(eventMsg); err != nil {
					log.ErrorC("error returned from eventReceiver.ProcessMessage event message will not be committed to consumer group", err, nil)
					continue
				}
				eventMsg.Commit()
			case <-c.ctx.Done():
				logger.Info("attempting to close down consumer", nil)
				return
			}
		}
	}()
}

// Close attempt to close the consumer Listen loop. If context is nil or no context deadline is set then the default is
// applied. Close will return when the Listen loop notifies it has exited or the timeout limit is reached.
func (c Consumer) Close(ctx context.Context) {
	if ctx == nil {
		var cancelFunc context.CancelFunc
		ctx, cancelFunc = context.WithTimeout(context.Background(), c.timeout)
		defer cancelFunc()
	}
	if _, ok := ctx.Deadline(); !ok {
		var cancelFunc context.CancelFunc
		ctx, cancelFunc = context.WithTimeout(context.Background(), c.timeout)
		defer cancelFunc()
	}

	// Call cancel to attempt to exit the consumer loop.
	c.cancel()

	// Wait for the consumer to tell is has exited or the context timeout occurs.
	select {
	case <-c.closed:
		logger.Info("gracefully shutdown consumer loop", nil)
	case <-ctx.Done():
		logger.Info("forced shutdown of consumer loop", nil)
	}
}
