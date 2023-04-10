package GoLib

import (
	"fmt"
	"os"

	logrus "github.com/sirupsen/logrus"
)

func InitLogger(serviceName string) (*logrus.Entry, *os.File) {
	logFile, err := os.OpenFile(fmt.Sprintf("/var/log/service_logs/%s.log", serviceName), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logrus.Fatalf("Failed to open logfile %s: %v", serviceName, err)
	}

	env := os.Getenv("Environment")

	if env == "Production" {
		// Only log the Info level severity or above.
		logrus.SetLevel(logrus.WarnLevel)
	} else {
		logrus.SetLevel(logrus.DebugLevel)
	}

	// Set Logrus output to the log file
	logrus.SetOutput(logFile)
	logrus.SetFormatter(&logrus.JSONFormatter{})

	// Disable to prevent 20%-40% overhead
	logrus.SetReportCaller(true)

	return logrus.WithFields(logrus.Fields{
		"service": serviceName,
	}), logFile
}
