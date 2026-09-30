package instruction

import (
	"bytes"
	"context"
	"errors"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/logboek"
	"github.com/werf/werf/v3/pkg/buildah"
)

var _ = ginkgo.Describe("Copy source lifecycle", func() {
	ginkgo.DescribeTable("releases the source and preserves the operation result",
		func(failure string, expected []string) {
			backend := &copyBackendStub{failure: failure}
			var output bytes.Buffer
			logger := logboek.NewLogger(&output, &output)
			logger.Streams().SetWidth(1000)
			ctx, cancel := context.WithCancel(logboek.NewContext(context.Background(), logger))
			defer cancel()
			backend.cancel = cancel
			instruction := NewCopy(instructions.CopyCommand{From: "source-image"})
			err := instruction.Apply(ctx, "destination", backend, buildah.CommonOpts{}, &copyArchiveStub{})
			gomega.Expect(backend.calls).To(gomega.Equal(expected))
			if failure == "" || failure == "unmount" || failure == "remove" {
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			} else {
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(errors.Is(err, backend.err)).To(gomega.BeTrue())
			}
			if failure == "unmount" || failure == "remove" {
				gomega.Expect(output.String()).To(gomega.ContainSubstring(`COPY --from="source-image"`))
				gomega.Expect(output.String()).To(gomega.ContainSubstring(`source container "source-container"`))
				gomega.Expect(output.String()).To(gomega.ContainSubstring(failure + " failed"))
			} else {
				gomega.Expect(output.String()).To(gomega.BeEmpty())
			}
		},
		ginkgo.Entry("success", "", []string{"from", "mount", "copy", "unmount", "remove"}),
		ginkgo.Entry("create error", "from", []string{"from"}),
		ginkgo.Entry("save error after creation", "save", []string{"from", "remove"}),
		ginkgo.Entry("mount error", "mount", []string{"from", "mount", "remove"}),
		ginkgo.Entry("copy error", "copy", []string{"from", "mount", "copy", "unmount", "remove"}),
		ginkgo.Entry("canceled copy", "cancel", []string{"from", "mount", "copy", "unmount", "remove"}),
		ginkgo.Entry("unmount error is logged and still removes", "unmount", []string{"from", "mount", "copy", "unmount", "remove"}),
		ginkgo.Entry("remove error is logged", "remove", []string{"from", "mount", "copy", "unmount", "remove"}),
	)
	ginkgo.It("preserves the copy error while reporting a cleanup failure", func() {
		backend := &copyBackendStub{failure: "copy", cleanupFailure: true}
		var output bytes.Buffer
		logger := logboek.NewLogger(&output, &output)
		logger.Streams().SetWidth(1000)
		ctx := logboek.NewContext(context.Background(), logger)
		err := NewCopy(instructions.CopyCommand{From: "source-image"}).Apply(ctx, "destination", backend, buildah.CommonOpts{}, &copyArchiveStub{})
		gomega.Expect(errors.Is(err, backend.err)).To(gomega.BeTrue())
		gomega.Expect(err.Error()).NotTo(gomega.ContainSubstring("remove failed"))
		gomega.Expect(output.String()).To(gomega.ContainSubstring("remove failed"))
		gomega.Expect(output.String()).To(gomega.ContainSubstring(`COPY --from="source-image"`))
	})
	ginkgo.It("releases the source when copy panics without swallowing the panic", func() {
		backend := &copyBackendStub{failure: "panic"}
		gomega.Expect(func() {
			NewCopy(instructions.CopyCommand{From: "source-image"}).Apply(context.Background(), "destination", backend, buildah.CommonOpts{}, &copyArchiveStub{})
		}).To(gomega.PanicWith("copy panicked"))
		gomega.Expect(backend.calls).To(gomega.Equal([]string{"from", "mount", "copy", "unmount", "remove"}))
	})
	ginkgo.It("does not create or clean a container for build-context copies", func() {
		backend := &copyBackendStub{}
		instruction := NewCopy(instructions.CopyCommand{})
		gomega.Expect(instruction.Apply(context.Background(), "destination", backend, buildah.CommonOpts{}, &copyArchiveStub{})).To(gomega.Succeed())
		gomega.Expect(backend.calls).To(gomega.Equal([]string{"copy"}))
	})
})
