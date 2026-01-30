package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")
	connection_string := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connection_string)
	if err != nil {
		fmt.Printf("Something went wrong: %v\n", err)
		return
	}
	defer conn.Close()
	fmt.Println("Successfully connecting to the RabbitMQ Server")

	channel, err := conn.Channel()
	if err != nil {
		fmt.Printf("Something went wrong: %v\n", err)
		return
	}
	defer channel.Close()
	fmt.Println("Successfully connecting to the channel")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Printf("Something went wrong: %v\n", err)
		return
	}

	if err != nil {
		fmt.Printf("Something went wrong: %v\n", err)
		return
	}

	game_state := gamelogic.NewGameState(username)

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey + "." + username,
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(game_state),
	)
	if err != nil {
		fmt.Printf("Something went wrong: %v", err)
		return
	}

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix + "." + username,
		routing.ArmyMovesPrefix + ".*",
		pubsub.Transient,
		handlerMove(game_state),
	)
	if err != nil {
		fmt.Printf("Something went wrong: %v", err)
		return
	}

	for {
		command := gamelogic.GetInput()
		if len(command) == 0 {
			continue
		}

		switch command[0] {
		case "spawn":
			if err := game_state.CommandSpawn(command); err != nil {
				log.Printf("Failed to spawn unit: %v", err)
				continue
			}
		case "move":
			move, err := game_state.CommandMove(command)
			if err != nil {
				log.Printf("Failed to move unit: %v", err)
				continue
			}
			fmt.Printf("Successfully moved unit to %v\n", move.ToLocation)
			err = pubsub.PublishJSON(
				channel, 
				routing.ExchangePerilTopic, 
				routing.ArmyMovesPrefix + "." + move.Player.Username, 
				move,
			)
			if err != nil {
				log.Printf("Failed to publish move: %v", err)
				continue
			}
			fmt.Printf("Player %v has moved it's army to %v", move.Player.Username, move.ToLocation)

		case "status":
			game_state.CommandStatus()
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
