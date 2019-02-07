package logger

import (
	"fmt"
	"io"
	"strings"
)

var (
	indentWidth = 0

	tag       = ""
	tagIndent = "  "
	tagWidth  = 20

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
	indent := strings.Repeat(" ", indentWidth)

	var formattedMsg string
	for _, r := range []rune(msg) {
		switch string(r) {
		case "\n", "\r":
			if isLoggerCursorOnStartingPosition {
				formattedMsg += formattedTag()
			}

			isLoggerCursorOnStartingPosition = true
		default:
			if isLoggerOptionalLnModeOn {
				if isLoggerOptionalLnModeTag == tag {
					formattedMsg += formattedTag()
				}

				formattedMsg += "\n"
				resetOptionalLnMode()
			}

			if isLoggerCursorOnStartingPosition {
				formattedMsg += formattedTag()
				formattedMsg += indent
				isLoggerCursorOnStartingPosition = false
			}
		}

		formattedMsg += string(r)
	}

	return logF(w, formattedMsg)
}

func logF(w io.Writer, format string, args ...interface{}) (int, error) {
	var msg string
	if len(args) == 0 {
		msg = format
	} else {
		msg = fmt.Sprintf(format, args...)
	}

	return fmt.Fprintf(w, msg)
}

func WithIndent(f func() error) error {
	IndentUp()
	err := f()
	IndentDown()

	return err
}

func WithoutIndent(f func() error) error {
	oldIndentWidth := indentWidth
	indentWidth = 0
	err := f()
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

func WithTag(value string, f func() error) error {
	oldTag := tag
	tag = value
	err := f()
	tag = oldTag

	return err
}

func SetTag(value string) {
	tag = value
}

func ResetTag(value string) {
	tag = value
}

func formattedTag() string {
	var fittedTag string
	longTagPostfix := " ..."

	if len(tag) == 0 {
		return ""
	} else if len(tag) > tagWidth {
		fittedTag = tag[:tagWidth-len(longTagPostfix)] + longTagPostfix
	} else {
		fittedTag = tag

	}

	padLeft := strings.Repeat(" ", tagWidth-len(fittedTag))
	colorizedTag := colorize(fittedTag, tagFormat...)

	return strings.Join([]string{padLeft, colorizedTag, tagIndent}, "")
}

func tagBlockWidth() int {
	if len(tag) == 0 {
		return 0
	}

	return tagWidth + len(tagIndent)
}

func LogOptionalLn() {
	isLoggerOptionalLnModeOn = true
	isLoggerOptionalLnModeTag = tag
}

func resetOptionalLnMode() {
	isLoggerOptionalLnModeOn = false
	isLoggerOptionalLnModeTag = ""
}
