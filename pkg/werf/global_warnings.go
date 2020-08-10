package werf

import (
	"github.com/werf/logboek"
)

var (
	GlobalWarningLines []string
)

func PrintGlobalWarnings() {
	for _, line := range GlobalWarningLines {
		printGlobalWarningLn(line)
	}
}

func GlobalWarningLn(line string) {
	GlobalWarningLines = append(GlobalWarningLines, line)
	printGlobalWarningLn(line)
}

func printGlobalWarningLn(line string) {
	logboek.Error().LogF("WARNING: %s\n", line)
}
