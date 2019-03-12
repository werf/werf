package logger

import (
	"fmt"
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

const (
	DefaultTerminalWidth = 140
)

var terminalWidth int

func initTerminalWidth() error {
	if terminal.IsTerminal(int(os.Stdout.Fd())) {
		w, _, err := terminal.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			return fmt.Errorf("get terminal size failed: %s", err)
		}

		SetTerminalWidth(w)
	} else {
		SetTerminalWidth(DefaultTerminalWidth)
	}

	return nil
}

func SetTerminalWidth(value int) {
	terminalWidth = value
}

func IsTerminal() bool {
	return terminal.IsTerminal(int(os.Stdout.Fd()))
}

func TerminalWidth() int {
	return terminalWidth
}
