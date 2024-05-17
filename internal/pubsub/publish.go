package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/temporalvoyage/learn-pub-sub-starter/internal/routing"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	dat, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        dat,
	})
}

func DeclareAndBind(conn *amqp.Connection, exchange, queueName, key string, simpleQueueType int) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {

		fmt.Println("3")
		return nil, amqp.Queue{}, err
	}

	queue, err := ch.QueueDeclare(queueName, routing.Durable == simpleQueueType, routing.Transient == simpleQueueType, routing.Transient == simpleQueueType, false, nil)
	if err != nil {

		fmt.Println("4")
		return nil, amqp.Queue{}, err
	}
	err = ch.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {

		fmt.Println("5")
		return nil, amqp.Queue{}, err
	}
	return ch, queue, nil
}

func SubscribeJSON[T any](conn *amqp.Connection, exchange, queueName, key string, simpleQueueType int, handler func(T)) error {
	ch, _, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		fmt.Println("1")
		return err
	}
	del, err := ch.Consume(queueName, "", false, false, false, false, nil)

	if err != nil {
		fmt.Println("2")
		return err
	}

	go func() {
		defer ch.Close()
		var dat T
		for msg := range del {
			err := json.Unmarshal(msg.Body, &dat)
			fmt.Println(msg.Body)
			if err != nil {
				continue
			}
			handler(dat)
			msg.Ack(false)
		}
	}()

	return nil
}
