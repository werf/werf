//go:build !windows

package build

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/opstats"
)

var _ = Describe("createBuildReport operations flush while writing", func() {
	It("keeps the observations recorded while the report file was being written", func() {
		collector := opstats.NewCollector()
		ctx := opstats.NewContext(context.Background(), collector)

		opstats.Observe(ctx, opstats.OperationStageBuild)()
		opstats.CountEvent(ctx, opstats.EventStageBuilt)

		reportPath := filepath.Join(GinkgoT().TempDir(), "report.json")
		Expect(syscall.Mkfifo(reportPath, 0o600)).To(Succeed())
		phase := newReportPhase(reportPath)
		// A report larger than the pipe buffer stalls the write until the reader
		// drains it, so the observations below are made after the report was
		// snapshotted and before the write completed, with no timing involved.
		for i := range 2000 {
			phase.ImagesReport.SetImageRecord(strconv.Itoa(i), ReportImageRecord{WerfImageName: strings.Repeat("x", 100)})
		}

		written := make(chan []byte, 1)
		go func() {
			defer GinkgoRecover()

			reader, err := os.Open(reportPath)
			Expect(err).NotTo(HaveOccurred())
			defer reader.Close()

			opstats.Observe(ctx, opstats.OperationImagePull)()
			opstats.CountEvent(ctx, opstats.EventStageBuilt)

			data, err := io.ReadAll(reader)
			Expect(err).NotTo(HaveOccurred())
			written <- data
		}()

		Expect(createBuildReport(ctx, phase, nil)).To(Succeed())

		decoded := decodeOperationsReport(<-written)
		Expect(decoded.Operations).To(HaveKey(string(opstats.OperationStageBuild)))
		Expect(decoded.Operations).NotTo(HaveKey(string(opstats.OperationImagePull)))
		Expect(decoded.StageCache).To(Equal(map[string]int{string(opstats.EventStageBuilt): 1}))

		pending := collector.PendingSummary(ctx)
		Expect(pending).To(HaveLen(1))
		Expect(pending[0].Operation).To(Equal(opstats.OperationImagePull))
		Expect(collector.PendingEventSummary(ctx)).To(Equal([]opstats.EventSummary{{Event: opstats.EventStageBuilt, Count: 1}}))
	})
})
