package main

import (
	"fmt"
	"os"
	"os/signal"

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

	signChan := make(chan os.Signal, 1)
	signal.Notify(signChan, os.Interrupt)
	<-signChan
	fmt.Println()
	fmt.Println("RabbitMQ connection close")

}
