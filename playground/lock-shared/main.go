package main

import (
	"fmt"
	"github.com/flant/dapp/pkg/lock"
	"time"
)

func main() {
	fl := lock.NewFileLock("helo")

	err := fl.WithLock(
		time.Second*10, true,
		func(doWait func() error) error {
			fmt.Printf("Before waiting\n")
			err := doWait()
			fmt.Printf("After waiting => %s\n", err)
			return err
		},
		func() error {
			fmt.Printf("Lock acquired! Sleep for 10 seconds\n")
			time.Sleep(10 * time.Second)
			return nil
		},
	)

	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
	}
}
