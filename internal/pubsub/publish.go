package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
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

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	// declare a byte buffer to encode to
	var buf bytes.Buffer
	// create a new encoder on the buffer
	enc := gob.NewEncoder(&buf)
	// encode the gamelog struct
	err := enc.Encode(val)
	// handle errors
	if err != nil {
		return err
	}
	// use the channel's .PublishWithContext method with the routhing key
	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/gob",
		Body:        buf.Bytes(),
	})
	if err != nil {
		return err
	}
	return nil
}
