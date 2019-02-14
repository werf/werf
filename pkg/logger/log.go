package logger

func LogLn(args ...interface{}) {
	loggerFormattedLogLn(outStream, args...)
}

func LogF(format string, args ...interface{}) {
	loggerFormattedLogF(outStream, format, args...)
}

func LogServiceLn(args ...interface{}) {
	LogServiceF("%s\n", args...)
}

func LogServiceF(format string, args ...interface{}) {
	colorizeAndFormattedLogF(outStream, colorizeService, format, args...)
}

func LogInfoLn(args ...interface{}) {
	LogInfoF("%s\n", args...)
}

func LogInfoF(format string, args ...interface{}) {
	colorizeAndFormattedLogF(outStream, colorizeInfo, format, args...)
}

func LogErrorLn(args ...interface{}) {
	LogErrorF("%s\n", args...)
}

func LogErrorF(format string, args ...interface{}) {
	colorizeAndFormattedLogF(errStream, colorizeWarning, format, args...)
}
