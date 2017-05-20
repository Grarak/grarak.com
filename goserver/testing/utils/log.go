package utils

import (
	"fmt"
	"strings"
)

func Passed(message ...interface{}) {
	fmt.Println("\x1b[32;1mPASSED\x1b[0m:", strings.Trim(fmt.Sprintf("%s", message), "[]"))
}

func Failed(message ...interface{}) {
	fmt.Println("\x1b[31;1mFAILED\x1b[0m:", strings.Trim(fmt.Sprintf("%s", message), "[]"))
}
