package main

import "github.com/flant/werf/pkg/logger"

func main() {
	logger.Init()

	logger.LogProcessStart("Running task", logger.LogProcessStartOptions{})
	logger.WithIndent(func() error {
		logger.LogProcessStart("Running subtask", logger.LogProcessStartOptions{})
		logger.LogF("HELO!\n")
		logger.LogF("HELO!\n")
		logger.LogProcessFail(logger.LogProcessEndOptions{})
		return nil
	})
	logger.LogProcessEnd(logger.LogProcessEndOptions{})

	logger.LogProcessStart("Running task", logger.LogProcessStartOptions{})
	logger.LogProcessStepEnd("Item X done")
	logger.LogProcessStepEnd("Item Y done")
	logger.LogProcessStepEnd("Item Z done")
	logger.LogProcessEnd(logger.LogProcessEndOptions{})

	logger.LogProcess("Running task", logger.LogProcessOptions{}, func() error {
		return logger.LogProcess("Running subtask", logger.LogProcessOptions{WithIndent: true}, func() error {
			logger.LogF("HELO!\n")
			logger.LogF("HELO!\n")
			return nil
		})
	})
}
