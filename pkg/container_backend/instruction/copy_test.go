package instruction

import (
	"context"
	"errors"
	"testing"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/buildah"
	"github.com/werf/werf/v3/pkg/container_backend"
)

func TestCopyCleanup(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Copy source cleanup")
}

var _ = ginkgo.Describe("Copy source lifecycle", func() {
	ginkgo.DescribeTable("releases the source on every returned path",
		func(failure string, expected []string) {
			backend := &copyBackendStub{failure: failure}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			backend.cancel = cancel
			instruction := NewCopy(instructions.CopyCommand{From: "source-image"})
			err := instruction.Apply(ctx, "destination", backend, buildah.CommonOpts{}, &copyArchiveStub{})
			gomega.Expect(backend.calls).To(gomega.Equal(expected))
			if failure == "" {
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			} else {
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(errors.Is(err, backend.err)).To(gomega.BeTrue())
			}
		},
		ginkgo.Entry("success", "", []string{"from", "mount", "copy", "unmount", "remove"}),
		ginkgo.Entry("create error", "from", []string{"from"}),
		ginkgo.Entry("save error after creation", "save", []string{"from", "remove"}),
		ginkgo.Entry("mount error", "mount", []string{"from", "mount", "remove"}),
		ginkgo.Entry("copy error", "copy", []string{"from", "mount", "copy", "unmount", "remove"}),
		ginkgo.Entry("canceled copy", "cancel", []string{"from", "mount", "copy", "unmount", "remove"}),
		ginkgo.Entry("unmount error still removes", "unmount", []string{"from", "mount", "copy", "unmount", "remove"}),
		ginkgo.Entry("remove error", "remove", []string{"from", "mount", "copy", "unmount", "remove"}),
	)
	ginkgo.It("does not create or clean a container for build-context copies", func() {
		backend := &copyBackendStub{}
		instruction := NewCopy(instructions.CopyCommand{})
		gomega.Expect(instruction.Apply(context.Background(), "destination", backend, buildah.CommonOpts{}, &copyArchiveStub{})).To(gomega.Succeed())
		gomega.Expect(backend.calls).To(gomega.Equal([]string{"copy"}))
	})
})

type copyBackendStub struct {
	buildah.Buildah
	failure string
	err     error
	calls   []string
	cancel  context.CancelFunc
}

var _ buildah.Buildah = (*copyBackendStub)(nil)

func (backend *copyBackendStub) record(operation string) error {
	backend.calls = append(backend.calls, operation)
	if backend.failure == operation {
		backend.err = errors.New(operation + " failed")
		return backend.err
	}
	return nil
}

func (backend *copyBackendStub) FromCommand(_ context.Context, name, image string, _ buildah.FromCommandOpts) (string, error) {
	gomega.Expect(name).To(gomega.BeEmpty())
	gomega.Expect(image).To(gomega.Equal("source-image"))
	if err := backend.record("from"); err != nil {
		return "", err
	}
	if backend.failure == "save" {
		backend.err = errors.New("save failed")
		return "source-container", backend.err
	}
	return "source-container", nil
}

func (backend *copyBackendStub) Mount(_ context.Context, name string, _ buildah.MountOpts) (string, error) {
	gomega.Expect(name).To(gomega.Equal("source-container"))
	return "/source", backend.record("mount")
}

func (backend *copyBackendStub) Copy(ctx context.Context, name, dir string, _ []string, _ string, _ buildah.CopyOpts) error {
	gomega.Expect(name).To(gomega.Equal("destination"))
	gomega.Expect(dir).To(gomega.Equal("/source"))
	err := backend.record("copy")
	if backend.failure == "cancel" {
		backend.cancel()
		backend.err = ctx.Err()
		return backend.err
	}
	return err
}

func (backend *copyBackendStub) Umount(ctx context.Context, name string, _ buildah.UmountOpts) error {
	gomega.Expect(ctx.Err()).NotTo(gomega.HaveOccurred())
	gomega.Expect(name).To(gomega.Equal("source-container"))
	return backend.record("unmount")
}

func (backend *copyBackendStub) Rm(ctx context.Context, name string, _ buildah.RmOpts) error {
	gomega.Expect(ctx.Err()).NotTo(gomega.HaveOccurred())
	gomega.Expect(name).To(gomega.Equal("source-container"))
	return backend.record("remove")
}

type copyArchiveStub struct {
	container_backend.BuildContextArchiver
}

var _ container_backend.BuildContextArchiver = (*copyArchiveStub)(nil)

func (*copyArchiveStub) ExtractOrGetExtractedDir(context.Context) (string, error) {
	return "/source", nil
}
