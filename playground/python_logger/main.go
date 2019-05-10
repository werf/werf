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

	logboek.LogProcess("Running task", logboek.LogProcessOptions{ColorizeMsgFunc: logboek.ColorizeHighlight}, func() error {
		return logboek.LogProcess("Running subtask", logboek.LogProcessOptions{WithIndent: true, ColorizeMsgFunc: logboek.ColorizeHighlight}, func() error {
			logboek.LogF("HELO!\n")
			logboek.LogF("HELO!\n")
			return nil
		})
	})
}
