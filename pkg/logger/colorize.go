package logger

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
)

var (
	highlightFormat = []color.Attribute{color.FgYellow, color.Bold}
	serviceFormat   = []color.Attribute{color.Bold}
	infoFormat      = []color.Attribute{color.FgHiBlue}
	warningFormat   = []color.Attribute{color.FgRed, color.Bold}

	failFormat    = warningFormat
	successFormat = []color.Attribute{color.FgGreen, color.Bold}

	tagFormat = []color.Attribute{color.FgCyan}
)

func initColorize() {
	if os.Getenv("WERF_LOG_FORCE_COLOR") != "" {
		color.NoColor = false
	}
}

func colorizeAndFormattedLogF(w io.Writer, colorizeFunc func(string) string, format string, args ...interface{}) {
	var msg string
	if len(args) > 0 {
		msg = colorizeBaseF(colorizeFunc, format, args...)
	} else {
		msg = colorizeBaseF(colorizeFunc, "%s", format)
	}

	loggerFormattedLogF(w, msg)
}

func colorizeBaseF(colorizeFunc func(string) string, format string, args ...interface{}) string {
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

func colorizeFail(msg string) string {
	return colorize(msg, failFormat...)
}

func colorizeSuccess(msg string) string {
	return colorize(msg, successFormat...)
}

func colorizeStep(msg string) string {
	return colorize(msg, highlightFormat...)
}

func colorizeService(msg string) string {
	return colorize(msg, serviceFormat...)
}

func colorizeInfo(msg string) string {
	return colorize(msg, infoFormat...)
}

func colorizeWarning(msg string) string {
	return colorize(msg, warningFormat...)
}

func colorize(msg string, attributes ...color.Attribute) string {
	return color.New(attributes...).Sprint(msg)
}
