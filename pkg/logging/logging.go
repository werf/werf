package logging

import (
	"github.com/werf/logboek"
)

func Error(message string) {
	logboek.Streams().DisablePrefix()
	logboek.Streams().DisableLineWrapping()
	logboek.Error().LogLn(message)
}
