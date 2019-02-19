package cleaning

import "github.com/flant/werf/pkg/logger"

type CleanupOptions struct {
	ImagesCleanupOptions ImagesCleanupOptions
	StagesCleanupOptions StagesCleanupOptions
}

func Cleanup(options CleanupOptions) error {
	if err := logger.LogProcess("Running images cleanup", logger.LogProcessOptions{WithIndent: true}, func() error {
		return ImagesCleanup(options.ImagesCleanupOptions)
	}); err != nil {
		return err
	}

	if err := logger.LogProcess("Running stages cleanup", logger.LogProcessOptions{WithIndent: true}, func() error {
		return StagesCleanup(options.StagesCleanupOptions)
	}); err != nil {
		return err
	}

	return nil
}
