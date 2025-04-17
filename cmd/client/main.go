package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/mogumogu934/learn-pub-sub-starter/internal/gamelogic"
	"github.com/mogumogu934/learn-pub-sub-starter/internal/pubsub"
	"github.com/mogumogu934/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")
	const rabbitConnStr = "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(rabbitConnStr)
	if err != nil {
		log.Fatalf("unable to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	fmt.Println("Peril game client connected to RabbitMQ")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Println(err)
	}

	queueName := fmt.Sprintf("%s.%s", routing.PauseKey, username)

	_, _, err = pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		queueName,
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
	)
	if err != nil {
		log.Println("unable to declare and bind queue to exchange", err)
	}
	fmt.Printf("Queue %v declared and bound\n", queueName)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		os.Exit(0)
	}()

	gameState := gamelogic.NewGameState(username)

	for {
		fmt.Println()
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}

		switch input[0] {
		case "spawn":
			err = gameState.CommandSpawn(input)
			if err != nil {
				fmt.Println(err)
			}
		case "move":
			_, err := gameState.CommandMove(input)
			if err != nil {
				fmt.Println(err)
			}
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			os.Exit(0)
		default:
			fmt.Println("unknown command")
			continue
		}
	}
}
