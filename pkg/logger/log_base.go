package logger

import (
	"fmt"
	"io"
	"strings"
)

var (
	indentWidth = 0

	isLoggerCursorOnStartingPosition = true
	isLoggerCursorOnRemoveCaret      = false
	isLoggerOptionalLnModeOn         = false

	cursor int
)

func loggerFormattedLogLn(w io.Writer, a ...interface{}) {
	loggerFormattedLogF(w, fmt.Sprintln(a...))
}

func loggerFormattedLogF(w io.Writer, format string, a ...interface{}) {
	if _, err := FormattedLogF(w, format, a...); err != nil {
		panic(err)
	}
}

func FormattedLogF(w io.Writer, format string, a ...interface{}) (int, error) {
	var msg string
	if len(a) != 0 {
		msg = fmt.Sprintf(format, a...)
	} else {
		msg = format
	}

	var formattedMsg string
	for _, r := range []rune(msg) {
		switch string(r) {
		case "\r":
			formattedMsg += loggerCursorOnStartingPositionRemoveCaretRuneHook()
			cursor = 0
			isLoggerCursorOnRemoveCaret = true
		case "\n":
			formattedMsg += loggerCursorOnStartingPositionNewLineCaretRuneHook()
			cursor = 0
			isLoggerCursorOnRemoveCaret = false
		default:
			formattedMsg += loggerOptionalLnModeDefaultHook()
			formattedMsg += loggerCursorOnStartingPositionDefaultHook()
			cursor += 1
			isLoggerCursorOnRemoveCaret = false
		}

		formattedMsg += string(r)
	}

	return logF(w, "%s", formattedMsg)
}

func loggerCursorOnStartingPositionRemoveCaretRuneHook() string {
	var result string

	if isLoggerCursorOnStartingPosition {
		result += formattedProcessBorders()
		result += formattedTag()
	}

	isLoggerCursorOnStartingPosition = true

	return result
}

func loggerCursorOnStartingPositionNewLineCaretRuneHook() string {
	var result string

	if isLoggerCursorOnStartingPosition && !isLoggerCursorOnRemoveCaret {
		result += formattedProcessBorders()
		result += formattedTag()
	}

	isLoggerCursorOnStartingPosition = true

	return result
}

func loggerCursorOnStartingPositionDefaultHook() string {
	var result string

	if isLoggerCursorOnStartingPosition {
		result += formattedProcessBorders()
		result += formattedTag()
		result += strings.Repeat(" ", indentWidth)

		isLoggerCursorOnStartingPosition = false
	}

	return result
}

func loggerOptionalLnModeDefaultHook() string {
	var result string

	if isLoggerOptionalLnModeOn {
		result += formattedProcessBorders()
		result += formattedTag()
		result += "\n"

		isLoggerCursorOnStartingPosition = true
		resetOptionalLnMode()
	}

	return result
}

func logF(w io.Writer, format string, a ...interface{}) (int, error) {
	return fmt.Fprintf(w, format, a...)
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
}

func resetOptionalLnMode() {
	isLoggerOptionalLnModeOn = false
	cursor = 0
}

func processOptionalLnMode() {
	if _, err := logF(outStream, loggerOptionalLnModeDefaultHook()); err != nil {
		panic(err)
	}

	resetOptionalLnMode()
}

func terminalContentWidth() int {
	return TerminalWidth() - terminalServiceWidth()
}

func terminalServiceWidth() int {
	return processBordersBlockWidth() + tagBlockWidth() + indentWidth
}
