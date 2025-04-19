package main

import (
	"fmt"
	"time"

	"github.com/mogumogu934/learn-pub-sub-starter/internal/gamelogic"
	"github.com/mogumogu934/learn-pub-sub-starter/internal/pubsub"
	"github.com/mogumogu934/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func HandlerWar(gs *gamelogic.GameState, publishChan *amqp.Channel) func(warDec gamelogic.RecognitionOfWar) pubsub.AckType {

	return func(warDec gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		warOutcome, winner, loser := gs.HandleWar(warDec)
		switch warOutcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue

		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard

		case gamelogic.WarOutcomeOpponentWon:
			err := pubsub.PublishGob(
				publishChan,
				routing.ExchangePerilTopic,
				fmt.Sprintf("%s.%s", routing.GameLogSlug, gs.GetUsername()),
				routing.GameLog{
					CurrentTime: time.Now(),
					Message:     fmt.Sprintf("%s won a war against %s", winner, loser),
					Username:    gs.GetUsername(),
				},
			)
			if err != nil {
				fmt.Printf("unable to publish war outcome: %v", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack

		case gamelogic.WarOutcomeYouWon:
			err := pubsub.PublishGob(
				publishChan,
				routing.ExchangePerilTopic,
				fmt.Sprintf("%s.%s", routing.GameLogSlug, gs.GetUsername()),
				routing.GameLog{
					CurrentTime: time.Now(),
					Message:     fmt.Sprintf("%s won a war against %s", winner, loser),
					Username:    gs.GetUsername(),
				},
			)
			if err != nil {
				fmt.Printf("unable to publish war outcome: %v", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack

		case gamelogic.WarOutcomeDraw:
			err := pubsub.PublishGob(
				publishChan,
				routing.ExchangePerilTopic,
				fmt.Sprintf("%s.%s", routing.GameLogSlug, gs.GetUsername()),
				routing.GameLog{
					CurrentTime: time.Now(),
					Message:     fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser),
					Username:    gs.GetUsername(),
				},
			)
			if err != nil {
				fmt.Printf("unable to publish war outcome: %v", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		}

		fmt.Println("error: unknown war outcome")
		return pubsub.NackDiscard
	}
}
