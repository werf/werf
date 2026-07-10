package logging

import (
	"context"

	"github.com/gookit/color"

	"github.com/werf/logboek"
)

func BoldF(format string, a ...interface{}) string {
	return styleF(color.New(color.Bold), format, a...)
}

func BlueF(format string, a ...interface{}) string {
	return styleF(color.New(color.FgBlue), format, a...)
}

func YellowF(format string, a ...interface{}) string {
	return styleF(color.New(color.FgYellow), format, a...)
}

func GreenF(format string, a ...interface{}) string {
	return styleF(color.New(color.FgGreen), format, a...)
}

func RedF(format string, a ...interface{}) string {
	return styleF(color.New(color.FgRed), format, a...)
}

func styleF(style color.Style, format string, a ...interface{}) string {
	return logboek.Context(context.Background()).ColorizeF(style, format, a...)
}
