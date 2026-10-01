package lock_manager

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/storage"
)

var _ = ginkgo.Describe("HTTP synchronization production boundary", func() {
	ginkgo.DescribeTable("bounds actual requests to ten seconds", func(ctx ginkgo.SpecContext, registered bool) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-r.Context().Done():
			case <-time.After(12 * time.Second):
			}
		}))
		defer srv.Close()
		repo := &clientRecordStorage{}
		if registered {
			repo.records = []*storage.ClientIDRecord{{ClientID: "client"}}
		}
		start := time.Now()
		synchronization, err := NewHttpSynchronization(ctx, SynchronizationParams{ProjectName: "project", ServerAddress: srv.URL, StagesStorage: repo, AllowFallback: true})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		manager, err := synchronization.GetStorageLockManager(ctx)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		handle, err := manager.LockStage(ctx, "project", "stage")
		elapsed := time.Since(start)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(elapsed).To(gomega.BeNumerically(">=", 9*time.Second))
		gomega.Expect(elapsed).To(gomega.BeNumerically("<", 11500*time.Millisecond))
		gomega.Expect(handle.skipped).To(gomega.BeTrue())
		gomega.Expect(manager.Unlock(ctx, handle)).To(gomega.Succeed())
	}, ginkgo.Entry("bootstrap", false), ginkgo.Entry("acquisition", true))

	ginkgo.DescribeTable("keeps known authentication failures strict with truncated bodies", func(ctx ginkgo.SpecContext, status int, registered bool) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Length", "100")
			w.WriteHeader(status)
			if _, err := io.WriteString(w, "denied"); err != nil {
				return
			}
		}))
		defer srv.Close()
		repo := &clientRecordStorage{}
		if registered {
			repo.records = []*storage.ClientIDRecord{{ClientID: "client"}}
		}
		synchronization, err := NewHttpSynchronization(ctx, SynchronizationParams{ProjectName: "project", ServerAddress: srv.URL, StagesStorage: repo, AllowFallback: true})
		if registered {
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			manager, managerErr := synchronization.GetStorageLockManager(ctx)
			gomega.Expect(managerErr).NotTo(gomega.HaveOccurred())
			_, err = manager.LockStage(ctx, "project", "stage")
			gomega.Expect(synchronization.fallback.disabled.Load()).To(gomega.BeFalse())
		}
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring(http.StatusText(status)))
		var unavailable *synchronizationUnavailableError
		gomega.Expect(errors.As(err, &unavailable)).To(gomega.BeFalse())
	},
		ginkgo.Entry("bootstrap unauthorized", http.StatusUnauthorized, false),
		ginkgo.Entry("bootstrap forbidden", http.StatusForbidden, false),
		ginkgo.Entry("acquire unauthorized", http.StatusUnauthorized, true),
		ginkgo.Entry("acquire forbidden", http.StatusForbidden, true))
})
