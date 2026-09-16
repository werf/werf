package parallel

import (
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/werf"
)

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
		Expect(out.writer).NotTo(BeNil())
		Expect(out.reader).To(BeNil())

		Expect(out.HalfClose()).To(Succeed())
		Expect(out.writer).To(BeNil(), "writer descriptor must be released when the task finishes")
		Expect(out.reader).To(BeNil(), "no reader descriptor until the printer gets to this task")

		content, err := io.ReadAll(out)
		Expect(err).To(Succeed())
		Expect(string(content)).To(Equal("hello\n"))
		Expect(out.reader).To(BeNil(), "reader descriptor must be released once everything is drained")
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

		n, err = out.Read(buf)
		Expect(err).To(Or(Succeed(), MatchError(io.EOF)))
		Expect(string(buf[:n])).To(Equal("second\n"))
		Expect(out.reader).To(BeNil())
		Expect(out.Readable()).To(BeFalse())
	})
})
