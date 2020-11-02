package rabbitmq

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/streadway/amqp"
)

const (
	defaultExchange = "health_check"
)

var (
	defaultConsumeTimeout = time.Second * 3
)

type (
	// Config is the RabbitMQ checker configuration settings container.
	Config struct {
		// DSN is the RabbitMQ instance connection DSN. Required.
		DSN string
		// Exchange is the application health check exchange. If not set - "health_check" is used.
		Exchange string
		// RoutingKey is the application health check routing key within health check exchange.
		// Can be an application or host name, for example.
		// If not set - host name is used.
		RoutingKey string
		// Queue is the application health check queue, that binds to the exchange with the routing key.
		// If not set - "<exchange>.<routing-key>" is used.
		Queue string
		// ConsumeTimeout is the duration that health check will try to consume published test message.
		// If not set - 3 seconds
		ConsumeTimeout time.Duration
	}
)

// New creates new RabbitMQ health check that verifies the following:
// - connection establishing
// - getting channel from the connection
// - declaring topic exchange
// - declaring queue
// - binding a queue to the exchange with the defined routing key
// - publishing a message to the exchange with the defined routing key
// - consuming published message
func New(config Config) func(ctx context.Context) error {
	(&config).defaults()

	return func(ctx context.Context) (checkErr error) {
		conn, err := amqp.Dial(config.DSN)
		if err != nil {
			checkErr = fmt.Errorf("RabbitMQ health check failed on dial phase: %w", err)
			return
		}
		defer func() {
			// override checkErr only if there were no other errors
			if err := conn.Close(); err != nil && checkErr == nil {
				checkErr = fmt.Errorf("RabbitMQ health check failed to close connection: %w", err)
			}
		}()

		ch, err := conn.Channel()
		if err != nil {
			checkErr = fmt.Errorf("RabbitMQ health check failed on getting channel phase: %w", err)
			return
		}
		defer func() {
			// override checkErr only if there were no other errors
			if err := ch.Close(); err != nil && checkErr == nil {
				checkErr = fmt.Errorf("RabbitMQ health check failed to close channel: %w", err)
			}
		}()

		if err := ch.ExchangeDeclare(config.Exchange, "topic", true, false, false, false, nil); err != nil {
			checkErr = fmt.Errorf("RabbitMQ health check failed during declaring exchange: %w", err)
			return
		}

		if _, err := ch.QueueDeclare(config.Queue, false, false, false, false, nil); err != nil {
			checkErr = fmt.Errorf("RabbitMQ health check failed during declaring queue: %w", err)
			return
		}

		if err := ch.QueueBind(config.Queue, config.RoutingKey, config.Exchange, false, nil); err != nil {
			checkErr = fmt.Errorf("RabbitMQ health check failed during binding: %w", err)
			return
		}

		messages, err := ch.Consume(config.Queue, "", true, false, false, false, nil)
		if err != nil {
			checkErr = fmt.Errorf("RabbitMQ health check failed during consuming: %w", err)
			return
		}

		done := make(chan struct{})

		go func() {
			// block until: a message is received, or message channel is closed (consume timeout)
			<-messages

			// release the channel resources, and unblock the receive on done below
			close(done)

			// now drain any incidental remaining messages
			for range messages {
			}
		}()

		p := amqp.Publishing{Body: []byte(time.Now().Format(time.RFC3339Nano))}
		if err := ch.Publish(config.Exchange, config.RoutingKey, false, false, p); err != nil {
			checkErr = fmt.Errorf("RabbitMQ health check failed during publishing: %w", err)
			return
		}

		for {
			select {
			case <-time.After(config.ConsumeTimeout):
				checkErr = fmt.Errorf("RabbitMQ health check failed due to consume timeout: %w", err)
				return
			case <-done:
				return
			}
		}
	}
}

func (c *Config) defaults() {
	if c.Exchange == "" {
		c.Exchange = defaultExchange
	}

	if c.RoutingKey == "" {
		host, err := os.Hostname()
		if nil != err {
			c.RoutingKey = "-unknown-"
		}
		c.RoutingKey = host
	}

	if c.Queue == "" {
		c.Queue = fmt.Sprintf("%s.%s", c.Exchange, c.RoutingKey)
	}

	if c.ConsumeTimeout == 0 {
		c.ConsumeTimeout = defaultConsumeTimeout
	}
}
