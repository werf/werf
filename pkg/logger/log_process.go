package logger

import (
	"fmt"
	"strings"
	"time"
)

const (
	logProcessSuccessStatus          = "[OK]"
	logProcessFailedStatus           = "[FAILED]"
	logProcessTimeFormat             = "%5.2f sec "
	logProcessInlineProcessMsgFormat = "%s ..."

	logStateRightPartsSeparator = " "
)

var (
	colorlessProcessBorders   []string
	processBorders            []string
	processBordersIndentWidth = 1
)

func LogProcessInline(msg string, processFunc func() error) error {
	return logProcessInlineBase(msg, processFunc, colorizeStep, colorizeSuccess)
}

func LogServiceProcessInline(msg string, processFunc func() error) error {
	return logProcessInlineBase(msg, processFunc, colorizeService, colorizeService)
}

func LogProcess(msg string, options LogProcessOptions, processFunc func() error) error {
	return logProcessBase(msg, options, processFunc, colorizeStep)
}

func LogServiceProcess(msg string, options LogProcessOptions, processFunc func() error) error {
	return logProcessBase(msg, options, processFunc, colorizeService)
}

func LogState(msg, state string) {
	options := logStateOptions{State: state}
	logStateBase(msg, options, colorizeService, colorizeService)
}

func LogServiceState(msg, state string) {
	options := logStateOptions{State: state}
	logStateBase(msg, options, colorizeService, colorizeService)
}

func logProcessInlineBase(processMsg string, processFunc func() error, colorizeProcessMsgFunc, colorizeSuccessFunc func(string) string) error {
	processMsg = fmt.Sprintf(logProcessInlineProcessMsgFormat, processMsg)
	colorizeAndLogF(outStream, colorizeProcessMsgFunc, "%s", processMsg)

	resultStatus := logProcessSuccessStatus
	resultColorize := colorizeSuccessFunc
	start := time.Now()

	err := WithIndent(processFunc)
	if err != nil {
		resultStatus = logProcessFailedStatus
		resultColorize = colorizeFail
	}

	elapsedSeconds := fmt.Sprintf(logProcessTimeFormat, time.Since(start).Seconds())

	rightPart := prepareLogStateRightPart(processMsg, resultStatus, elapsedSeconds, resultColorize)
	loggerFormattedLogLn(outStream, rightPart)

	return err
}

type LogProcessOptions struct {
	WithIndent           bool
	WithoutBorder        bool
	WithoutLogOptionalLn bool
	InfoSectionFunc      func(err error)
}

func logProcessBase(msg string, options LogProcessOptions, processFunc func() error, colorizeMsgFunc func(string) string) error {
	processOptionalLnMode()

	headerFunc := func() error {
		logStateOptions := logStateOptions{IgnoreIndent: true}
		logStateBase(msg, logStateOptions, colorizeMsgFunc, colorizeSuccess)
		return nil
	}

	if !options.WithoutBorder {
		headerFunc = decorateByWithExtraProcessBorder("┌", colorizeMsgFunc, headerFunc)
	}

	_ = headerFunc()

	start := time.Now()
	resultStatus := logProcessSuccessStatus

	bodyFunc := func() error {
		return processFunc()
	}

	if options.WithIndent {
		bodyFunc = decorateByWithIndent(bodyFunc)
	}

	if !options.WithoutBorder {
		bodyFunc = decorateByWithExtraProcessBorder("│", colorizeMsgFunc, bodyFunc)
	}

	err := bodyFunc()

	resetOptionalLnMode()

	if options.InfoSectionFunc != nil {
		loggerFormattedLogLn(outStream, colorizeMsgFunc(fmt.Sprintf("├ %s (info)", msg)))
		_ = decorateByWithExtraProcessBorder("│", colorizeMsgFunc, func() error {
			options.InfoSectionFunc(err)
			return nil
		})()
	}

	elapsedSeconds := fmt.Sprintf(logProcessTimeFormat, time.Since(start).Seconds())

	if err != nil {
		resultStatus = logProcessFailedStatus

		footerFunc := func() error {
			logStateOptions := logStateOptions{
				State:        resultStatus,
				Time:         elapsedSeconds,
				IgnoreIndent: true,
			}
			logStateBase(msg, logStateOptions, colorizeFail, colorizeFail)
			return nil
		}

		if !options.WithoutBorder {
			footerFunc = decorateByWithExtraProcessBorder("└", colorizeMsgFunc, footerFunc)
		}

		_ = footerFunc()

		if !options.WithoutLogOptionalLn {
			LogOptionalLn()
		}

		return err
	}

	footerFunc := func() error {
		logStateOptions := logStateOptions{
			State:        resultStatus,
			Time:         elapsedSeconds,
			IgnoreIndent: true,
		}
		logStateBase(msg, logStateOptions, colorizeMsgFunc, colorizeSuccess)
		return nil
	}

	if !options.WithoutBorder {
		footerFunc = decorateByWithExtraProcessBorder("└", colorizeMsgFunc, footerFunc)
	}

	_ = footerFunc()

	if !options.WithoutLogOptionalLn {
		LogOptionalLn()
	}

	return nil
}

