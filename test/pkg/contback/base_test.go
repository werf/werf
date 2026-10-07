package contback

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

type cleanupBackendStub struct {
	ContainerBackend
	created string
	removed string
}

var _ ContainerBackend = (*cleanupBackendStub)(nil)

func (backend *cleanupBackendStub) RunSleepingContainer(_ context.Context, name, _ string) {
	backend.created = name
}

func (backend *cleanupBackendStub) Exec(context.Context, string, ...string) {
	panic("content check failed")
}

func (backend *cleanupBackendStub) Rm(ctx context.Context, name string) {
	gomega.Expect(ctx.Err()).NotTo(gomega.HaveOccurred())
	backend.removed = name
}

var _ = ginkgo.Describe("Content container cleanup", func() {
	ginkgo.It("removes the container after a failed check with a fresh context", func() {
		backend := &cleanupBackendStub{}
		ginkgo.DeferCleanup(func() {
			gomega.Expect(backend.created).NotTo(gomega.BeEmpty())
			gomega.Expect(backend.removed).To(gomega.Equal(backend.created))
		})
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		gomega.Expect(func() {
			expectCmdsToSucceed(ctx, backend, "test-image", "false")
		}).To(gomega.PanicWith("content check failed"))
		cancel()
	})
})
