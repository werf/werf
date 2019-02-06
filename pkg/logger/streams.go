package logger

import (
	"io"
	"io/ioutil"
	"os"
)

var (
	outStream io.Writer = os.Stdout
	errStream io.Writer = os.Stderr
)

type WriterProxy struct {
	io.Writer
}

func (p WriterProxy) Write(data []byte) (n int, err error) {
	return FormattedLogF(p.Writer, "", string(data))
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
