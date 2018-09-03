package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	stdoutLines := 1
	stderrLines := 1

	for {
		fmt.Printf("stdout line %d\n", stdoutLines)
		stdoutLines++
		fmt.Printf("stdout line %d\n", stdoutLines)
		stdoutLines++

		time.Sleep(1 * time.Second)

		fmt.Fprintf(os.Stderr, "stderr line %d\n", stderrLines)
		stderrLines++

		time.Sleep(1 * time.Second)

		if stdoutLines > 10 {
			break
		}
	}
}
