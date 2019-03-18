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

var (
	activeLogProcesses []*logProcessDescriptor
)

type logProcessDescriptor struct {
	StartedAt time.Time
	Msg       string
}

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

type LogProcessStartOptions struct {
	ColorizeMsgFunc func(...interface{}) string
}

type LogProcessEndOptions struct {
	WithoutLogOptionalLn bool
}

type LogProcessStepEndOptions struct {
	WithIndent      bool
	InfoSectionFunc func(err error)
}

type LogProcessOptions struct {
	WithIndent           bool
	WithoutLogOptionalLn bool
	InfoSectionFunc      func(err error)
	ColorizeMsgFunc      func(...interface{}) string
}

func LogProcessStart(msg string, options LogProcessStartOptions) {
	baseLogProcessStart(msg, options, colorizeHighlight)
}

func LogProcessEnd(options LogProcessEndOptions) {
	baseLogProcessEnd(options, colorizeHighlight)
}

func LogProcessStepEnd(msg string) {
	baseLogProcessStepEnd(msg, colorizeHighlight)
}

func LogProcessFail(options LogProcessEndOptions) {
	baseLogProcessFail(options, colorizeHighlight)
}

func LogProcess(msg string, options LogProcessOptions, processFunc func() error) error {
	return logProcessBase(msg, options, processFunc, colorizeHighlight)
}

func LogSecondaryProcess(msg string, options LogProcessOptions, processFunc func() error) error {
	return logProcessBase(msg, options, processFunc, colorizeSecondary)
}

func baseLogProcessStart(msg string, options LogProcessStartOptions, colorizeMsgFunc func(...interface{}) string) {
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

	appendProcessBorder(logProcessVerticalBorderSign, colorizeMsgFunc)

	logProcess := &logProcessDescriptor{StartedAt: time.Now(), Msg: msg}
	activeLogProcesses = append(activeLogProcesses, logProcess)
}

func baseLogProcessStepEnd(msg string, colorizeMsgFunc func(...interface{}) string) {
	msgFunc := func() error {
		return WithoutIndent(func() error {
			loggerFormattedLogLn(outStream, prepareLogProcessMsgLeftPart(msg, colorizeMsgFunc))
			return nil
		})
	}

	msgFunc = decorateByWithExtraProcessBorder(logProcessVerticalAndRightBorderSign, colorizeMsgFunc, msgFunc)
	msgFunc = decorateByWithoutLastProcessBorder(msgFunc)

	_ = msgFunc()
}

func applyInfoLogProcessStep(userError error, infoSectionFunc func(err error), withIndent bool, colorizeMsgFunc func(...interface{}) string) {
	infoHeaderFunc := func() error {
		return WithoutIndent(func() error {
			loggerFormattedLogLn(outStream, prepareLogProcessMsgLeftPart("Info", colorizeMsgFunc))
			return nil
		})
	}

	infoHeaderFunc = decorateByWithExtraProcessBorder(logProcessVerticalAndRightBorderSign, colorizeMsgFunc, infoHeaderFunc)
	infoHeaderFunc = decorateByWithoutLastProcessBorder(infoHeaderFunc)

	_ = infoHeaderFunc()

	infoFunc := func() error {
		infoSectionFunc(userError)
		return nil
	}

	if withIndent {
		infoFunc = decorateByWithIndent(infoFunc)
	}

	infoFunc = decorateByWithExtraProcessBorder(logProcessVerticalBorderSign, colorizeMsgFunc, infoFunc)
	infoFunc = decorateByWithoutLastProcessBorder(infoFunc)

	_ = infoFunc()
}

