package main

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/temporalvoyage/learn-pub-sub-starter/internal/gamelogic"
	"github.com/temporalvoyage/learn-pub-sub-starter/internal/pubsub"
	"github.com/temporalvoyage/learn-pub-sub-starter/internal/routing"
)

func main() {
	fmt.Println("Starting Peril client...")
	const rabbitConnString = "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(rabbitConnString)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	fmt.Println("Peril game client connected to RabbitMQ!")

	publishCh, err := conn.Channel()
	if err != nil {
		log.Printf("Error creating channel: %v", err)
	}

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("welcome failed: %v", err)
	}

	gameState := gamelogic.NewGameState(username)

	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, fmt.Sprintf("%v.%v", routing.PauseKey, username), routing.PauseKey, routing.SimpleQueueTransient, handlerPause(gameState))
	if err != nil {
		log.Fatalf("Couldn't subscribe to Pause: %v", err)
	}

	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, fmt.Sprintf("%v.%v", routing.ArmyMovesPrefix, username), fmt.Sprintf("%v.*", routing.ArmyMovesPrefix), routing.SimpleQueueTransient, handlerMove(gameState, publishCh))
	if err != nil {
		log.Fatalf("Couldn't subscribe to Move: %v", err)
	}
	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix, fmt.Sprintf("%v.*", routing.WarRecognitionsPrefix), routing.SimpleQueueDurable, handlerWar(gameState))
	if err != nil {
		log.Fatalf("Couldn't subscribe to War: %v", err)
	}

	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}
		switch input[0] {
		case "spawn":
			err := gameState.CommandSpawn(input)
			if err != nil {
				log.Printf("Error spawning unit: %v", err)
			}
		case "move":
			mv, err := gameState.CommandMove(input)
			if err != nil {
				log.Printf("Error moving unit: %v", err)
			}
			err = pubsub.PublishJSON(publishCh, routing.ExchangePerilTopic, fmt.Sprintf("%v.%v", routing.ArmyMovesPrefix, username), mv)
			if err != nil {
				log.Printf("Error publishing move: %v", err)
			}
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("Unknown command")
		}
	}
}

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

func handlerWar(gs *gamelogic.GameState) func(dw gamelogic.RecognitionOfWar) routing.Acktype {
	return func(dw gamelogic.RecognitionOfWar) routing.Acktype {
		defer fmt.Print("> ")
		warOutcome, _, _ := gs.HandleWar(dw)
		switch warOutcome {
		case gamelogic.WarOutcomeNotInvolved:
			return routing.NackRe
		case gamelogic.WarOutcomeNoUnits:
			return routing.NackDis
		case gamelogic.WarOutcomeOpponentWon:
			return routing.Ack
		case gamelogic.WarOutcomeYouWon:
			return routing.Ack
		case gamelogic.WarOutcomeDraw:
			return routing.Ack
		}

		fmt.Println("Error: Unknown war")
		return routing.NackDis
	}
}
