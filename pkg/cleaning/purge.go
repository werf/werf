package cleaning

import "github.com/flant/werf/pkg/logger"

type PurgeOptions struct {
	CommonRepoOptions    CommonRepoOptions
	CommonProjectOptions CommonProjectOptions
}

func Purge(options PurgeOptions) error {
	if err := logger.LogProcess("Running images purge", logger.LogProcessOptions{WithIndent: true}, func() error {
		return ImagesPurge(options.CommonRepoOptions)
	}); err != nil {
		return err
	}

	if err := logger.LogProcess("Running stages purge", logger.LogProcessOptions{WithIndent: true}, func() error {
		return StagesPurge(options.CommonProjectOptions)
	}); err != nil {
		return err
	}

	return nil
}
