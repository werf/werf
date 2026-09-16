package parallel

import (
	"io"
	"os"

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

	It("holds no descriptor between finishing and being drained, and none after the drain", func() {
		out, err := NewTaskOutput(0, 0)
		Expect(err).To(Succeed())
		defer func() {
			Expect(out.Cleanup()).To(Succeed())
		}()

		_, err = out.Write([]byte("hello\n"))
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
		out, err := NewTaskOutput(0, 1)
		Expect(err).To(Succeed())
		defer func() {
			Expect(out.Cleanup()).To(Succeed())
		}()

		_, err = out.Write([]byte("first\n"))
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
		out, err := NewTaskOutput(0, 2)
		Expect(err).To(Succeed())
		defer func() {
			Expect(out.Cleanup()).To(Succeed())
		}()

		_, err = out.Write([]byte("all of it\n"))
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
})
