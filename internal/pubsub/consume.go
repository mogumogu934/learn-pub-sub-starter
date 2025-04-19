package pubsub

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mogumogu934/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType string

const (
	Ack         AckType = "ack"
	NackRequeue AckType = "nackrequeue"
	NackDiscard AckType = "nackdiscard"
)

const (
	SimpleQueueDurable   = 0
	SimpleQueueTransient = 1
)

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int,
	handler func(T) AckType,
	unmarshaller func([]byte) (T, error),
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
			value, err := unmarshaller(d.Body)
			if err != nil {
				log.Printf("unable to decode message:\n%v", err)
				continue
			}

			ackType := handler(value)
			switch ackType {
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
				log.Printf("unknown ack type: %v", ackType)
			}
		}
	}()

	return nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int,
	handler func(T) AckType,
) error {

	unmarshaller := func(data []byte) (T, error) {
		var value T
		err := json.Unmarshal(data, &value)
		return value, err
	}

	return subscribe(conn, exchange, queueName, key, simpleQueueType, handler, unmarshaller)
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int,
	handler func(T) AckType,
) error {

	unmarshaller := func(data []byte) (T, error) {
		var value T
		buffer := bytes.NewBuffer(data)
		dec := gob.NewDecoder(buffer)
		err := dec.Decode(&value)
		return value, err
	}

	return subscribe(conn, exchange, queueName, key, simpleQueueType, handler, unmarshaller)
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int,
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
