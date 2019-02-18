package logger

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
)

var (
	highlightFormat = []color.Attribute{color.Bold}
	secondaryFormat []color.Attribute
	infoFormat      = []color.Attribute{color.FgHiBlue}
	warningFormat   = []color.Attribute{color.FgRed, color.Bold}

	failFormat    = warningFormat
	successFormat = []color.Attribute{color.FgGreen, color.Bold}
)

func initColorize() {
	if os.Getenv("WERF_LOG_FORCE_COLOR") != "" {
		color.NoColor = false
	}
}

func colorizeAndFormattedLogF(w io.Writer, colorizeFunc func(...interface{}) string, format string, args ...interface{}) {
	var msg string
	if len(args) > 0 {
		msg = colorizeBaseF(colorizeFunc, format, args...)
	} else {
		msg = colorizeBaseF(colorizeFunc, "%s", format)
	}

	loggerFormattedLogF(w, msg)
}

func colorizeBaseF(colorizeFunc func(...interface{}) string, format string, args ...interface{}) string {
	var colorizeLines []string
	lines := strings.Split(fmt.Sprintf(format, args...), "\n")
	for _, line := range lines {
		if line == "" {
			colorizeLines = append(colorizeLines, line)
		} else {
			colorizeLines = append(colorizeLines, colorizeFunc(line))
		}
	}

	return strings.Join(colorizeLines, "\n")
}

func colorizeFail(a ...interface{}) string {
	return colorize(failFormat, a...)
}

func colorizeSuccess(a ...interface{}) string {
	return colorize(successFormat, a...)
}

func colorizeHighlight(a ...interface{}) string {
	return colorize(highlightFormat, a...)
}

func colorizeSecondary(a ...interface{}) string {
	return colorize(secondaryFormat, a...)
}

func colorizeInfo(a ...interface{}) string {
	return colorize(infoFormat, a...)
}

func colorizeWarning(a ...interface{}) string {
	return colorize(warningFormat, a...)
}

func colorize(attributes []color.Attribute, a ...interface{}) string {
	return color.New(attributes...).Sprint(a...)
}
