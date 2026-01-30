package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")
	connection_string := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connection_string)
	if err != nil {
		fmt.Println("Something wrong when connecting with RabbitMQ server")
		return
	}
	defer conn.Close()
	fmt.Println("Successfuly connecting to RabbitMQ server")

	channel, err := conn.Channel()
	if err != nil {
		fmt.Println("Something went wrong: creating a channel")
		return
	}

	_, _, err = pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug + ".*",
		pubsub.Durable,
	)
	if err != nil {
		fmt.Printf("Something went wrong: %v\n", err)
		return
	}

	gamelogic.PrintServerHelp()

	for {
		command := gamelogic.GetInput()
		if len(command) == 0 {
			continue
		}

		switch command[0] {
		case "pause":
			fmt.Println("Game is paused")
			err = pubsub.PublishJSON(
				channel, 
				routing.ExchangePerilDirect, 
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: true,
				},
			)
			if err != nil {
				fmt.Printf("Something went wrong when pausing the game: %v\n", err)
				return
			}
		case "resume":
			fmt.Println("Game resumed")
			err = pubsub.PublishJSON(
				channel, 
				routing.ExchangePerilDirect, 
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: false,
				},
			)
			if err != nil {
				fmt.Printf("Something went wrong when pausing the game: %v\n", err)
				return
			}
		case "quit":
			fmt.Println("Exitting the game")
			return

		default:
			fmt.Println("Usage : ")
			gamelogic.PrintServerHelp()
			continue
		}
	

	}

}
