package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/werf/werf/v2/pkg/werf/exec"
)

func main() {
	if isForegroundMode() {
		startDetachedProcess()
	} else {
		handleDetachedProcess()
	}
}

func isForegroundMode() bool {
	return os.Getenv("_WERF_BACKGROUND_MODE_ENABLED") != "1"
}

// startDetachedProcess starts detached process and exits currently running process.
func startDetachedProcess() {
	ctx := context.Background()
	args := os.Args[1:]

	if err := exec.Detach(ctx, args); err != nil {
		log.Fatalf("detaching error: %s\n", err.Error())
	}
	return
}

func handleDetachedProcess() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	signalsChan := make(chan os.Signal, 1)
	signal.Notify(signalsChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-signalsChan:
		log.Println("Signal received")
	case <-ctx.Done():
		log.Fatal("Timeout exceeded") // exit_code = 1
	}
}
