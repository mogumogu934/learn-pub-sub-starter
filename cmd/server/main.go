package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mogumogu934/learn-pub-sub-starter/internal/gamelogic"
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
	defer conn.Close()
	fmt.Println("Peril game server connected to RabbitMQ")

	publishChan, err := conn.Channel()
	if err != nil {
		log.Fatalf("unable to create channel: %v", err)
	}

	_, _, err = pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		"game_logs.*",
		pubsub.SimpleQueueDurable,
	)
	if err != nil {
		log.Fatalf("unable to declare and bind queue '%s' to exchange '%s': %v", routing.GameLogSlug, routing.ExchangePerilTopic, err)
	}
	fmt.Printf("Queue '%s' declared and bound to exchange '%s'\n", routing.GameLogSlug, routing.ExchangePerilTopic)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		fmt.Println("\nReceived interrupt. Shutting down...")
		os.Exit(0)
	}()

	for {
		fmt.Println()
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}

		switch input[0] {
		case "pause":
			fmt.Println("Publishing paused game state")
			err = pubsub.PublishJSON(
				publishChan,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: true,
				},
			)
			if err != nil {
				log.Printf("unable to publish paused game state: %v", err)
			}

		case "resume":
			fmt.Println("Publishing resumes game state")
			err = pubsub.PublishJSON(
				publishChan,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: false,
				},
			)
			if err != nil {
				log.Printf("unable to publish resumes game state: %v", err)
			}

		case "quit":
			fmt.Println("Exiting the game")
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		default:
			fmt.Println("unknown command")
		}
	}
}
