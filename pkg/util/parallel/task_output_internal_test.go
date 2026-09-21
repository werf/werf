package parallel

import (
	"io"
	"os"
	"runtime"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/werf"
)

// expectClosed checks the handle itself, not the field it was stored in: a
// mutation that nils the field but forgets Close() must still be caught.
func expectClosed(file *os.File, what string) {
	GinkgoHelper()
	_, err := file.Stat()
	Expect(err).To(MatchError(os.ErrClosed), "%s descriptor must be closed", what)
}

var _ = Describe("TaskOutput descriptor lifecycle", func() {
	BeforeEach(func() {
		Expect(werf.Init(GinkgoT().TempDir(), "")).To(Succeed())
	})

	It("touches the disk only once something is written", func() {
		// Most tasks of a cleanup run log nothing at all; a file per such
		// task would put one inode per task in the tmp dir for as long as
		// the queue head holds them back.
		out := NewTaskOutput(0, 0)
		defer func() {
			Expect(out.Cleanup()).To(Succeed())
		}()

		Expect(out.path).To(BeEmpty())
		Expect(out.Readable()).To(BeTrue(), "the task is still running, output may still come")

		n, err := out.Read(make([]byte, 8))
		Expect(n).To(BeZero())
		Expect(err).To(MatchError(io.EOF))
		Expect(out.path).To(BeEmpty(), "reading an output nothing was written to must not create one either")

		Expect(out.HalfClose()).To(Succeed())
		Expect(out.Readable()).To(BeFalse())
		Expect(out.path).To(BeEmpty())
	})

	It("holds no descriptor between finishing and being drained, and none after the drain", func() {
		out := NewTaskOutput(0, 0)
		defer func() {
			Expect(out.Cleanup()).To(Succeed())
		}()

		_, err := out.Write([]byte("hello\n"))
		Expect(err).To(Succeed())
		writer := out.writer
		Expect(writer).NotTo(BeNil())
		Expect(out.reader).To(BeNil())

		Expect(out.HalfClose()).To(Succeed())
		Expect(out.writer).To(BeNil(), "writer descriptor must be released when the task finishes")
		expectClosed(writer, "writer")
		Expect(out.reader).To(BeNil(), "no reader descriptor until the printer gets to this task")

		buf := make([]byte, 3)
		n, err := out.Read(buf)
		Expect(err).To(Succeed())
		Expect(string(buf[:n])).To(Equal("hel"))
		reader := out.reader
		Expect(reader).NotTo(BeNil(), "reader stays open until the drain completes")

		content, err := io.ReadAll(out)
		Expect(err).To(Succeed())
		Expect(string(content)).To(Equal("lo\n"))
		Expect(out.reader).To(BeNil(), "reader descriptor must be released once everything is drained")
		expectClosed(reader, "reader")
		Expect(out.Readable()).To(BeFalse())
	})

	It("keeps the reader open while the task is still writing and releases it with the last read", func() {
		out := NewTaskOutput(0, 1)
		defer func() {
			Expect(out.Cleanup()).To(Succeed())
		}()

		_, err := out.Write([]byte("first\n"))
		Expect(err).To(Succeed())

		buf := make([]byte, 64)
		n, err := out.Read(buf)
		Expect(err).To(Or(Succeed(), MatchError(io.EOF)))
		Expect(string(buf[:n])).To(Equal("first\n"))
		Expect(out.reader).NotTo(BeNil(), "reader stays open: more output may still come")

		_, err = out.Write([]byte("second\n"))
		Expect(err).To(Succeed())
		Expect(out.HalfClose()).To(Succeed())
		reader := out.reader
		Expect(reader).NotTo(BeNil(), "unread output remains, the reader stays")

		n, err = out.Read(buf)
		Expect(err).To(Or(Succeed(), MatchError(io.EOF)))
		Expect(string(buf[:n])).To(Equal("second\n"))
		Expect(out.reader).To(BeNil())
		expectClosed(reader, "reader")
		Expect(out.Readable()).To(BeFalse())
	})

	It("releases a reader that already drained everything when the task then finishes", func() {
		// The printer streams a live task and can reach the end of its output
		// before the task returns; Readable() then flips to false at HalfClose
		// and no further Read happens, so HalfClose itself has to let go of
		// the reader.
		out := NewTaskOutput(0, 2)
		defer func() {
			Expect(out.Cleanup()).To(Succeed())
		}()

		_, err := out.Write([]byte("all of it\n"))
		Expect(err).To(Succeed())

		buf := make([]byte, 64)
		n, err := out.Read(buf)
		Expect(err).To(Or(Succeed(), MatchError(io.EOF)))
		Expect(string(buf[:n])).To(Equal("all of it\n"))
		reader := out.reader
		Expect(reader).NotTo(BeNil())

		Expect(out.HalfClose()).To(Succeed())
		Expect(out.Readable()).To(BeFalse())
		Expect(out.reader).To(BeNil(), "reader must not outlive a task whose output was fully read before it finished")
		expectClosed(reader, "reader")
	})

	It("removes the file once and treats a second Cleanup as a no-op", func() {
		// The Printer removes a printed file early; the final sweep in
		// runWorkers must be able to call Cleanup on it again without error.
		out := NewTaskOutput(0, 3)
		_, err := out.Write([]byte("something, so that there is a file to remove\n"))
		Expect(err).To(Succeed())

		Expect(out.Cleanup()).To(MatchError(ContainSubstring("not half closed yet")))
		Expect(out.HalfClose()).To(Succeed())

		Expect(out.Cleanup()).To(Succeed())
		_, err = os.Stat(out.path)
		Expect(err).To(MatchError(os.ErrNotExist))
		Expect(out.Cleanup()).To(Succeed())
	})
})

