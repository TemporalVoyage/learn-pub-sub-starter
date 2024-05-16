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

	username, err := gamelogic.ClientWelcome()

	if err != nil {
		log.Fatalf("welcome failed: %v", err)
	}
	queName := fmt.Sprintf("%v.%v", routing.PauseKey, username)
	_, queue, err := pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, queName, routing.PauseKey, routing.Transient)

	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}
	fmt.Printf("Queue %v declared and bound!\n", queue.Name)

	gameState := gamelogic.NewGameState(username)

	pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, queName, routing.PauseKey, routing.Transient, handlerPause(gameState))

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
			_, err := gameState.CommandMove(input)
			if err != nil {
				log.Printf("Erro moving unit: %v", err)
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

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(ps routing.PlayingState) {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
	}
}
