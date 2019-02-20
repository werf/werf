package logger

import (
	"io"
	"io/ioutil"
	"os"
)

var (
	outStream io.Writer = os.Stdout
	errStream io.Writer = os.Stderr

	isRawStreamsOutputModeOn    = false
	isFittedStreamsOutputModeOn = false
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
		msg = fitText(msg, contentCursor, terminalContentWidth(), true)
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

func RawStreamsOutputModeOn() {
	isRawStreamsOutputModeOn = true
}

func WithFittedOutputOn(f func() error) error {
	savedIsFittedOutputModeOn := isFittedStreamsOutputModeOn
	isFittedStreamsOutputModeOn = true
	err := f()
	isFittedStreamsOutputModeOn = savedIsFittedOutputModeOn

	return err
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
