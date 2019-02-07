package logger

import (
	"fmt"
	"io"
	"strings"
)

var (
	indentWidth                      = 0
	isLoggerCursorOnStartingPosition = true
)

func loggerFormattedLogLn(w io.Writer, msg string) {
	loggerFormattedLogF(w, "%s\n", msg)
}

func loggerFormattedLogF(w io.Writer, format string, args ...interface{}) {
	if _, err := FormattedLogF(w, format, args...); err != nil {
		panic(err)
	}
}

// TODO
// AppendTag("a")
// PopTag() // => delete "a"
// WithTags("a", "b", "c") {
// 	WithTags("d") {
// 		// a b c d : ...
// 		// a b c d : ...
// 		// a b c d : ...
// 		// a b c d : ...
// 	}
// }

func FormattedLogF(w io.Writer, format string, args ...interface{}) (int, error) {
	msg := fmt.Sprintf(format, args...)
	indent := strings.Repeat(" ", indentWidth)

	var formattedMsg string
	for _, r := range []rune(msg) {
		switch string(r) {
		case "\n", "\r":
			isLoggerCursorOnStartingPosition = true
		default:
			if isLoggerCursorOnStartingPosition {
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

func WithLogIndent(f func() error) error {
	IndentUp()
	err := f()
	IndentDown()

	return err
}

func WithoutLogIndent(f func() error) error {
	oldIndentWidth := indentWidth
	indentWidth = 0
	err := f()
	indentWidth = oldIndentWidth

	return err
}

func IndentUp() {
	indentWidth += 2
}

func IndentDown() {
	if indentWidth == 0 {
		return
	}

	indentWidth -= 2
}
