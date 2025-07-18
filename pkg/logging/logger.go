package logging

import (
	"context"
	"io"
	"log"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/types"
)

// WithLogger returns new logger and bounds it to given context.
func WithLogger(ctx context.Context) context.Context {
	return logboek.NewContext(ctx, NewLogger())
}

// NewLogger returns new logger for any (foreground or background) mode.
func NewLogger() types.LoggerInterface {
	logger := logboek.DefaultLogger()
	logger.Warn().SetStyle(color.New(color.Yellow))

	captureOutputFromAnotherLoggers(logger.OutStream())

	return logger
}

func captureOutputFromAnotherLoggers(writer io.Writer) {
	log.SetOutput(writer)
	logrus.StandardLogger().SetOutput(writer)
}
