package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Acktype int
type SimpleQueueType int

const (
	Durable SimpleQueueType = iota
	Transient
)

const (
	Ack Acktype = iota
	NackDiscard
	NackRequeue
)

// PublishJSON publishes a JSON-encoded value to the given exchange with the given routing key.
func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	// marshal the val to JSON bytes
	bytes, err := json.Marshal(val)
	if err != nil {
		return err
	}
	// use the channel's .PublishWithContext method with the routhing key
	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        bytes,
	})
	if err != nil {
		return err
	}
	return nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) Acktype,
) error {
	// Call DeclareAndBind to make sure that the given queue exists
	// and is bound to the exchange
	ch, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	// get a new chan of amqp.Delivery structs by using the channel.Consume method
	msgs, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	// start a new goroutine that ranges over the channel of deliveries
	go func() {
		defer ch.Close()
		for msg := range msgs {
			var val T
			// unmarshal the body into generic T type
			err = json.Unmarshal(msg.Body, &val)
			if err != nil {
				fmt.Println("Failed to unmarshal message:", err)
				continue
			}
			// call the handler with the unmarshaled value
			ackType := handler(val)
			switch ackType {
			case Ack:
				fmt.Println("Ack: Processed Successfully")
				msg.Ack(false)
			case NackRequeue:
				fmt.Println("NackRequeue: Not processed, requeue")
				msg.Nack(false, true)
			case NackDiscard:
				fmt.Println("NackDiscard: Not processed, discard")
				msg.Nack(false, false)
			default:
				fmt.Println("Err: Unknown ack type")
			}
		}
	}()
	return nil
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	// create a new channel on the connection
	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	// the channel is created, now declare a new queue
	queue, err := ch.QueueDeclare(
		queueName,              // name
		queueType == Durable,   // durable
		queueType == Transient, // delete when unused
		queueType == Transient, // exclusive
		false,                  // no-wait
		amqp.Table{"x-dead-letter-exchange": "peril_dlx"}, // args
	)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	err = ch.QueueBind(
		queueName,
		key,
		exchange,
		false,
		nil,
	)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	return ch, queue, nil
}
