package main

import (
	"log"
	"os"
	"time"

	"github.com/streadway/amqp"
)

const (
	queue                = "drivers-queue"
	consumerTag          = "drivers-consumer"
	durableQueue         = false
	deleteQueueIfNotUsed = true
	lifetime             = 5 * time.Second //lifetime of process before shutdown (0s=infinite)
)

// Consumer ...
type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	tag     string
	done    chan error
}

// NewConsumer create and initialize a new consumer with defaultURI if AMQP_URL
// environment variable is not defined.
func NewConsumer(defaultURI string) (*Consumer, error) {
	c := &Consumer{
		conn:    nil,
		channel: nil,
		done:    make(chan error),
	}

	uri := defaultURI
	env := os.Getenv("AMQP_URL")
	if env != "" {
		uri = env
	}

	var err error
	c.conn, err = amqp.Dial(uri)
	if err != nil {
		return nil, err
	}
	log.Printf("Connect to AMPQ: %s", uri)

	c.channel, err = c.conn.Channel()
	if err != nil {
		return nil, err
	}

	return c, nil
}

// Bind add a queue from `exchange` with a `routingKey`.
func (c *Consumer) Bind(exchange, routingKey string) error {
	// (name, durable, delete when unused, exclusive, no wait, arguments)
	_, err := c.channel.QueueDeclare(queue, durableQueue, deleteQueueIfNotUsed,
		false, false, nil)
	if err != nil {
		return err
	}

	// (name, routing key, source exchange, no wait, arguments)
	err = c.channel.QueueBind(queue, routingKey, exchange, false, nil)
	if err != nil {
		return err
	}
	log.Printf("Bind  [ %s ] ---( %s )---> [ %s ]", exchange, routingKey, queue)
	return nil
}

// Stop and clean the consumer.
func (c *Consumer) Stop() error {
	if err := c.channel.Cancel(consumerTag, true); err != nil {
		return err
	}

	if err := c.channel.Close(); err != nil {
		return err
	}

	if err := c.conn.Close(); err != nil {
		return err
	}

	defer log.Printf("Consumer stopped!")
	return <-c.done
}

// HandleFunc ...
type HandleFunc func(messages <-chan amqp.Delivery, done chan error)

// StartConsumer start an asynchronous consumer.
func StartConsumer(defaultURI, exchange, routingKey string, handle HandleFunc) error {
	c, err := NewConsumer(defaultURI)
	if err != nil {
		return err
	}

	if err := c.Bind(exchange, routingKey); err != nil {
		return err
	}

	// (queue, consumer, autoAck, exclusive, noLocal, noWait, args)
	msgs, err := c.channel.Consume(queue, consumerTag, true, false, false, false, nil)
	if err != nil {
		return err
	}

	forever := make(chan bool)
	go handle(msgs, c.done)
	<-forever

	return c.Stop()
}
