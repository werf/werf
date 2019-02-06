package logger

func LogServiceLn(msg string) {
	LogServiceF("%s\n", msg)
}

func LogServiceF(format string, args ...interface{}) {
	colorizeAndLogF(outStream, colorizeService, format, args...)
}

func LogInfoLn(msg string) {
	LogInfoF("%s\n", msg)
}

func LogInfoF(format string, args ...interface{}) {
	colorizeAndLogF(outStream, colorizeInfo, format, args...)
}

func LogErrorLn(msg string) {
	LogErrorF("%s\n", msg)
}

func LogErrorF(format string, args ...interface{}) {
	colorizeAndLogF(errStream, colorizeWarning, format, args...)
}
