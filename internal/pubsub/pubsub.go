package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mogumogu934/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	SimpleQueueDurable   = 0
	SimpleQueueTransient = 1
)

type AckType string

const (
	Ack         AckType = "ack"
	NackRequeue AckType = "nackrequeue"
	NackDiscard AckType = "nackdiscard"
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
		amqp.Table{
			"x-dead-letter-exchange": routing.ExchangePerilDeadLetter,
		},
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

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int, // an enum to represent "durable" or "transient"
	handler func(T) (ack AckType),
) error {

	subscribeChan, _, err := DeclareAndBind(
		conn,
		exchange,
		queueName,
		key,
		simpleQueueType,
	)
	if err != nil {
		return fmt.Errorf("unable to declare and bind queue to exchange: %v", err)
	}

	deliveryChan, err := subscribeChan.Consume("", "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("unable to get delivery channel: %v", err)
	}

	go func() {
		for d := range deliveryChan {
			var target T
			if err = json.Unmarshal(d.Body, &target); err != nil {
				log.Printf("unable to unmarshal %v:\n%v", d.Body, err)
			}

			switch handler(target) {
			case Ack:
				if err = d.Ack(false); err != nil {
					log.Printf("could not acknowledge delivery: %v", err)
				}
			case NackRequeue:
				if err = d.Nack(false, true); err != nil { // requeue=true
					log.Printf("could not negative acknowledge delivery: %v", err)
				}
			case NackDiscard:
				if err = d.Nack(false, false); err != nil { // requeue=false
					log.Printf("could not negative acknowledge delivery: %v", err)
				}
			default:
				log.Printf("unknown ack type: %v", handler(target))
			}
		}
	}()

	return nil
}
