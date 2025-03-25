package logging

import (
	"context"
	"io"
	"log"

	"github.com/sirupsen/logrus"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/types"
	"github.com/werf/werf/v2/pkg/background"
)

// WithLogger returns new logger and bounds it to giver context.
func WithLogger(ctx context.Context) (context.Context, background.CloseFunc) {
	logger, closeOutput := NewLogger()
	return logboek.NewContext(ctx, logger), closeOutput
}

// NewLogger returns new logger and closeOutput func. The closeOutput function must be called.
// In foreground mode output is stdout and stderr.
// In background mode (detached host cleanup process) output is a file.
func NewLogger() (types.LoggerInterface, background.CloseFunc) {
	var logger types.LoggerInterface
	var closeFn func()

	if background.IsBackgroundModeEnabled() {
		out, closeOutput := background.Output()
		logger = logboek.NewLogger(out, out)
		closeFn = closeOutput
	}

	logger = logboek.DefaultLogger()

	captureOutputFromAnotherLoggers(logger.OutStream())

	return logger, closeFn
}

func captureOutputFromAnotherLoggers(writer io.Writer) {
	log.SetOutput(writer)
	logrus.StandardLogger().SetOutput(writer)
}
