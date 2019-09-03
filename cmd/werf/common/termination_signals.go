package common

import (
	"os"
	"os/signal"
	"syscall"
)

var (
	terminationSignalsTrapEnabled     bool
	terminationSignalsChan            chan os.Signal
	disableTerminationSignalsTrapChan chan struct{}
)

func EnableTerminationSignalsTrap() {
	disableTerminationSignalsTrapChan = make(chan struct{}, 1)
	terminationSignalsChan = make(chan os.Signal, 1)
	signals := []os.Signal{os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT}
	signal.Notify(terminationSignalsChan, signals...)

	go func() {
		select {
		case <-terminationSignalsChan:
			LogError("interrupted")
			os.Exit(17)
		case <-disableTerminationSignalsTrapChan:
			return
		}
	}()

	terminationSignalsTrapEnabled = true
}

func DisableTerminationSignalsTrap() {
	signal.Stop(terminationSignalsChan)
	disableTerminationSignalsTrapChan <- struct{}{}
	terminationSignalsTrapEnabled = false
}

func WithoutTerminationSignalsTrap(f func() error) error {
	savedTrapEnabled := terminationSignalsTrapEnabled

	if terminationSignalsTrapEnabled {
		DisableTerminationSignalsTrap()
	}

	defer func() {
		if savedTrapEnabled {
			EnableTerminationSignalsTrap()
		}
	}()

	return f()
}

func WithTerminationSignalsTrap(f func() error) error {
	savedTrapEnabled := terminationSignalsTrapEnabled

	if !terminationSignalsTrapEnabled {
		EnableTerminationSignalsTrap()
	}

	defer func() {
		if !savedTrapEnabled {
			DisableTerminationSignalsTrap()
		}
	}()

	return f()
}
