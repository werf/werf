package main

import (
	"fmt"
	"os"
	"time"

	"github.com/flant/werf/pkg/ssh_agent"
)

func main() {
	fmt.Printf("keys: %v\n", os.Args[1:])

	err := ssh_agent.Init(os.Args[1:])
	if err != nil {
		panic(err)
	}

	time.Sleep(60 * time.Second)
}
