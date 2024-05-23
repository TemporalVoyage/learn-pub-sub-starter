package main

import (
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/temporalvoyage/learn-pub-sub-starter/internal/gamelogic"
	"github.com/temporalvoyage/learn-pub-sub-starter/internal/pubsub"
	"github.com/temporalvoyage/learn-pub-sub-starter/internal/routing"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) routing.Acktype {
	return func(ps routing.PlayingState) routing.Acktype {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return routing.Ack
	}
}

func handlerMove(gs *gamelogic.GameState, publishCh *amqp.Channel) func(gamelogic.ArmyMove) routing.Acktype {
	return func(mv gamelogic.ArmyMove) routing.Acktype {
		defer fmt.Print("> ")
		switch gs.HandleMove(mv) {
		case gamelogic.MoveOutcomeSamePlayer:
			return routing.Ack
		case gamelogic.MoveOutComeSafe:
			return routing.Ack
		case gamelogic.MoveOutcomeMakeWar:
			err := pubsub.PublishJSON(publishCh, routing.ExchangePerilTopic, fmt.Sprintf("%v.%v", routing.WarRecognitionsPrefix, gs.GetUsername()), gamelogic.RecognitionOfWar{
				Attacker: mv.Player,
				Defender: gs.GetPlayerSnap(),
			})
			if err != nil {
				fmt.Printf("error: %s\n", err)
				return routing.NackRe
			}
			return routing.Ack
		}
		return routing.NackDis
	}
}

func handlerWar(gs *gamelogic.GameState, publishCh *amqp.Channel) func(dw gamelogic.RecognitionOfWar) routing.Acktype {
	return func(dw gamelogic.RecognitionOfWar) routing.Acktype {
		defer fmt.Print("> ")
		warOutcome, winner, loser := gs.HandleWar(dw)
		gameLog := routing.GameLog{
			Username:    gs.GetUsername(),
			CurrentTime: time.Now(),
		}
		switch warOutcome {
		case gamelogic.WarOutcomeNotInvolved:
			return routing.NackRe
		case gamelogic.WarOutcomeNoUnits:
			return routing.NackDis
		case gamelogic.WarOutcomeOpponentWon:
			gameLog.Message = fmt.Sprintf("%v won a war against %v", winner, loser)
			err := pubsub.PublishGob(publishCh, routing.ExchangePerilTopic, fmt.Sprintf("%v.%v", routing.GameLogSlug, gs.GetUsername()), gameLog)
			if err != nil {
				return routing.NackRe
			}
			return routing.Ack
		case gamelogic.WarOutcomeYouWon:
			gameLog.Message = fmt.Sprintf("%v won a war against %v", winner, loser)
			err := pubsub.PublishGob(publishCh, routing.ExchangePerilTopic, fmt.Sprintf("%v.%v", routing.GameLogSlug, gs.GetUsername()), gameLog)
			if err != nil {
				return routing.NackRe
			}
			return routing.Ack
		case gamelogic.WarOutcomeDraw:
			gameLog.Message = fmt.Sprintf("A war between %v and %v resulted in a draw", winner, loser)
			err := pubsub.PublishGob(publishCh, routing.ExchangePerilTopic, fmt.Sprintf("%v.%v", routing.GameLogSlug, gs.GetUsername()), gameLog)
			if err != nil {
				return routing.NackRe
			}
			return routing.Ack
		}

		fmt.Println("Error: Unknown war")
		return routing.NackDis
	}
}
