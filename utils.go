package GoLib

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	amqp "github.com/rabbitmq/amqp091-go"
)

type serviceFunc func(message amqp.Delivery) amqp.Publishing

func StartMessageLoop(fn serviceFunc, messages <-chan amqp.Delivery, channel *amqp.Channel, routingKey string) {
	// Message loop stays alive
	for msg := range messages {
		log.Printf("Received message: %v", string(msg.Body))
		anonymizedMsg := fn(msg)

		err := channel.PublishWithContext(context.Background(), "topic_exchange", routingKey, false, false, anonymizedMsg)
		if err != nil {
			log.Fatalf("Error publishing message: %v", err)
		}
	}
}

func ReadFile(fileName string) (string, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return string(""), err
	}
	str := strings.TrimSuffix(string(data), "\n")

	return str, nil
}

func GetAMQConnectionString() (string, error) {
	user := os.Getenv("AMQ_USER")
	pwFile := os.Getenv("AMQ_PASSWORD_FILE")
	pw, err := ReadFile(pwFile)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("amqp://%s:%s@rabbit:5672/", user, pw), nil
}

func GetSQLConnectionString() (string, error) {
	user := os.Getenv("DB_USER")
	pwFile := os.Getenv("MYSQL_ROOT_PASSWORD_FILE")
	pw, err := ReadFile(pwFile)
	if err != nil {
		return "", err
	}

	pw = strings.TrimSuffix(pw, "\n")

	return fmt.Sprintf("%s:%s@tcp(mysql:3306)/%s", user, pw, os.Getenv("MYSQL_DATABASE")), nil
}
