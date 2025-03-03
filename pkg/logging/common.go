package logging

import (
	"context"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/level"
)

// DoWithLevel Saves current log level and restores it after execution of fn
func DoWithLevel(ctx context.Context, lvl level.Level, fn func() error) error {
	logger := logboek.Context(ctx)
	defer logger.SetAcceptedLevel(logger.AcceptedLevel())
	logger.SetAcceptedLevel(lvl)
	return fn()
}
