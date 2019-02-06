package logger

import (
	"fmt"
	"io"
	"strings"
)

var (
	indentWidth = 0
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
	var linesWithIndent []string

	lines := strings.Split(fmt.Sprintf(format, args...), "\n")
	indent := strings.Repeat(" ", indentWidth)
	for _, line := range lines {
		if line == "" {
			linesWithIndent = append(linesWithIndent, line)
		} else {
			linesWithIndent = append(linesWithIndent, fmt.Sprintf("%s%s", indent, line))
		}
	}

	return logF(w, strings.Join(linesWithIndent, "\n"))
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
