package GoLib

import (
	"github.com/streadway/amqp"
)

var conn *amqp.Connection
var channel *amqp.Channel

func Connect() (*amqp.Connection, error) {
	var err error
	conn, err = amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return nil, err
	}

	channel, err = conn.Channel()
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func DeclareQueue(name string) (*amqp.Queue, error) {
	queue, err := channel.QueueDeclare(
		name,  // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, err
	}
	return &queue, nil
}

func Close() {
	channel.Close()
	conn.Close()
}

func Consume(queueName string) (<-chan amqp.Delivery, error) {
	messages, err := channel.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return nil, err
	}
	return messages, nil
}
