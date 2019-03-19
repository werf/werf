package logger

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

var (
	outStream io.Writer = os.Stdout
	errStream io.Writer = os.Stderr

	isRawStreamsOutputModeOn    = false
	isFittedStreamsOutputModeOn = false

	streamsFitterState fitterState
)

type WriterProxy struct {
	io.Writer
}

func (p WriterProxy) Write(data []byte) (int, error) {
	if isRawStreamsOutputModeOn {
		return logF(p.Writer, "%s", string(data))
	}

	msg := string(data)
	if isFittedStreamsOutputModeOn {
		msg, streamsFitterState = fitText(msg, streamsFitterState, TerminalContentWidth(), true, true)
	}

	_, err := FormattedLogF(p.Writer, "%s", msg)
	return len(data), err
}

func WithRawStreamsOutputModeOn(f func() error) error {
	savedIsRawOutputModeOn := isRawStreamsOutputModeOn
	isRawStreamsOutputModeOn = true
	err := f()
	isRawStreamsOutputModeOn = savedIsRawOutputModeOn

	return err
}

func GetRawStreamsOutputMode() bool {
	return isRawStreamsOutputModeOn
}

func RawStreamsOutputModeOn() {
	isRawStreamsOutputModeOn = true
}

func RawStreamsOutputModeOff() {
	isRawStreamsOutputModeOn = false
}

func WithFittedStreamsOutputOn(f func() error) error {
	savedStreamsFitterState := streamsFitterState
	savedIsFittedOutputModeOn := isFittedStreamsOutputModeOn

	streamsFitterState = fitterState{}
	isFittedStreamsOutputModeOn = true
	err := f()
	streamsFitterState = savedStreamsFitterState
	isFittedStreamsOutputModeOn = savedIsFittedOutputModeOn

	return err
}

func FittedStreamsOutputOn() {
	streamsFitterState = fitterState{}
	isFittedStreamsOutputModeOn = true
}

func FittedStreamsOutputOff() {
	streamsFitterState = fitterState{}
	isFittedStreamsOutputModeOn = false
}

func GetOutStream() io.Writer {
	return WriterProxy{outStream}
}

func GetErrStream() io.Writer {
	return WriterProxy{errStream}
}

func MuteOut() {
	outStream = ioutil.Discard
}

func UnmuteOut() {
	outStream = os.Stdout
}

func MuteErr() {
	errStream = ioutil.Discard
}

func UnmuteErr() {
	errStream = os.Stderr
}

func OutF(format string, a ...interface{}) {
	fmt.Fprintf(GetOutStream(), format, a...)
}

func ErrF(format string, a ...interface{}) {
	fmt.Fprintf(GetErrStream(), format, a...)
}
