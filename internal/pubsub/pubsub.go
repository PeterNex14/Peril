package pubsub

import (
	"context"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	Durable SimpleQueueType = iota
	Transient
)

var queueTypeEnum = map[SimpleQueueType]string {
	Durable: "durable",
	Transient: "transient",
}

type Acktype int

const (
	Ack Acktype = iota
	NackRequeue
	NackDiscard
)



func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	json_byte, err := json.Marshal(val)
	if err != nil {
		return err
	}

	err = ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body: json_byte,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
) (*amqp.Channel, amqp.Queue, error) {
	channel, err := conn.Channel()
	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, err
	}

	queue, err := channel.QueueDeclare(
		queueName,
		queueType == Durable,
		queueType == Transient,
		queueType == Transient,
		false,
		amqp.Table{
			"x-dead-letter-exchange": "peril_dlx",
		},
	)
	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, err
	}

	err = channel.QueueBind(
		queueName,
		key,
		exchange,
		false,
		nil,
	)
	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, err
	}

	return channel, queue, nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) Acktype,
) error {
	channel, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	err = channel.Qos(10, 0, false)
	if err != nil {
		return err
	}

	delivery, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	
	go func ()  {
		for deliv := range delivery {
			var temp T
			err := json.Unmarshal(deliv.Body, &temp)
			if err != nil {
				log.Printf("Something went wrong: %v", err)
				continue
			} 

			ack_type := handler(temp)
			switch ack_type {
			case Ack:
				deliv.Ack(false)
				log.Printf("Acknowledge Occured")
			case NackRequeue:
				deliv.Nack(false, true)
				log.Printf("Negative Acknowledge and Requeue occured")
			case NackDiscard:
				deliv.Nack(false, false)
				log.Printf("Negative Acknowledge and Discard occured")
			}
		}
	}()

	return nil
}