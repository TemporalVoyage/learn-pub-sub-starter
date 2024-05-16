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
		return nil, amqp.Queue{}, err
	}

	queue, err := ch.QueueDeclare(queueName, routing.Durable == simpleQueueType, routing.Transient == simpleQueueType, routing.Transient == simpleQueueType, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	ch.QueueBind(queueName, routing.PauseKey, exchange, false, nil)
	return ch, queue, nil
}

func SubscribeJSON[T any](conn *amqp.Connection, exchange, queueName, key string, simpleQueueType int, handler func(T)) error {
	ch, _, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		return err
	}
	del, err := ch.Consume(queueName, "", false, false, false, false, nil)

	if err != nil {
		return err
	}

	go func(d <-chan amqp.Delivery) {
		var dat T
		for msg := range d {
			err := json.Unmarshal(msg.Body, &dat)
			fmt.Println(msg.Body)
			if err != nil {
				continue
			}
			handler(dat)
			msg.Ack(false)
		}
	}(del)

	return nil
}