type logStateOptions struct {
	State        string
	Time         string
	IgnoreIndent bool
}

func logStateBase(msg string, options logStateOptions, colorizeLeftPartFunc, colorizeRightPartFunc func(string) string) {
	action := func() error {
		leftPart := prepareLogStateLeftPart(msg, options.State, options.Time, colorizeLeftPartFunc)

		var rightPart string
		if options.State != "" {
			rightPart = prepareLogStateRightPart(msg, options.State, options.Time, colorizeRightPartFunc)
		}

		loggerFormattedLogLn(outStream, fmt.Sprintf("%s%s", leftPart, rightPart))

		return nil
	}

	if options.IgnoreIndent {
		action = decorateByWithoutIndent(action)
	}

	_ = action()
}

func prepareLogStateLeftPart(msg, state, time string, colorizeFunc func(string) string) string {
	var result string

	spaceLength := availableTerminalLineSpace(state, timeOrStub(time))
	if spaceLength > 0 {
		if spaceLength > len(msg) {
			result = msg
		} else {
			result = msg[0:spaceLength]
		}
	} else {
		return ""
	}

	return colorizeFunc(result)
}

func prepareLogStateRightPart(msg, state, time string, colorizeFunc func(string) string) string {
	var result string
	spaceLength := availableTerminalLineSpace(msg)

	rightPartLength := len(state + timeOrStub(time) + logStateRightPartsSeparator)
	if spaceLength-rightPartLength > 0 {
		result += strings.Repeat(" ", spaceLength-rightPartLength)
	}

	var rightPart []string
	rightPart = append(rightPart, colorizeFunc(state))
	rightPart = append(rightPart, colorizeFunc(time))

	result += strings.Join(rightPart, logStateRightPartsSeparator)

	return result
}

func timeOrStub(time string) string {
	if time == "" {
		return fmt.Sprintf(logProcessTimeFormat, 0.0)
	}

	return time
}

func availableTerminalLineSpace(parts ...string) int {
	msgsLength := len(strings.Join(parts, " "))

	return TerminalWidth() - tagBlockWidth() - indentWidth - processBordersBlockWidth() - msgsLength
}

func decorateByWithExtraProcessBorder(colorlessBorder string, colorizeFunc func(string) string, decoratedFunc func() error) func() error {
	return func() error {
		return withExtraProcessBorder(colorlessBorder, colorizeFunc, decoratedFunc)
	}
}

func withExtraProcessBorder(colorlessValue string, colorizeFunc func(string) string, decoratedFunc func() error) error {
	appendProcessBorder(colorlessValue, colorizeFunc)
	err := decoratedFunc()
	popProcessBorder()

	return err
}

func appendProcessBorder(colorlessValue string, colorizeFunc func(string) string) {
	colorlessProcessBorders = append(colorlessProcessBorders, colorlessValue)
	processBorders = append(processBorders, colorizeFunc(colorlessValue))
}

func popProcessBorder() {
	if len(processBorders) == 0 {
		return
	}

	colorlessProcessBorders = colorlessProcessBorders[:len(colorlessProcessBorders)-1]
	processBorders = processBorders[:len(processBorders)-1]
}

func formattedProcessBorders() string {
	if len(processBorders) == 0 {
		return ""
	}

	return strings.Join(processBorders, " ") + strings.Repeat(" ", processBordersIndentWidth)
}

func processBordersBlockWidth() int {
	if len(colorlessProcessBorders) == 0 {
		return 0
	}

	return len(strings.Join(colorlessProcessBorders, " ")) + processBordersIndentWidth
}
