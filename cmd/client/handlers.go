package main

import (
	"fmt"

	"github.com/mogumogu934/learn-pub-sub-starter/internal/gamelogic"
	"github.com/mogumogu934/learn-pub-sub-starter/internal/pubsub"
	"github.com/mogumogu934/learn-pub-sub-starter/internal/routing"
)

func HandlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(playingState routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(playingState)
		return pubsub.Ack
	}
}

func HandlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		switch gs.HandleMove(move) {
		case gamelogic.MoveOutComeSafe:
			fmt.Printf("ack\n")
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			fmt.Printf("ack\n")
			return pubsub.Ack
		case gamelogic.MoveOutcomeSamePlayer:
			fmt.Printf("nack, discard\n")
			return pubsub.NackDiscard
		default:
			fmt.Printf("nack, discard\n")
			return pubsub.NackDiscard
		}
	}
}
