package pubsub

import (
	"bytes"
	"encoding/gob"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeGob[T any](
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

	delivery, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	
	go func ()  {
		for deliv := range delivery {
			var data T
			buff := bytes.NewBuffer(deliv.Body)
			decoder := gob.NewDecoder(buff)
			err := decoder.Decode(&data)
			if err != nil {
				log.Printf("Something went wrong when decoding: %v", err)
				return
			}

			ack_type := handler(data)
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