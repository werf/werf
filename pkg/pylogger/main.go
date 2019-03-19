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

//export SetTerminalWidth
func SetTerminalWidth(width C.int) {
	logger.SetTerminalWidth(int(width))
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

//export Log
func Log(data *C.char) {
	logger.LogF("%s", C.GoString(data))
}

//export LogHighlight
func LogHighlight(data *C.char) {
	logger.LogHighlightF("%s", C.GoString(data))
}

//export LogService
func LogService(data *C.char) {
	logger.LogServiceF("%s", C.GoString(data))
}

//export LogInfo
func LogInfo(data *C.char) {
	logger.LogInfoF("%s", C.GoString(data))
}

//export LogError
func LogError(data *C.char) {
	logger.LogErrorF("%s", C.GoString(data))
}

//export LogProcessStart
func LogProcessStart(msg *C.char) {
	logger.LogProcessStart(C.GoString(msg), logger.LogProcessStartOptions{})
}

//export LogProcessEnd
func LogProcessEnd(withoutLogOptionalLn bool) {
	logger.LogProcessEnd(logger.LogProcessEndOptions{WithoutLogOptionalLn: withoutLogOptionalLn})
}

//export LogProcessStepEnd
func LogProcessStepEnd(msg *C.char) {
	logger.LogProcessStepEnd(C.GoString(msg))
}

//export LogProcessFail
func LogProcessFail(withoutLogOptionalLn bool) {
	logger.LogProcessFail(logger.LogProcessEndOptions{WithoutLogOptionalLn: withoutLogOptionalLn})
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

//export FittedStreamsOutputOn
func FittedStreamsOutputOn() {
	logger.FittedStreamsOutputOn()
}

//export FittedStreamsOutputOff
func FittedStreamsOutputOff() {
	logger.FittedStreamsOutputOff()
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

//export Out
func Out(msg *C.char) {
	logger.OutF("%s", C.GoString(msg))
}

//export Err
func Err(msg *C.char) {
	logger.ErrF("%s", C.GoString(msg))
}

func main() {}
