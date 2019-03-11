package logger

import (
	"fmt"
	"strings"
	"time"
)

const (
	logProcessTimeFormat = "%.2f seconds"

	logProcessInlineProcessMsgFormat = "%s ..."
	logStateRightPartsSeparator      = " "
)

var (
	logProcessDownAndRightBorderSign     = "┌"
	logProcessVerticalBorderSign         = "│"
	logProcessVerticalAndRightBorderSign = "├"
	logProcessUpAndRightBorderSign       = "└"

	processesBorderValues             []string
	processesBorderFormattedValues    []string
	processesBorderBetweenIndentWidth = 1
	processesBorderIndentWidth        = 1
)

func disableLogProcessBorder() {
	logProcessDownAndRightBorderSign = ""
	logProcessVerticalBorderSign = "  "
	logProcessVerticalAndRightBorderSign = ""
	logProcessUpAndRightBorderSign = ""

	processesBorderIndentWidth = 0
	processesBorderBetweenIndentWidth = 0
}

func LogProcessInline(msg string, processFunc func() error) error {
	return logProcessInlineBase(msg, processFunc, colorizeHighlight)
}

func LogSecondaryProcessInline(msg string, processFunc func() error) error {
	return logProcessInlineBase(msg, processFunc, colorizeSecondary)
}

func logProcessInlineBase(processMsg string, processFunc func() error, colorizeProcessMsgFunc func(...interface{}) string) error {
	processMsg = fmt.Sprintf(logProcessInlineProcessMsgFormat, processMsg)
	colorizeAndFormattedLogF(outStream, colorizeProcessMsgFunc, "%s", processMsg)

	resultColorize := colorizeProcessMsgFunc
	start := time.Now()

	resultFormat := " (%s)\n"

	var err error
	if err = WithIndent(processFunc); err != nil {
		resultColorize = colorizeFail
		resultFormat = " (%s) FAILED\n"
	}

	elapsedSeconds := fmt.Sprintf(logProcessTimeFormat, time.Since(start).Seconds())
	colorizeAndFormattedLogF(outStream, resultColorize, resultFormat, elapsedSeconds)

	return err
}

func prepareLogProcessMsgLeftPart(leftPart string, colorizeFunc func(...interface{}) string, rightParts ...string) string {
	var result string

	spaceWidth := TerminalContentWidth() - len(strings.Join(rightParts, logStateRightPartsSeparator))
	if spaceWidth > 0 {
		if spaceWidth > len(leftPart) {
			result = leftPart
		} else {
			service := " ..."
			if spaceWidth > len(" ...") {
				result = leftPart[0:spaceWidth-len(service)] + service
			} else {
				result = leftPart[0:spaceWidth]
			}
		}
	} else {
		return ""
	}

	return colorizeFunc(result)
}

type LogProcessOptions struct {
	WithIndent           bool
	WithoutLogOptionalLn bool
	InfoSectionFunc      func(err error)
	ColorizeMsgFunc      func(...interface{}) string
}

func LogProcess(msg string, options LogProcessOptions, processFunc func() error) error {
	return logProcessBase(msg, options, processFunc, colorizeHighlight)
}

func LogSecondaryProcess(msg string, options LogProcessOptions, processFunc func() error) error {
	return logProcessBase(msg, options, processFunc, colorizeSecondary)
}

