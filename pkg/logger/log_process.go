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

func LogProcessInline(msg string, processFunc func() error) error {
	return logProcessInlineBase(msg, processFunc, colorizeStep, colorizeSuccess)
}

func LogServiceProcessInline(msg string, processFunc func() error) error {
	return logProcessInlineBase(msg, processFunc, colorizeService, colorizeService)
}

func LogProcess(msg, processMsg string, processFunc func() error) error {
	return logProcessBase(msg, processMsg, processFunc, colorizeStep, colorizeSuccess)
}

func LogServiceProcess(msg, processMsg string, processFunc func() error) error {
	return logProcessBase(msg, processMsg, processFunc, colorizeService, colorizeService)
}

func LogState(msg, state string) {
	logStateBase(msg, state, "", colorizeService, colorizeService)
}

func LogServiceState(msg, state string) {
	logStateBase(msg, state, "", colorizeService, colorizeService)
}

func logProcessInlineBase(processMsg string, processFunc func() error, colorizeProcessMsgFunc, colorizeSuccessFunc func(string) string) error {
	processMsg = fmt.Sprintf(logProcessInlineProcessMsgFormat, processMsg)
	colorizeAndLogF(outStream, colorizeProcessMsgFunc, "%s", processMsg)

	resultStatus := logProcessSuccessStatus
	resultColorize := colorizeSuccessFunc
	start := time.Now()

	err := WithLogIndent(processFunc)
	if err != nil {
		resultStatus = logProcessFailedStatus
		resultColorize = colorizeFail
	}

	elapsedSeconds := fmt.Sprintf(logProcessTimeFormat, time.Since(start).Seconds())

	rightPart := prepareLogStateRightPart(processMsg, resultStatus, elapsedSeconds, resultColorize)
	loggerFormattedLogLn(outStream, rightPart)

	return err
}

func logProcessBase(msg, processMsg string, processFunc func() error, colorizeMsgFunc, colorizeSuccessFunc func(string) string) error {
	logStateBase(msg, processMsg, "", colorizeMsgFunc, colorizeSuccessFunc)

	start := time.Now()
	resultStatus := logProcessSuccessStatus

	err := WithLogIndent(processFunc)

	elapsedSeconds := fmt.Sprintf(logProcessTimeFormat, time.Since(start).Seconds())

	if err != nil {
		resultStatus = logProcessFailedStatus
		logStateBase(msg, resultStatus, elapsedSeconds, colorizeFail, colorizeFail)

		return err
	}

	logStateBase(msg, resultStatus, elapsedSeconds, colorizeMsgFunc, colorizeSuccessFunc)

	return nil
}

func logStateBase(msg string, state, time string, colorizeLeftPartFunc, colorizeRightPartFunc func(string) string) {
	leftPart := prepareLogStateLeftPart(msg, state, time, colorizeLeftPartFunc)

	var rightPart string
	if state != "" {
		rightPart = prepareLogStateRightPart(msg, state, time, colorizeRightPartFunc)
	}

	loggerFormattedLogLn(outStream, fmt.Sprintf("%s%s", leftPart, rightPart))
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

	return TerminalWidth() - indentWidth - msgsLength
}
