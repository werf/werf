package logger

func Init() {
	initTerminalWidth()
}

func DisablePrettyLog() {
	RawStreamsOutputModeOn()
	disableLogProcessBorder()
}
