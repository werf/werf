package logger

import "fmt"

func LogLn(a ...interface{}) {
	loggerFormattedLogLn(outStream, a...)
}

func LogF(format string, a ...interface{}) {
	loggerFormattedLogF(outStream, format, a...)
}

func LogHighlightLn(a ...interface{}) {
	LogHighlightF("%s", fmt.Sprintln(a...))
}

func LogHighlightF(format string, a ...interface{}) {
	colorizeAndFormattedLogF(outStream, colorizeHighlight, format, a...)
}

func LogServiceLn(a ...interface{}) {
	LogServiceF("%s", fmt.Sprintln(a...))
}

func LogServiceF(format string, a ...interface{}) {
	colorizeAndFormattedLogF(outStream, colorizeSecondary, format, a...)
}

func LogInfoLn(a ...interface{}) {
	LogInfoF("%s", fmt.Sprintln(a...))
}

func LogInfoF(format string, a ...interface{}) {
	colorizeAndFormattedLogF(outStream, colorizeInfo, format, a...)
}

func LogErrorLn(a ...interface{}) {
	LogErrorF("%s", fmt.Sprintln(a...))
}

func LogErrorF(format string, a ...interface{}) {
	colorizeAndFormattedLogF(errStream, colorizeWarning, format, a...)
}
