package main

import (
	"fmt"
	"log"

	"github.com/mogumogu934/learn-pub-sub-starter/internal/gamelogic"
	"github.com/mogumogu934/learn-pub-sub-starter/internal/pubsub"
	"github.com/mogumogu934/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	const rabbitConnStr = "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(rabbitConnStr)
	if err != nil {
		log.Fatalf("unable to connect to RabbitMQ: %v", err)
	}
	fmt.Println("Connection to RabbitMQ successful")
	defer conn.Close()

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Println(err)
	}

	_, _, err = pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%s.%s", routing.PauseKey, username),
		routing.PauseKey,
		pubsub.QueueTypeTransient,
	)
	if err != nil {
		log.Println("unable to declare and bind queue to exchange", err)
	}

	for {
		continue
	}
}
