package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/flant/shluz"

	"github.com/flant/werf/pkg/werf"
)

func main() {
	err := shluz.Init(filepath.Join(werf.GetServiceDir(), "locks"))
	if err != nil {
		panic(err)
	}

	opts := shluz.LockOptions{}

	if os.Getenv("READONLY") == "1" {
		opts.ReadOnly = true
	}

	err = shluz.WithLock("helo", opts, func() error {
		fmt.Printf("Lock acquired! Sleep for 10 seconds \n")
		time.Sleep(10 * time.Second)
		return nil
	})

	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
	}
}
