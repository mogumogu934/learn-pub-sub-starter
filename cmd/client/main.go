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
		log.Fatalf("could not get username: %v", err)
	}

	gameState := gamelogic.NewGameState(username)

	pauseQueue := fmt.Sprintf("%s.%s", routing.PauseKey, username)

	_, _, err = pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		pauseQueue,
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
	)
	if err != nil {
		log.Fatalf("unable to declare and bind queue '%s' to exchange '%s': %v", pauseQueue, routing.ExchangePerilDirect, err)
	}
	fmt.Printf("Queue '%s' declared and bound to exchange '%s'\n", pauseQueue, routing.ExchangePerilDirect)

	moveQueue := fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username)
	moveKey := fmt.Sprintf("%s.*", routing.ArmyMovesPrefix)

	moveChan, _, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilTopic,
		moveQueue,
		moveKey,
		pubsub.SimpleQueueTransient,
	)
	if err != nil {
		log.Fatalf("unable to declare and bind queue '%s' to exchange '%s': %v", moveQueue, routing.ExchangePerilTopic, err)
	}
	fmt.Printf("Queue '%s' declared and bound to exchange '%s'\n", moveQueue, routing.ExchangePerilTopic)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		os.Exit(0)
	}()

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		pauseQueue,
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
		HandlerPause(gameState),
	)
	if err != nil {
		log.Fatalf("unable to subscribe to routing key '%s': %v", routing.PauseKey, err)
	}

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		moveQueue,
		moveKey,
		pubsub.SimpleQueueTransient,
		HandlerMove(gameState),
	)
	if err != nil {
		log.Fatalf("unable to subscribe to routing key '%s': %v", moveKey, err)
	}

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
			move, err := gameState.CommandMove(input)
			if err != nil {
				fmt.Println(err)
				continue
			}

			err = pubsub.PublishJSON(
				moveChan,
				routing.ExchangePerilTopic,
				moveQueue,
				move,
			)
			if err != nil {
				log.Printf("unable to publish move: %v", err)
				continue
			}
			fmt.Printf("Moved unit to %s\n", move.ToLocation)

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
		}
	}
}
