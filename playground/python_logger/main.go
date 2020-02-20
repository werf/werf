package main

import "github.com/flant/logboek"

func main() {
	logboek.Init()

	logboek.LogProcessStart("Running task", logboek.LogProcessStartOptions{})
	logboek.WithIndent(func() error {
		logboek.LogProcessStart("Running subtask", logboek.LogProcessStartOptions{})
		logboek.LogF("HELO!\n")
		logboek.LogF("HELO!\n")
		logboek.LogProcessFail(logboek.LogProcessEndOptions{})
		return nil
	})
	logboek.LogProcessEnd(logboek.LogProcessEndOptions{})

	logboek.LogProcessStart("Running task", logboek.LogProcessStartOptions{})
	logboek.LogProcessStepEnd("Item X done", logboek.LogProcessStepEndOptions{})
	logboek.LogProcessStepEnd("Item Y done", logboek.LogProcessStepEndOptions{})
	logboek.LogProcessStepEnd("Item Z done", logboek.LogProcessStepEndOptions{})
	logboek.LogProcessEnd(logboek.LogProcessEndOptions{})

	logboek.Default.LogProcess("Running task", logboek.LevelLogProcessOptions{Style: logboek.HighlightStyle()}, func() error {
		return logboek.Default.LogProcess("Running subtask", logboek.LevelLogProcessOptions{WithIndent: true, Style: logboek.HighlightStyle()}, func() error {
			logboek.LogF("HELO!\n")
			logboek.LogF("HELO!\n")
			return nil
		})
	})
}
