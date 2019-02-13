package logger

import (
	"fmt"
	"io"
	"strings"
)

var (
	indentWidth = 0

	isLoggerCursorOnStartingPosition = true
	isLoggerOptionalLnModeOn         = false
	isLoggerOptionalLnModeTag        = ""
)

func loggerFormattedLogLn(w io.Writer, msg string) {
	loggerFormattedLogF(w, "%s\n", msg)
}

func loggerFormattedLogF(w io.Writer, format string, args ...interface{}) {
	if _, err := FormattedLogF(w, format, args...); err != nil {
		panic(err)
	}
}

func FormattedLogF(w io.Writer, format string, args ...interface{}) (int, error) {
	msg := fmt.Sprintf(format, args...)

	var formattedMsg string
	for _, r := range []rune(msg) {
		switch string(r) {
		case "\n", "\r":
			formattedMsg += loggerCursorOnStartingPositionCaretRuneHook()
		default:
			formattedMsg += loggerOptionalLnModeDefaultHook()
			formattedMsg += loggerCursorOnStartingPositionDefaultHook()
		}

		formattedMsg += string(r)
	}

	return logF(w, formattedMsg)
}

func loggerCursorOnStartingPositionCaretRuneHook() string {
	var result string

	if isLoggerCursorOnStartingPosition {
		result += formattedTag()
		result += formattedProcessBorders()
	}

	isLoggerCursorOnStartingPosition = true

	return result
}

func loggerCursorOnStartingPositionDefaultHook() string {
	var result string

	if isLoggerCursorOnStartingPosition {
		result += formattedTag()
		result += formattedProcessBorders()
		result += strings.Repeat(" ", indentWidth)
		isLoggerCursorOnStartingPosition = false
	}

	return result
}

func loggerOptionalLnModeDefaultHook() string {
	var result string

	if isLoggerOptionalLnModeOn {
		if isLoggerOptionalLnModeTag == colorlessTag {
			result += formattedTag()
			result += formattedProcessBorders()
		}

		result += "\n"
		resetOptionalLnMode()
	}

	return result
}

func logF(w io.Writer, format string, args ...interface{}) (int, error) {
	var msg string
	if len(args) == 0 {
		msg = format
	} else {
		msg = fmt.Sprintf(format, args...)
	}

	return fmt.Fprintf(w, "%s", msg)
}

func decorateByWithIndent(decoratedFunc func() error) func() error {
	return func() error {
		return WithIndent(decoratedFunc)
	}
}

func WithIndent(f func() error) error {
	IndentUp()
	err := f()
	IndentDown()

	return err
}

func decorateByWithoutIndent(decoratedFunc func() error) func() error {
	return func() error {
		return WithoutIndent(decoratedFunc)
	}
}

func WithoutIndent(decoratedFunc func() error) error {
	oldIndentWidth := indentWidth
	indentWidth = 0
	err := decoratedFunc()
	indentWidth = oldIndentWidth

	return err
}

func IndentUp() {
	resetOptionalLnMode()
	indentWidth += 2
}

func IndentDown() {
	if indentWidth == 0 {
		return
	}

	resetOptionalLnMode()
	indentWidth -= 2
}

func LogOptionalLn() {
	isLoggerOptionalLnModeOn = true
	isLoggerOptionalLnModeTag = colorlessTag
}

func resetOptionalLnMode() {
	isLoggerOptionalLnModeOn = false
	isLoggerOptionalLnModeTag = ""
}

func processOptionalLnMode() {
	_, _ = logF(outStream, loggerOptionalLnModeDefaultHook())
}
