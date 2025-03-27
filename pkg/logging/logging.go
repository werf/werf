package logging

import (
	"os"

	"github.com/werf/logboek"
)

func Fatal(exitMessage string, exitCode int) {
	logboek.Streams().DisablePrefix()
	logboek.Streams().DisableLineWrapping()
	logboek.Error().LogLn(exitMessage)
	os.Exit(exitCode)
}
