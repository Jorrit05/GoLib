package GoLib

import (
	"context"
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

var conn *amqp.Connection
var channel *amqp.Channel

func SetupConnection(serviceName string, routingKey string) (<-chan amqp.Delivery, *amqp.Connection, *amqp.Channel, error) {
	connectionString, err := GetAMQConnectionString()
	if err != nil {
		return nil, nil, nil, err
	}

	conn, err := Connect(connectionString)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
		return nil, nil, nil, err
	}

	err = Exchange()
	if err != nil {
		log.Fatalf("Failed to create exchange: %v", err)
		return nil, nil, nil, err
	}

	queue, err := DeclareQueue(serviceName)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	if err := channel.QueueBind(queue.Name, routingKey, "topic_exchange", false, nil); err != nil {
		log.Fatalf("Queue Bind: %s", err)
	}

	// Start listening to queue defined by environment var INPUT_QUEUE
	messages, err := Consume(os.Getenv("INPUT_QUEUE"))
	if err != nil {
		log.Fatalf("Failed to register consumer: %v", err)
	}

	return messages, conn, channel, nil
}

func Connect(connectionString string) (*amqp.Connection, error) {
	var err error
	conn, err = amqp.Dial(connectionString)
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

func Exchange() error {
	if err := channel.ExchangeDeclare(
		"topic_exchange",
		"topic",
		true,  // durable
		false, // auto delete
		false, // internal
		false, // no-wait
		nil);  // arguments
	err != nil {
		return err
	}
	return nil
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

func Publish(queueName string, message []byte) error {
	err := channel.PublishWithContext(context.Background(), "", queueName, false, false, amqp.Publishing{
		ContentType: "text/json",
		Body:        message,
	})
	if err != nil {
		return err
	}
	return nil
}
