package GoLib

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func StartLog() *os.File {
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	return f
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

func GetDefaultOutputQueue(serviceName string) string {
	routingKey := os.Getenv("ROUTING_KEY")
	if routingKey == "" {
		routingKey = "service." + strings.Split(serviceName, "_")[0]
	}
	return routingKey
}

// func ConvertRequest(req *http.Request) {
// 	body, err := ioutil.ReadAll(req.Body)
// 	if err != nil {
// 		log.Printf("Error reading body: %v", err)
// 		http.Error(w, "Error reading request body", http.StatusBadRequest)
// 		return
// 	}
// 	defer req.Body.Close()
// 	return body
// }
