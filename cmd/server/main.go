package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/mogumogu934/learn-pub-sub-starter/internal/pubsub"
	"github.com/mogumogu934/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	const rabbitConnStr = "amqp://guest:guest@localhost:5672/"

	fmt.Println("Starting Peril server...")
	conn, err := amqp.Dial(rabbitConnStr)
	if err != nil {
		log.Fatalf("unable to connect to RabbitMQ: %v", err)
	}
	fmt.Println("Connection to RabbitMQ successful")
	defer conn.Close()

	connChan, err := conn.Channel()
	if err != nil {
		log.Fatalf("unable to open channel: %v", err)
	}

	err = pubsub.PublishJSON(
		connChan,
		routing.ExchangePerilDirect,
		string(routing.PauseKey),
		routing.PlayingState{
			IsPaused: true,
		},
	)
	if err != nil {
		log.Fatalf("unable to publish JSON: %v", err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("Received an interrupt. Shutting down program and closing connection.")
	return
}
