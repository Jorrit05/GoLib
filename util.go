package GoLib

import (
	"log"
	"os"
)

func StartFileLog() (*os.File, error) {

	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	log.SetOutput(f)
	return f, nil
}
