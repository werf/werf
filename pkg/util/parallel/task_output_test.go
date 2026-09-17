package parallel_test

import (
	"bytes"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/util/parallel"
	"github.com/werf/werf/v2/pkg/werf"
)

var _ = DescribeTable(
	"task output should discard writes after half-close",
	func(doHalfClose, doClose bool) {
		Expect(werf.Init(GinkgoT().TempDir(), "")).To(Succeed())

		out, err := parallel.NewTaskOutput(1, 0)
		Expect(err).To(Succeed())

		defer func() {
			Expect(out.Cleanup()).To(Succeed())
		}()

		data := []byte("hello")
		reader := bytes.NewReader(data)

		if doHalfClose {
			Expect(out.HalfClose()).To(Succeed())
		}

		if doClose {
			Expect(out.Close()).To(Succeed()) // half-close implicitly
		}

		offset, err := io.Copy(out, reader)
		Expect(err).NotTo(HaveOccurred())
		Expect(offset).To(Equal(int64(len(data))))

		if doHalfClose {
			content, err := io.ReadAll(out)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(BeEmpty())
		}
	},
	Entry(
		"half-close explicitly",
		true,
		false,
	),
	Entry(
		"half-close implicitly via close",
		false,
		true,
	),
)
