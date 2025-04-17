package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	SimpleQueueDurable   = 0
	SimpleQueueTransient = 1
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	data, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("unable to marshal val: %v", err)
	}

	err = ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        data,
		},
	)
	if err != nil {
		return fmt.Errorf("unable to publish message: %v", err)
	}

	return nil
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {

	connChan, err := conn.Channel()
	if err != nil {
		log.Fatalf("unable to open channel: %v", err)
	}

	durable := true
	autoDelete := false
	exclusive := false
	noWait := false

	if simpleQueueType == SimpleQueueTransient {
		durable = false
		autoDelete = true
		exclusive = true
	}

	queue, err := connChan.QueueDeclare(
		queueName,
		durable,
		autoDelete,
		exclusive,
		noWait,
		nil,
	)
	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, fmt.Errorf("unable to declare queue: %v", err)
	}

	err = connChan.QueueBind(
		queueName,
		key,
		exchange,
		noWait,
		nil,
	)
	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, fmt.Errorf("unable to bind queue to exchange: %v", err)
	}

	return connChan, queue, nil
}
