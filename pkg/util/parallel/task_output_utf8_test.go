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

		out := parallel.NewTaskOutput(1, 0)

		// Padding chosen so the 1024-byte read boundary lands in the middle
		// of the "│" character's 3-byte UTF-8 encoding (E2 94 82), matching
		// the reported failure mode.
		line := strings.Repeat("a", 1019) + "│ └ done\n"
		_, err := out.Write([]byte(line))
		Expect(err).To(Succeed())

		// Read while the output is still open (not half-closed), so the
		// boundary-holdback logic under test is actually exercised.
		sink := &runeSplittingWriter{}
		readBuf := make([]byte, 1024) // matches printer.go's read-buffer size
		n, readErr := out.Read(readBuf)
		Expect(readErr).To(Succeed())
		_, err = sink.Write(readBuf[:n])
		Expect(err).To(Succeed())

		Expect(out.HalfClose()).To(Succeed())

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

		out := parallel.NewTaskOutput(3, 0)

		// Same boundary alignment as the previous test, but this time the
		// output is half-closed BEFORE any read happens - matching the
		// real-world case of a fast task finishing before the printer starts
		// draining its temp file.
		line := strings.Repeat("a", 1019) + "│ └ done\n"
		_, err := out.Write([]byte(line))
		Expect(err).To(Succeed())
		Expect(out.HalfClose()).To(Succeed())

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

	It("holds back a read that finds nothing but the bytes of a truncated rune", func() {
		Expect(werf.Init(GinkgoT().TempDir(), "")).To(Succeed())

		out := parallel.NewTaskOutput(4, 0)

		// The printer drains faster than the task writes: the first read stops
		// at the rune boundary, and the second finds only the truncated tail.
		_, err := out.Write([]byte("a\xe2\x94"))
		Expect(err).To(Succeed())

		readBuf := make([]byte, 1024)
		n, err := out.Read(readBuf)
		Expect(err).To(Succeed())
		Expect(string(readBuf[:n])).To(Equal("a"))

		n, err = out.Read(readBuf)
		Expect(n).To(BeZero(), "a fragment of a rune must not reach the logger")
		Expect(err).To(MatchError(io.EOF), "the printer must be parked, not spun on a 0, nil read")

		_, err = out.Write([]byte("\x82 done\n"))
		Expect(err).To(Succeed())
		Expect(out.HalfClose()).To(Succeed())

		sink := &runeSplittingWriter{}
		for out.Readable() {
			n, readErr := out.Read(readBuf)
			if n > 0 {
				_, err = sink.Write(readBuf[:n])
				Expect(err).To(Succeed())
			}
			Expect(readErr).To(Or(Succeed(), MatchError(io.EOF)))
		}

		Expect(sink.buf.String()).To(Equal("│ done\n"))
		Expect(sink.buf.String()).NotTo(ContainSubstring("\uFFFD"))

		Expect(out.Cleanup()).To(Succeed())
	})

	It("gives up the holdback for a buffer too small to hold one rune, rather than stalling the drain", func() {
		Expect(werf.Init(GinkgoT().TempDir(), "")).To(Succeed())

		out := parallel.NewTaskOutput(5, 0)

		_, err := out.Write([]byte("│\n"))
		Expect(err).To(Succeed())

		// Two bytes can never carry the three of "│": holding the fragment
		// back here would leave the task readable forever.
		var result bytes.Buffer
		readBuf := make([]byte, 2)
		for result.Len() < len("│\n") {
			n, readErr := out.Read(readBuf)
			Expect(readErr).To(Or(Succeed(), MatchError(io.EOF)))
			Expect(n).NotTo(BeZero(), "the drain must keep advancing")
			result.Write(readBuf[:n])
		}
		Expect(result.String()).To(Equal("│\n"))

		Expect(out.HalfClose()).To(Succeed())
		Expect(out.Cleanup()).To(Succeed())
	})

	It("flushes a trailing incomplete rune once the task is half-closed, without hanging", func() {
		Expect(werf.Init(GinkgoT().TempDir(), "")).To(Succeed())

		out := parallel.NewTaskOutput(2, 0)

		_, err := out.Write([]byte("└"))
		Expect(err).To(Succeed())
		Expect(out.HalfClose()).To(Succeed())

		var result bytes.Buffer
		readBuf := make([]byte, 2)
		for out.Readable() {
			n, readErr := out.Read(readBuf)
			if n > 0 {
				result.Write(readBuf[:n])
			}
			Expect(readErr).To(Or(Succeed(), MatchError(io.EOF)))
		}

		// HalfClose terminates the unfinished line; the rune itself must
		// arrive intact in front of that newline.
		Expect(result.String()).To(Equal("└\n"))

		Expect(out.Cleanup()).To(Succeed())
	})
})
