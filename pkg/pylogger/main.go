package main

import (
	"C"
)
import "github.com/flant/logboek"

//export Init
func Init() *C.char {
	if err := logboek.Init(); err != nil {
		return C.CString(err.Error())
	}
	return nil
}

//export DisablePrettyLog
func DisablePrettyLog() {
	logboek.DisablePrettyLog()
}

//export EnableLogColor
func EnableLogColor() {
	logboek.EnableLogColor()
}

//export DisableLogColor
func DisableLogColor() {
	logboek.DisableLogColor()
}

//export SetTerminalWidth
func SetTerminalWidth(width C.int) {
	logboek.SetTerminalWidth(int(width))
}

//export IndentUp
func IndentUp() {
	logboek.IndentUp()
}

//export IndentDown
func IndentDown() {
	logboek.IndentDown()
}

//export OptionalLnModeOn
func OptionalLnModeOn() {
	logboek.OptionalLnModeOn()
}

//export Log
func Log(data *C.char) {
	logboek.LogF("%s", C.GoString(data))
}

//export LogHighlight
func LogHighlight(data *C.char) {
	logboek.LogHighlightF("%s", C.GoString(data))
}

//export LogService
func LogService(data *C.char) {
	logboek.LogServiceF("%s", C.GoString(data))
}

//export LogInfo
func LogInfo(data *C.char) {
	logboek.LogInfoF("%s", C.GoString(data))
}

//export LogError
func LogError(data *C.char) {
	logboek.LogErrorF("%s", C.GoString(data))
}

//export LogProcessStart
func LogProcessStart(msg *C.char) {
	logboek.LogProcessStart(C.GoString(msg), logboek.LogProcessStartOptions{})
}

//export LogProcessEnd
func LogProcessEnd(withoutLogOptionalLn bool) {
	logboek.LogProcessEnd(logboek.LogProcessEndOptions{WithoutLogOptionalLn: withoutLogOptionalLn})
}

//export LogProcessStepEnd
func LogProcessStepEnd(msg *C.char) {
	logboek.LogProcessStepEnd(C.GoString(msg))
}

//export LogProcessFail
func LogProcessFail(withoutLogOptionalLn bool) {
	logboek.LogProcessFail(logboek.LogProcessEndOptions{WithoutLogOptionalLn: withoutLogOptionalLn})
}

//export FitText
func FitText(text *C.char, extraIndentWidth, maxWidth int, markWrappedFile bool) *C.char {
	return C.CString(logboek.FitText(C.GoString(text), logboek.FitTextOptions{
		ExtraIndentWidth: extraIndentWidth,
		MaxWidth:         maxWidth,
		MarkWrappedLine:  markWrappedFile,
	}))
}

//export GetRawStreamsOutputMode
func GetRawStreamsOutputMode() bool {
	return logboek.GetRawStreamsOutputMode()
}

//export RawStreamsOutputModeOn
func RawStreamsOutputModeOn() {
	logboek.RawStreamsOutputModeOn()
}

//export RawStreamsOutputModeOff
func RawStreamsOutputModeOff() {
	logboek.RawStreamsOutputModeOff()
}

//export FittedStreamsOutputOn
func FittedStreamsOutputOn() {
	logboek.FittedStreamsOutputOn()
}

//export FittedStreamsOutputOff
func FittedStreamsOutputOff() {
	logboek.FittedStreamsOutputOff()
}

//export MuteOut
func MuteOut() {
	logboek.MuteOut()
}

//export UnmuteOut
func UnmuteOut() {
	logboek.UnmuteOut()
}

//export MuteErr
func MuteErr() {
	logboek.MuteErr()
}

//export UnmuteErr
func UnmuteErr() {
	logboek.UnmuteErr()
}

//export Out
func Out(msg *C.char) {
	logboek.OutF("%s", C.GoString(msg))
}

//export Err
func Err(msg *C.char) {
	logboek.ErrF("%s", C.GoString(msg))
}

func main() {}
