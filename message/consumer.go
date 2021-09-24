package message

import (
	"context"
	"time"

	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/log.go/v2/log"
)

//go:generate moq -out ./message_generated_mocks_test.go . Receiver

// Receiver defines a struct that processes a kafka message
type Receiver interface {
	ProcessMessage(ctx context.Context, event kafka.Message) error
}

// Consumer consumes incoming reportEvent Messages from the event-reporter kafka topic
type Consumer struct {
	closed        chan bool
	ctx           context.Context
	cancel        context.CancelFunc
	consumer      kafka.IConsumerGroup
	eventReceiver Receiver
	timeout       time.Duration
}

// NewConsumer Create a new Consumer
func NewConsumer(consumer kafka.IConsumerGroup, eventReceiver Receiver, timeout time.Duration) Consumer {
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
func (c *Consumer) Listen(ctx context.Context) {
	go func() {
		for keepListening := true; keepListening; {
			select {
			case eventMsg, ok := <-c.consumer.Channels().Upstream:
				if !ok {
					break
				}
				// log.Event(ctx, "incoming received a message", log.INFO, log.Data{"msg": eventMsg})
				log.Info(ctx, "incoming received a message")

				if err := c.eventReceiver.ProcessMessage(ctx, eventMsg); err != nil {
					log.Error(ctx, "error returned from eventReceiver.ProcessMessage ", err)
				}
				eventMsg.CommitAndRelease()
			case <-c.ctx.Done():
				log.Info(ctx, "context done, consumer.Listen loop closing")
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
		log.Info(ctx, "gracefully shutdown consumer loop")
	case <-ctx.Done():
		log.Info(ctx, "forced shutdown of consumer loop")
	}
}
