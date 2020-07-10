package main

import (
	"os"
	"strconv"
	"time"
)

// Time to run the chaos in minutes
var chaosTime time.Duration

func main() {
	var duration int
	if len(os.Args) > 1 {
		duration, _ = strconv.Atoi(os.Args[1])
	} else {
		duration = 20
	}
	chaosTime = time.Duration(duration)
	chaosTest()
}
