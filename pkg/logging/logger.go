package logging

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/types"
	"github.com/werf/werf/v2/pkg/background"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/pkg/werf/global_warnings"
)

// WithLogger returns new logger and bounds it to given context.
func WithLogger(ctx context.Context) (context.Context, error) {
	logger, err := NewLogger(ctx)
	if err != nil {
		return nil, err
	}
	return logboek.NewContext(ctx, logger), nil
}

// NewLogger returns new logger for both foreground and background modes.
// Also, it checks existence of errors in recent background running, and it warns globally, if there were.
func NewLogger(ctx context.Context) (types.LoggerInterface, error) {
	// We need initialize werf variables to be able to call werf.GetServiceDir().
	if err := werf.PartialInit("", ""); err != nil {
		return nil, err
	}

	var logger types.LoggerInterface

	if background.IsBackgroundModeEnabled() {
		outStream, errStream, err := backgroundOutput(werf.GetServiceDir())
		if err != nil {
			return nil, err
		}

		logger = logboek.NewLogger(outStream, errStream)
	} else {
		logger = logboek.DefaultLogger()

		err := globalWarnIfBackgroundErrorHappened(ctx, werf.GetServiceDir())
		if err != nil {
			return nil, err
		}
	}

	captureOutputFromAnotherLoggers(logger.OutStream())

	return logger, nil
}

func captureOutputFromAnotherLoggers(writer io.Writer) {
	log.SetOutput(writer)
	logrus.StandardLogger().SetOutput(writer)
}

func globalWarnIfBackgroundErrorHappened(ctx context.Context, werfServiceDir string) error {
	errFilename := backgroundErrorFilename(werfServiceDir)

	file, err := openFile(errFilename, os.O_RDONLY)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	if stat.Size() == 0 {
		return nil
	}

	global_warnings.GlobalWarningLn(ctx, fmt.Sprintf(`Recent running of "werf host cleanup" in background mode was ended with errors.

Please, check these errors in %s file and remove its file after.`, errFilename))

	return nil
}
