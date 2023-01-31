package GoLib

import (
	"fmt"
	"os"
	"strings"
)

func ReadFile(fileName string) (string, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return string(""), err
	}
	return string(data), nil
}

func GetAMQConnectionString() (string, error) {
	user := os.Getenv("AMQ_USER")
	pwFile := os.Getenv("AMQ_PASSWORD_FILE")
	pw, err := ReadFile(pwFile)
	if err != nil {
		return "", err
	}

	pw = strings.TrimSuffix(pw, "\n")

	return fmt.Sprintf("amqp://%s:%s@rabbit:5672/", user, pw), nil
}