var _ = DescribeTable("TaskOutput.HalfClose leaves the output on a line boundary",
	func(writes []string, expected string) {
		Expect(werf.Init(GinkgoT().TempDir(), "")).To(Succeed())

		out := NewTaskOutput(0, 0)
		defer func() {
			Expect(out.Cleanup()).To(Succeed())
		}()

		var err error
		for _, w := range writes {
			_, err = out.Write([]byte(w))
			Expect(err).To(Succeed())
		}
		Expect(out.HalfClose()).To(Succeed())

		content, err := io.ReadAll(out)
		Expect(err).To(Succeed())
		Expect(string(content)).To(Equal(expected))
	},
	Entry("nothing written stays empty", nil, ""),
	Entry("a terminated line is left alone", []string{"a\n"}, "a\n"),
	Entry("a line written after the logger was flushed, right before half-close, is terminated", []string{"a\n", "late"}, "a\nlate\n"),
)

var _ = It("TaskOutput.HalfClose keeps the line boundary against a writer racing with it", func() {
	// Terminating the last line and stopping writes must happen in one
	// critical section: a write that lands between the two would follow the
	// terminator and leave the block unfinished again. A goroutine hammering
	// Write while HalfClose runs would slip into any gap in that section.
	Expect(werf.Init(GinkgoT().TempDir(), "")).To(Succeed())

	for i := 0; i < 200; i++ {
		halfCloseAgainstRacingWriter(i)
	}
})

// halfCloseAgainstRacingWriter is one iteration of the race above; the
// writer goroutine is joined and the buffer released in defers, so a failed
// assertion does not leave a goroutine writing into a file nobody removes.
func halfCloseAgainstRacingWriter(iteration int) {
	GinkgoHelper()

	out := NewTaskOutput(0, iteration)
	defer func() {
		Expect(out.Close()).To(Succeed())
		Expect(out.Cleanup()).To(Succeed())
	}()

	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer GinkgoRecover()
		defer close(done)
		for {
			select {
			case <-stop:
				return
			default:
				_, err := out.Write([]byte("x"))
				Expect(err).To(Succeed())
			}
		}
	}()
	defer func() {
		close(stop)
		<-done
	}()

	_, err := out.Write([]byte("a"))
	Expect(err).To(Succeed())
	runtime.Gosched()
	Expect(out.HalfClose()).To(Succeed())

	content, err := io.ReadAll(out)
	Expect(err).To(Succeed())
	Expect(string(content)).To(HaveSuffix("\n"), "iteration %d: a write slipped in between the terminator and the close", iteration)
	Expect(strings.Count(string(content), "\n")).To(Equal(1), "iteration %d: exactly one terminator", iteration)
}
