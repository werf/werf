package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/werf/werf/v2/pkg/background"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/pkg/werf/exec"
)

func main() {
	if background.IsBackgroundModeEnabled() {
		handleDetachedProcess()
	} else {
		startDetachedProcess()
	}
}

// startDetachedProcess starts detached process and exits currently running process.
func startDetachedProcess() {
	ctx := context.TODO()
	args := os.Args[1:]

	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("unable to get working directory: %s\n", err.Error())
	}

	if err := werf.Init(workingDir, workingDir); err != nil {
		log.Fatalf("werf init error: %s\n", err.Error())
	}

	if err := exec.Detach(ctx, args, nil); err != nil {
		log.Fatalf("detaching error: %s\n", err.Error())
	}
	return
}

func handleDetachedProcess() {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*10)
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
