package main

import (
	"./api"
	"fmt"
	"os"
)

type Test interface {
	RunTest([]string)
}

func main() {

	tests := []struct {
		name          string
		testingStruct Test
		args          []string
	}{
		{"kerneladiutor",
		 api.KernelAdiutorTest{},
		 []string{"port"},
		},
	}

	if len(os.Args) > 1 {
		var validTest bool = false
		for _, test := range tests {
			if test.name == os.Args[1] {
				validTest = true

				if len(os.Args)-2 == len(test.args) {
					test.testingStruct.RunTest(os.Args[2:])
				} else {
					var testArgs string = ""
					for _, arg := range test.args {
						testArgs = testArgs + "[" + arg + "]"
					}
					fmt.Println(os.Args[0], os.Args[1], testArgs)
				}
			}
		}

		if !validTest {
			fmt.Println(os.Args[1], "is not a test")
		}
	} else {
		fmt.Println(os.Args[0], "[test] <args>")
	}
}
