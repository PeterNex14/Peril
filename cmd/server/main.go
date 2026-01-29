package main

import (
	"fmt"
	"os"
	"os/signal"

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

	err = pubsub.PublishJSON(
		channel, 
		routing.ExchangePerilDirect, 
		routing.PauseKey,
		routing.PlayingState{
			IsPaused: true,
		},
	)
	if err != nil {
		fmt.Printf("Something went wrong: %v\n", err)
		return
	}

	signChan := make(chan os.Signal, 1)
	signal.Notify(signChan, os.Interrupt)
	<-signChan
	fmt.Println()
	fmt.Println("RabbitMQ Server connection close")

}
