package main

import (
	"C"

	"github.com/flant/werf/pkg/logger"
)

//export Init
func Init() *C.char {
	if err := logger.Init(); err != nil {
		return C.CString(err.Error())
	}
	return nil
}

//export DisablePrettyLog
func DisablePrettyLog() {
	logger.DisablePrettyLog()
}

//export EnableLogColor
func EnableLogColor() {
	logger.EnableLogColor()
}

//export DisableLogColor
func DisableLogColor() {
	logger.DisableLogColor()
}

//export IndentUp
func IndentUp() {
	logger.IndentUp()
}

//export IndentDown
func IndentDown() {
	logger.IndentDown()
}

//export OptionalLnModeOn
func OptionalLnModeOn() {
	logger.OptionalLnModeOn()
}

//export LogLn
func LogLn(data *C.char) {
	logger.LogLn(C.GoString(data))
}

//export LogHighlightLn
func LogHighlightLn(data *C.char) {
	logger.LogHighlightLn(C.GoString(data))
}

//export LogServiceLn
func LogServiceLn(data *C.char) {
	logger.LogServiceLn(C.GoString(data))
}

//export LogInfoLn
func LogInfoLn(data *C.char) {
	logger.LogInfoLn(C.GoString(data))
}

//export LogErrorLn
func LogErrorLn(data *C.char) {
	logger.LogErrorLn(C.GoString(data))
}

//export FitText
func FitText(text *C.char, extraIndentWidth, maxWidth int, markWrappedFile bool) *C.char {
	return C.CString(logger.FitText(C.GoString(text), logger.FitTextOptions{
		ExtraIndentWidth: extraIndentWidth,
		MaxWidth:         maxWidth,
		MarkWrappedLine:  markWrappedFile,
	}))
}

//export GetRawStreamsOutputMode
func GetRawStreamsOutputMode() bool {
	return logger.GetRawStreamsOutputMode()
}

//export RawStreamsOutputModeOn
func RawStreamsOutputModeOn() {
	logger.RawStreamsOutputModeOn()
}

//export RawStreamsOutputModeOff
func RawStreamsOutputModeOff() {
	logger.RawStreamsOutputModeOff()
}

//export MuteOut
func MuteOut() {
	logger.MuteOut()
}

//export UnmuteOut
func UnmuteOut() {
	logger.UnmuteOut()
}

//export MuteErr
func MuteErr() {
	logger.MuteErr()
}

//export UnmuteErr
func UnmuteErr() {
	logger.UnmuteErr()
}

//func StartLogProcess(msg *C.char, withIndent bool, withoutLogOptionalLn bool) *C.char {
//
//}

//func StopLogProcess() *C.char {
//
//}

// func LogProcessStart(msg string) *C.char {
// 	err := logger.LogProcess(msg, logger.LogProcessOptions{}, func() error {
// 		return nil
// 	})
//
// 	if err != nil {
// 		return C.CString(err.Error())
// 	}
// 	return nil
// }

func main() {}
