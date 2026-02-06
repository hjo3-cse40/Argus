package mq

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func Connect(amqpURL string) (*Client, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("mq dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("mq channel: %w", err)
	}

	log.Println("connected to RabbitMQ")

	return &Client{conn: conn, ch: ch}, nil
}

func (c *Client) DeclareQueue(name string) error {
	_, err := c.ch.QueueDeclare(
		name,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("declare queue %s: %w", name, err)
	}
	return nil
}

func (c *Client) Publish(queue string, body []byte) error {
	return c.ch.Publish(
		"",    // default exchange
		queue, // routing key = queue name
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (c *Client) Close() {
	if c.ch != nil {
		_ = c.ch.Close()
	}
	if c.conn != nil {
		_ = c.conn.Close()
	}
}
