package parallel_test

import (
	"bytes"
	"fmt"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/util/parallel"
	"github.com/werf/werf/v2/pkg/werf"
)

var _ = DescribeTable(
	"worker should return writing error if it was half-closed",
	func(doHalfClose, doClose bool) {
		Expect(werf.Init(GinkgoT().TempDir(), "")).To(Succeed())

		worker, err := parallel.NewWorker(1)
		Expect(err).To(Succeed())

		defer func() {
			Expect(worker.Cleanup()).To(Succeed())
		}()

		data := []byte("hello")
		reader := bytes.NewReader(data)

		if doHalfClose {
			Expect(worker.HalfClose()).To(Succeed())
		}

		if doClose {
			Expect(worker.Close()).To(Succeed()) // half-close implicitly
		}

		offset, err := io.Copy(worker, reader)
		Expect(err).To(MatchError(fmt.Errorf("worker is half closed but tries to write: %s", data)))
		Expect(offset).To(Equal(int64(0)))
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
