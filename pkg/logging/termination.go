package logging

import (
	"github.com/werf/logboek"
)

func Error(message string) {
	logboek.Streams().DisablePrefix()
	logboek.Streams().DisableLineWrapping()
	logboek.Error().LogF("Error: %s\n", message)
}

func Default(message string) {
	logboek.Default().LogLn(message)
}
