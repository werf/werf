package main

import (
	"fmt"
	"os"
	"time"

	"github.com/flant/werf/pkg/lock"
)

func main() {
	err := lock.Init()
	if err != nil {
		panic(err)
	}

	opts := lock.LockOptions{}

	if os.Getenv("READONLY") == "1" {
		opts.ReadOnly = true
	}

	err = lock.WithLock("helo", opts, func() error {
		fmt.Printf("Lock acquired! Sleep for 10 seconds \n")
		time.Sleep(10 * time.Second)
		return nil
	})

	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
	}
}
