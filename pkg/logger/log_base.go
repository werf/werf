package logger

import (
	"fmt"
	"io"
	"strings"
)

var (
	indentWidth = 0

	isLoggerCursorOnStartingPosition     = true
	isPrevLoggerCursorStateOnRemoveCaret = false
	isLoggerOptionalLnModeOn             = false
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
		case "\r", "\n":
			formattedMsg += processCaret(string(r))
		default:
			formattedMsg += processOptionalLnMode()
			formattedMsg += processDefault()
		}

		formattedMsg += string(r)
	}

	return logF(w, "%s", formattedMsg)
}

func processCaret(caret string) string {
	var result string

	if isLoggerCursorOnStartingPosition && !isPrevLoggerCursorStateOnRemoveCaret {
		result += processService()
	}

	isPrevLoggerCursorStateOnRemoveCaret = caret == "\r"
	isLoggerCursorOnStartingPosition = true

	return result
}

func processOptionalLnMode() string {
	var result string

	if isLoggerOptionalLnModeOn {
		result += processService()
		result += "\n"

		resetOptionalLnMode()
		isLoggerCursorOnStartingPosition = true
	}

	return result
}

func processDefault() string {
	var result string

	if isLoggerCursorOnStartingPosition {
		result += processService()
		result += strings.Repeat(" ", indentWidth)

		isLoggerCursorOnStartingPosition = false
	}

	isPrevLoggerCursorStateOnRemoveCaret = false

	return result
}

func processService() string {
	var result string

	result += formattedProcessBorders()
	result += formattedTag()

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
	indentWidth += 2
	resetOptionalLnMode()
}

func IndentDown() {
	if indentWidth == 0 {
		return
	}

	indentWidth -= 2
	resetOptionalLnMode()
}

func OptionalLnModeOn() {
	isLoggerOptionalLnModeOn = true
}

func resetOptionalLnMode() {
	isLoggerOptionalLnModeOn = false
}

func applyOptionalLnMode() {
	if _, err := logF(outStream, processOptionalLnMode()); err != nil {
		panic(err)
	}
}

func terminalContentWidth() int {
	return TerminalWidth() - terminalServiceWidth()
}

func terminalServiceWidth() int {
	return processBordersBlockWidth() + tagBlockWidth() + indentWidth
}
