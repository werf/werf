package parallel_test

import (
	"bytes"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/util/parallel"
	"github.com/werf/werf/v2/pkg/werf"
)

var _ = Describe("worker", func() {
	var worker *parallel.Worker

	BeforeEach(func() {
		// tmp_manager requires werf init
		Expect(werf.Init(GinkgoT().TempDir(), "")).To(Succeed())

		var err error
		worker, err = parallel.NewWorker(1)
		Expect(err).To(Succeed())
	})
	AfterEach(func() {
		Expect(worker.Cleanup()).To(Succeed())
	})

	It("should discard writes into underlying file if it was half-closed (or closed)", func() {
		Expect(worker.HalfClose()).To(Succeed())
		Expect(worker.Close()).To(Succeed())

		reader := bytes.NewReader([]byte("hello"))
		offset, err := io.Copy(worker, reader)
		Expect(err).To(Succeed())
		Expect(offset).To(Equal(reader.Size()))
	})
})
