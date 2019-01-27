package logger

import (
	"fmt"
	"os"
	"strconv"

	"golang.org/x/crypto/ssh/terminal"
)

const (
	defaultTerminalWidth = 120
)

func TerminalWidth() int {
	if wtw, ok := os.LookupEnv("WERF_TERMINAL_WIDTH"); ok {
		if i, err := strconv.Atoi(wtw); err != nil {
			panic(fmt.Sprintf("Unexpected WERF_TERMINAL_WIDTH: %s", err))
		} else {
			return i
		}
	} else {
		if terminal.IsTerminal(int(os.Stdout.Fd())) {
			w, _, err := terminal.GetSize(int(os.Stdout.Fd()))
			if err != nil {
				panic(err)
			}

			return w
		}
	}

	return defaultTerminalWidth
}
