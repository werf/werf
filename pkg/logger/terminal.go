package logger

import (
	"fmt"
	"os"
	"strconv"

	"golang.org/x/crypto/ssh/terminal"
)

const (
	defaultTerminalWidth = 140
)

var terminalWidth int

func initTerminalWidth() {
	if wtw, ok := os.LookupEnv("WERF_TERMINAL_WIDTH"); ok {
		if i, err := strconv.Atoi(wtw); err != nil {
			panic(fmt.Sprintf("Unexpected WERF_TERMINAL_WIDTH: %s", err))
		} else {
			terminalWidth = i

			return
		}
	} else {
		if terminal.IsTerminal(int(os.Stdout.Fd())) {
			w, _, err := terminal.GetSize(int(os.Stdout.Fd()))
			if err != nil {
				panic(err)
			}

			terminalWidth = w

			return
		}
	}

	terminalWidth = defaultTerminalWidth
}

func IsTerminal() bool {
	return terminal.IsTerminal(int(os.Stdout.Fd()))
}

func TerminalWidth() int {
	return terminalWidth
}
