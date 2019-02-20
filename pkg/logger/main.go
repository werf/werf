package logger

func Init() {
	initColorize()
	initTerminalWidth()
}

func DisablePrettyLog() {
	RawStreamsOutputModeOn()
	disableLogProcessBorder()
}
