package utils

import (
	"fmt"
	"time"
	"io/ioutil"
	"os"
)

func LogI(tag, message string) {
	log("I", tag, message)
}

func LogE(tag, message string) {
	log("E", tag, message)
}

func log(t, tag, message string) {
	text := fmt.Sprintf("%s: %s/%s: %s\n", time.Now().Format("2006-01-02 15:04:05"), t, tag, message)
	fmt.Printf(text)

	logFile, err := os.OpenFile(SERVERDATA+"/log.txt", os.O_APPEND|os.O_WRONLY, 0644)
	if err == nil {
		logFile.Write([]byte(text))
		logFile.Close()
	} else {
		ioutil.WriteFile(SERVERDATA+"/log.txt", []byte(text), 0644)
	}
}
