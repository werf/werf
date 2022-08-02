package util

import (
	"io"

	"github.com/djherbis/buffer"
	"github.com/djherbis/nio/v3"
)

func BufferedPipedWriterProcess(f func(w io.WriteCloser)) io.ReadCloser {
	buf := buffer.New(64 * 1024 * 1024)
	r, w := nio.Pipe(buf)
	go f(w)
	return r
}
