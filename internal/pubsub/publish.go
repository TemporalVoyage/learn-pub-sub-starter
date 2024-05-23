package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"

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

func DeclareAndBind(conn *amqp.Connection, exchange, queueName, key string, simpleQueueType routing.SimpleQueueType) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {

		fmt.Println("3")
		return nil, amqp.Queue{}, err
	}
	config := amqp.Table{
		"x-dead-letter-exchange": "peril_dlx",
	}
	queue, err := ch.QueueDeclare(queueName, routing.SimpleQueueDurable == simpleQueueType, routing.SimpleQueueTransient == simpleQueueType, routing.SimpleQueueTransient == simpleQueueType, false, config)
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

func SubscribeJSON[T any](conn *amqp.Connection, exchange, queueName, key string, simpleQueueType routing.SimpleQueueType, handler func(T) routing.Acktype) error {
	ch, _, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		return err
	}
	del, err := ch.Consume(queueName, "", false, false, false, false, nil)

	if err != nil {
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
			ackType := handler(dat)
			switch ackType {
			case routing.Ack:
				msg.Ack(false)
				log.Println("Ack")
			case routing.NackDis:
				msg.Nack(false, false)
				log.Println("NackDis")
			case routing.NackRe:
				msg.Nack(false, true)
				log.Println("NackRe")
			default:
				log.Println("Unknown")
			}
		}
	}()

	return nil
}

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var arr bytes.Buffer
	enc := gob.NewEncoder(&arr)
	err := enc.Encode(val)
	if err != nil {
		fmt.Println("Gob Error")
		return err
	}
	return ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/gob",
		Body:        arr.Bytes(),
	})
}
