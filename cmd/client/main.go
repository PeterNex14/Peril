package main

import (
	"fmt"
	"os"
	"os/signal"

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

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Printf("Something went wrong: %v\n", err)
		return
	}

	_, _, err = pubsub.DeclareAndBind(
		conn,
		"peril_direct",
		routing.PauseKey + "." + username,
		routing.PauseKey,
		pubsub.Transient,
	)
	if err != nil {
		fmt.Printf("Something went wrong: %v\n", err)
		return
	}

	signChan := make(chan os.Signal, 1)
	signal.Notify(signChan, os.Interrupt)
	<-signChan
	fmt.Println()
	fmt.Println("RabbitMQ Client connection close")


}
