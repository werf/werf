package parallel_test

import (
	"bytes"
	"io"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/util/parallel"
	"github.com/werf/werf/v2/pkg/werf"
)

// runeSplittingWriter mimics logboek's internal proxy-formatting layer
// (github.com/werf/logboek/internal/logger/manager.go's proxyStream.Write ->
// internal/stream/stream.go's Stream.FormatAndLogF), which decodes each
// Write() call's byte slice into runes via []rune(string(p)). If a Write()
// carries a truncated multi-byte UTF-8 sequence, decoding replaces the
// invalid bytes with the Unicode replacement character (U+FFFD) - mojibake.
type runeSplittingWriter struct {
	buf bytes.Buffer
}

func (w *runeSplittingWriter) Write(p []byte) (int, error) {
	for _, r := range []rune(string(p)) {
		w.buf.WriteRune(r)
	}
	return len(p), nil
}

var _ = Describe("TaskOutput.Read UTF-8 boundary safety", func() {
	It("never splits a multi-byte rune across two reads while the task is still writing", func() {
		Expect(werf.Init(GinkgoT().TempDir(), "")).To(Succeed())

		out, err := parallel.NewTaskOutput(1, 0)
		Expect(err).To(Succeed())

		// Padding chosen so the 1024-byte read boundary lands in the middle
		// of the "│" character's 3-byte UTF-8 encoding (E2 94 82), matching
		// the reported failure mode.
		line := strings.Repeat("a", 1019) + "│ └ done\n"
		_, err = out.Write([]byte(line))
		Expect(err).To(Succeed())

		// Read while the output is still open (not half-closed), so the
		// boundary-holdback logic under test is actually exercised.
		sink := &runeSplittingWriter{}
		readBuf := make([]byte, 1024) // matches printer.go's read-buffer size
		n, readErr := out.Read(readBuf)
		Expect(readErr).To(Succeed())
		_, err = sink.Write(readBuf[:n])
		Expect(err).To(Succeed())

		out.HalfClose()

		for out.Readable() {
			n, readErr = out.Read(readBuf)
			if n > 0 {
				_, err = sink.Write(readBuf[:n])
				Expect(err).To(Succeed())
			}
			Expect(readErr).To(Or(Succeed(), MatchError(io.EOF)))
		}

		Expect(sink.buf.String()).To(Equal(line))
		Expect(sink.buf.String()).NotTo(ContainSubstring("\uFFFD"))

		Expect(out.Cleanup()).To(Succeed())
	})

	It("never splits a multi-byte rune across two reads even when the task half-closed before printing started", func() {
		Expect(werf.Init(GinkgoT().TempDir(), "")).To(Succeed())

		out, err := parallel.NewTaskOutput(3, 0)
		Expect(err).To(Succeed())

		// Same boundary alignment as the previous test, but this time the
		// output is half-closed BEFORE any read happens - matching the
		// real-world case of a fast task finishing before the printer starts
		// draining its temp file.
		line := strings.Repeat("a", 1019) + "│ └ done\n"
		_, err = out.Write([]byte(line))
		Expect(err).To(Succeed())
		out.HalfClose()

		sink := &runeSplittingWriter{}
		readBuf := make([]byte, 1024)
		for out.Readable() {
			n, readErr := out.Read(readBuf)
			if n > 0 {
				_, err = sink.Write(readBuf[:n])
				Expect(err).To(Succeed())
			}
			Expect(readErr).To(Or(Succeed(), MatchError(io.EOF)))
		}

		Expect(sink.buf.String()).To(Equal(line))
		Expect(sink.buf.String()).NotTo(ContainSubstring("\uFFFD"))

		Expect(out.Cleanup()).To(Succeed())
	})

	It("flushes a trailing incomplete rune once the task is half-closed, without hanging", func() {
		Expect(werf.Init(GinkgoT().TempDir(), "")).To(Succeed())

		out, err := parallel.NewTaskOutput(2, 0)
		Expect(err).To(Succeed())

		_, err = out.Write([]byte("└"))
		Expect(err).To(Succeed())
		out.HalfClose()

		var result bytes.Buffer
		readBuf := make([]byte, 2)
		for out.Readable() {
			n, readErr := out.Read(readBuf)
			if n > 0 {
				result.Write(readBuf[:n])
			}
			Expect(readErr).To(Or(Succeed(), MatchError(io.EOF)))
		}

		Expect(result.String()).To(Equal("└"))

		Expect(out.Cleanup()).To(Succeed())
	})
})
