package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/cryptidcodes/learn-pub-sub/internal/gamelogic"
	"github.com/cryptidcodes/learn-pub-sub/internal/pubsub"
	"github.com/cryptidcodes/learn-pub-sub/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")
	var connectionString string = "amqp://guest:guest@localhost:5672/"

	connection, err := amqp.Dial(connectionString)
	if err != nil {
		fmt.Println("Failed to connect to RabbitMQ:", err)
		return
	}

	// connection to RabbitMQ is now open
	defer connection.Close()
	fmt.Println("Connected to RabbitMQ")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Println("Failed to get username:", err)
		return
	}

	gs := gamelogic.NewGameState(username)

	// declare a queue for pause messages, bound to the pause key
	err = pubsub.SubscribeJSON(connection, routing.ExchangePerilDirect, fmt.Sprintf("%s.%s", routing.PauseKey, username), routing.PauseKey, pubsub.Transient, handlerPause(gs))
	if err != nil {
		fmt.Println("Failed to subscribe to pause messages:", err)
		return
	}

	// declare a queue and channel for move commands from other players
	ch, err := connection.Channel()
	if err != nil {
		fmt.Println("Failed to open a channel:", err)
		return
	}
	err = pubsub.SubscribeJSON(connection, routing.ExchangePerilTopic, fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username), fmt.Sprintf("%s.*", routing.ArmyMovesPrefix), pubsub.Transient, handlerMove(gs, ch))
	if err != nil {
		fmt.Println("Failed to subscribe to move messages:", err)
		return
	}

	// declare a queue for war recognitions from other players
	err = pubsub.SubscribeJSON(connection, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix, fmt.Sprintf("%s.*", routing.WarRecognitionsPrefix), pubsub.Durable, handlerWar(gs))
	if err != nil {
		fmt.Println("Failed to subscribe to war messages:", err)
		return
	}

	// start an infinite REPL
	for i := 0; i == 0; {
		input := gamelogic.GetInput()
		if len(input) > 0 {
			command := input[0]
			switch command {
			case "spawn":
				err = gs.CommandSpawn(input)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Printf("Spawned unit '%s' in location '%s'\n", input[2], input[1])
			case "move":
				move, err := gs.CommandMove(input)
				if err != nil {
					fmt.Println(err)
				}
				// create a channel on the connection
				ch, err := connection.Channel()
				if err != nil {
					fmt.Println("Failed to open a channel:", err)
					return
				}
				defer ch.Close()
				err = pubsub.PublishJSON(ch, routing.ExchangePerilTopic, fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username), move)
				if err != nil {
					fmt.Println("Failed to publish move message:", err)
					return
				}
				fmt.Printf("Moved unit %s to location '%s'\n", input[2], input[1])
			case "status":
				gs.CommandStatus()
			case "help":
				gamelogic.PrintClientHelp()
			case "spam":
				fmt.Println("Spamming not allowed yet!")
			case "quit":
				gamelogic.PrintQuit()
				i = 1
			default:
				fmt.Println("Unknown command. Type 'help' for a list of commands.")
			}
		}
	}

	// wait for ctrl-c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	fmt.Println("\nShutting down Peril client...")
}
