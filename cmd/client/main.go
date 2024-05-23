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
	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix, fmt.Sprintf("%v.*", routing.WarRecognitionsPrefix), routing.SimpleQueueDurable, handlerWar(gameState, publishCh))
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