func baseLogProcessEnd(options LogProcessEndOptions, colorizeMsgFunc func(...interface{}) string) {
	popProcessBorder()

	logProcess := activeLogProcesses[len(activeLogProcesses)-1]
	activeLogProcesses = activeLogProcesses[:len(activeLogProcesses)-1]

	resetOptionalLnMode()

	elapsedSeconds := fmt.Sprintf(logProcessTimeFormat, time.Since(logProcess.StartedAt).Seconds())

	footerFunc := func() error {
		return WithoutIndent(func() error {
			timePart := fmt.Sprintf(" (%s)", elapsedSeconds)
			loggerFormattedLogF(outStream, prepareLogProcessMsgLeftPart(logProcess.Msg, colorizeMsgFunc, timePart))
			colorizeAndFormattedLogF(outStream, colorizeMsgFunc, "%s\n", timePart)

			return nil
		})
	}

	footerFunc = decorateByWithExtraProcessBorder(logProcessUpAndRightBorderSign, colorizeMsgFunc, footerFunc)

	_ = footerFunc()

	if !options.WithoutLogOptionalLn {
		OptionalLnModeOn()
	}
}

func baseLogProcessFail(options LogProcessEndOptions, colorizeMsgFunc func(...interface{}) string) {
	popProcessBorder()

	logProcess := activeLogProcesses[len(activeLogProcesses)-1]
	activeLogProcesses = activeLogProcesses[:len(activeLogProcesses)-1]

	resetOptionalLnMode()

	elapsedSeconds := fmt.Sprintf(logProcessTimeFormat, time.Since(logProcess.StartedAt).Seconds())

	footerFunc := func() error {
		return WithoutIndent(func() error {
			timePart := fmt.Sprintf(" (%s) FAILED", elapsedSeconds)
			loggerFormattedLogF(outStream, prepareLogProcessMsgLeftPart(logProcess.Msg, colorizeFail, timePart))
			colorizeAndFormattedLogF(outStream, colorizeFail, "%s\n", timePart)

			return nil
		})
	}

	footerFunc = decorateByWithExtraProcessBorder(logProcessUpAndRightBorderSign, colorizeMsgFunc, footerFunc)

	_ = footerFunc()

	if !options.WithoutLogOptionalLn {
		OptionalLnModeOn()
	}
}

func logProcessBase(msg string, options LogProcessOptions, processFunc func() error, colorizeMsgFunc func(...interface{}) string) error {
	baseLogProcessStart(msg, LogProcessStartOptions{ColorizeMsgFunc: options.ColorizeMsgFunc}, colorizeMsgFunc)

	bodyFunc := func() error {
		return processFunc()
	}

	if options.WithIndent {
		bodyFunc = decorateByWithIndent(bodyFunc)
	}

	err := bodyFunc()

	resetOptionalLnMode()

	if options.InfoSectionFunc != nil {
		applyInfoLogProcessStep(err, options.InfoSectionFunc, options.WithIndent, colorizeMsgFunc)
	}

	if err != nil {
		baseLogProcessFail(LogProcessEndOptions{WithoutLogOptionalLn: options.WithoutLogOptionalLn}, colorizeMsgFunc)
		return err
	}

	baseLogProcessEnd(LogProcessEndOptions{WithoutLogOptionalLn: options.WithoutLogOptionalLn}, colorizeMsgFunc)
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

func decorateByWithoutLastProcessBorder(decoratedFunc func() error) func() error {
	return func() error {
		return withoutLastProcessBorder(decoratedFunc)
	}
}

func withoutLastProcessBorder(f func() error) error {
	oldBorderValue := processesBorderValues[len(processesBorderValues)-1]
	processesBorderValues = processesBorderValues[:len(processesBorderValues)-1]

	oldBorderFormattedValue := processesBorderFormattedValues[len(processesBorderFormattedValues)-1]
	processesBorderFormattedValues = processesBorderFormattedValues[:len(processesBorderFormattedValues)-1]

	err := f()

	processesBorderValues = append(processesBorderValues, oldBorderValue)
	processesBorderFormattedValues = append(processesBorderFormattedValues, oldBorderFormattedValue)

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
