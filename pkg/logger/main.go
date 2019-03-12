package logger

func Init() error {
	return initTerminalWidth()
}

func DisablePrettyLog() {
	RawStreamsOutputModeOn()
	disableLogProcessBorder()
}