func logProcessBase(msg string, options LogProcessOptions, processFunc func() error, colorizeMsgFunc func(...interface{}) string) error {
	applyOptionalLnMode()

	if options.ColorizeMsgFunc != nil {
		colorizeMsgFunc = options.ColorizeMsgFunc
	}

	headerFunc := func() error {
		return WithoutIndent(func() error {
			loggerFormattedLogLn(outStream, prepareLogProcessMsgLeftPart(msg, colorizeMsgFunc))
			return nil
		})
	}

	headerFunc = decorateByWithExtraProcessBorder(logProcessDownAndRightBorderSign, colorizeMsgFunc, headerFunc)

	_ = headerFunc()

	start := time.Now()

	bodyFunc := func() error {
		return processFunc()
	}

	if options.WithIndent {
		bodyFunc = decorateByWithIndent(bodyFunc)
	}

	bodyFunc = decorateByWithExtraProcessBorder(logProcessVerticalBorderSign, colorizeMsgFunc, bodyFunc)

	err := bodyFunc()

	resetOptionalLnMode()

	if options.InfoSectionFunc != nil {
		infoHeaderFunc := func() error {
			return WithoutIndent(func() error {
				loggerFormattedLogLn(outStream, prepareLogProcessMsgLeftPart("Info", colorizeMsgFunc))
				return nil
			})
		}

		infoHeaderFunc = decorateByWithExtraProcessBorder(logProcessVerticalAndRightBorderSign, colorizeMsgFunc, infoHeaderFunc)

		_ = infoHeaderFunc()

		infoFunc := func() error {
			options.InfoSectionFunc(err)
			return nil
		}

		if options.WithIndent {
			infoFunc = decorateByWithIndent(infoFunc)
		}

		infoFunc = decorateByWithExtraProcessBorder(logProcessVerticalBorderSign, colorizeMsgFunc, infoFunc)

		_ = infoFunc()
	}

	elapsedSeconds := fmt.Sprintf(logProcessTimeFormat, time.Since(start).Seconds())

	if err != nil {
		footerFunc := func() error {
			return WithoutIndent(func() error {
				timePart := fmt.Sprintf(" (%s) FAILED", elapsedSeconds)
				loggerFormattedLogF(outStream, prepareLogProcessMsgLeftPart(msg, colorizeFail, timePart))
				colorizeAndFormattedLogF(outStream, colorizeFail, "%s\n", timePart)

				return nil
			})
		}

		footerFunc = decorateByWithExtraProcessBorder(logProcessUpAndRightBorderSign, colorizeMsgFunc, footerFunc)

		_ = footerFunc()

		if !options.WithoutLogOptionalLn {
			OptionalLnModeOn()
		}

		return err
	}

	footerFunc := func() error {
		return WithoutIndent(func() error {
			timePart := fmt.Sprintf(" (%s)", elapsedSeconds)
			loggerFormattedLogF(outStream, prepareLogProcessMsgLeftPart(msg, colorizeMsgFunc, timePart))
			colorizeAndFormattedLogF(outStream, colorizeMsgFunc, "%s\n", timePart)

			return nil
		})
	}

	footerFunc = decorateByWithExtraProcessBorder(logProcessUpAndRightBorderSign, colorizeMsgFunc, footerFunc)

	_ = footerFunc()

	if !options.WithoutLogOptionalLn {
		OptionalLnModeOn()
	}

	return nil
}

func decorateByWithExtraProcessBorder(colorlessBorder string, colorizeFunc func(...interface{}) string, decoratedFunc func() error) func() error {
	return func() error {
		return withExtraProcessBorder(colorlessBorder, colorizeFunc, decoratedFunc)
	}
}

func withExtraProcessBorder(colorlessValue string, colorizeFunc func(...interface{}) string, decoratedFunc func() error) error {
	appendProcessBorder(colorlessValue, colorizeFunc)
	err := decoratedFunc()
	popProcessBorder()

	return err
}

func appendProcessBorder(colorlessValue string, colorizeFunc func(...interface{}) string) {
	processesBorderValues = append(processesBorderValues, colorlessValue)
	processesBorderFormattedValues = append(processesBorderFormattedValues, colorizeFunc(colorlessValue))
}

func popProcessBorder() {
	if len(processesBorderValues) == 0 {
		return
	}

	processesBorderValues = processesBorderValues[:len(processesBorderValues)-1]
	processesBorderFormattedValues = processesBorderFormattedValues[:len(processesBorderFormattedValues)-1]
}

func formattedProcessBorders() string {
	if len(processesBorderValues) == 0 {
		return ""
	}

	return strings.Join(processesBorderFormattedValues, strings.Repeat(" ", processesBorderBetweenIndentWidth)) + strings.Repeat(" ", processesBorderIndentWidth)
}

func processBordersBlockWidth() int {
	if len(processesBorderValues) == 0 {
		return 0
	}

	return len([]rune(strings.Join(processesBorderValues, strings.Repeat(" ", processesBorderBetweenIndentWidth)))) + processesBorderIndentWidth
}
