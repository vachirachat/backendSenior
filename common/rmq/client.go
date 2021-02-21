package rmq

import (
	"fmt"
	"github.com/streadway/amqp"
)

type RMQClient struct {
	url     string // connection URI
	conn    *amqp.Connection
	channel *amqp.Channel
}

func New(url string) *RMQClient {
	return &RMQClient{
		url: url,
	}
}

func (c *RMQClient) Connect() error {
	conn, err := amqp.Dial(c.url)
	if err != nil {
		return fmt.Errorf("error connecting: %w", err)
	}
	c.conn = conn

	channel, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("error connecting: %w", err)
	}
	c.channel = channel
	return nil
}

func (c *RMQClient) EnsureQueue(queue string) error {
	_, err := c.channel.QueueDeclare(
		queue, // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("ensuring queue: %w", err)
	}
	return nil
}

// shortcut for subscribing to queue
func (c *RMQClient) Consume(queue string) (<-chan amqp.Delivery, error) {
	msgs, err := c.channel.Consume(queue, "", false, false, false, false, nil)
	return msgs, err
}

func (c *RMQClient) Publish(queue string, message []byte) error {
	err := c.channel.Publish(
		"", queue, false, false, amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         message,
		})
	return err
}
