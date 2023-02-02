package GoLib

import (
	"context"
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

var conn *amqp.Connection
var channel *amqp.Channel

type serviceFunc func(message amqp.Delivery) (amqp.Publishing, error)

// SetupConnection establishes a connection to RabbitMQ and sets up a topic exchange, queue, and consumer to listen to
// messages with the specified routing key.
// It returns a channel to receive delivery messages, the AMQP connection, and channel objects, and an error if any occurs during the setup.
// The connection string and the routing key are passed as arguments.
// The service name is used to declare the queue.
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
		return nil, nil, nil, err
	}

	// Bind queue to "topic_exchange"
	// TODO: Make "topic_exchange" flexible?
	if err := channel.QueueBind(queue.Name, routingKey, "topic_exchange", false, nil); err != nil {
		log.Fatalf("Queue Bind: %s", err)
		return nil, nil, nil, err
	}

	// Start listening to queue defined by environment var INPUT_QUEUE
	messages, err := Consume(os.Getenv("INPUT_QUEUE"))
	if err != nil {
		log.Fatalf("Failed to register consumer: %v", err)
		return nil, nil, nil, err
	} else {
		log.Printf("Registered consumer: %s", os.Getenv("INPUT_QUEUE"))
	}

	return messages, conn, channel, nil
}

func StartMessageLoop(fn serviceFunc, messages <-chan amqp.Delivery, channel *amqp.Channel, routingKey string, exchangeName string) {
	if exchangeName == "" {
		exchangeName = "topic_exchange"
	}

	log.Printf("before messageloop of %s", routingKey)
	// Message loop stays alive
	for msg := range messages {
		log.Printf("StartMessageLoop: Received message: %v", string(msg.Body))
		newMsg, err := fn(msg)

		if err != nil {
			publishing := amqp.Publishing{
				Body: []byte("Error executing query: " + err.Error()),
			}
			err := channel.PublishWithContext(context.Background(), "dead-letter-exchange", routingKey, false, false, publishing)
			if err != nil {
				log.Fatalf("StartMessageLoop: Error publishing message: %v", err)
			}
		} else {
			err := channel.PublishWithContext(context.Background(), exchangeName, routingKey, false, false, newMsg)
			if err != nil {
				log.Printf("StartMessageLoop: Error publishing message: %v", err)
			}
		}
	}
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
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		amqp.Table{
			"x-dead-letter-exchange": "dead-letter-exchange",
		},
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

func Publish(queueName string, message amqp.Publishing, exchangeName string) error {
	if exchangeName == "" {
		exchangeName = "topic_exchange"
	}

	err := channel.PublishWithContext(context.Background(), exchangeName, queueName, false, false, message)
	if err != nil {
		return err
	}
	return nil
}
