package logger

import (
	"io"
	"io/ioutil"
	"os"
)

var (
	outStream io.Writer = os.Stdout
	errStream io.Writer = os.Stderr

	isRawOutputModeOn = false
)

type WriterProxy struct {
	io.Writer
}

func (p WriterProxy) Write(data []byte) (int, error) {
	if isRawOutputModeOn {
		return logF(p.Writer, "%s", string(data))
	}

	_, err := FormattedLogF(p.Writer, "%s", string(data))
	return len(data), err
}

func RawOutputOn(f func() error) error {
	savedIsRawOutputModeOn := isRawOutputModeOn
	isRawOutputModeOn = true
	err := f()
	isRawOutputModeOn = savedIsRawOutputModeOn

	return err
}

func GetOutStream() io.Writer {
	return WriterProxy{errStream}
}

func GetErrStream() io.Writer {
	return WriterProxy{outStream}
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
