package utils

import (
	"fmt"
	"time"
)

func LogI(tag, message string) {
	log("I", tag, message)
}

func LogE(tag, message string) {
	log("E", tag, message)
}

func log(t, tag, message string) {
	fmt.Printf("%s: %s/%s: %s\n", time.Now().Format("2006-01-02 15:04:05"), t, tag, message)
}
