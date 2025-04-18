package logging

import (
	"context"
	"io"
	"log"

	"github.com/sirupsen/logrus"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/level"
	"github.com/werf/logboek/pkg/types"
	"github.com/werf/werf/v2/pkg/background"
)

// WithLogger returns new logger and bounds it to given context.
func WithLogger(ctx context.Context) context.Context {
	return logboek.NewContext(ctx, NewLogger())
}

// NewLogger returns new logger for any (foreground or background) mode.
func NewLogger() types.LoggerInterface {
	logger := logboek.DefaultLogger()

	if background.IsBackgroundModeEnabled() {
		logger.SetErrorStreamRedirection(level.Error)
	}

	captureOutputFromAnotherLoggers(logger.OutStream())

	return logger
}

func captureOutputFromAnotherLoggers(writer io.Writer) {
	log.SetOutput(writer)
	logrus.StandardLogger().SetOutput(writer)
}
