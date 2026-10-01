package lock_manager

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/lockgate"
	"github.com/werf/lockgate/pkg/distributed_locker"
	"github.com/werf/lockgate/pkg/distributed_locker/optimistic_locking_store"
	"github.com/werf/logboek"
	"github.com/werf/werf/v3/pkg/storage"
)

var _ = ginkgo.Describe("HTTP synchronization fallback", func() {
	ginkgo.DescribeTable("classifies only availability failures", func(err error, eligible bool) {
		classified := classifySynchronizationRequestError(context.Background(), err)
		var unavailable *synchronizationUnavailableError
		gomega.Expect(errors.As(classified, &unavailable)).To(gomega.Equal(eligible))
	},
		ginkgo.Entry("DNS", &net.DNSError{Err: "no such host", Name: "sync.example"}, true),
		ginkgo.Entry("connect", &net.OpError{Op: "dial", Err: syscall.ECONNREFUSED}, true),
		ginkgo.Entry("connection reset", &net.OpError{Op: "read", Err: syscall.ECONNRESET}, true),
		ginkgo.Entry("EOF", io.EOF, true),
		ginkgo.Entry("canceled", context.Canceled, false),
		ginkgo.Entry("certificate verification", &tls.CertificateVerificationError{Err: x509.UnknownAuthorityError{}}, false),
		ginkgo.Entry("TLS verification wrapped as dial failure", &net.OpError{Op: "dial", Err: &tls.CertificateVerificationError{Err: x509.UnknownAuthorityError{}}}, false),
		ginkgo.Entry("certificate hostname", x509.HostnameError{}, false),
		ginkgo.Entry("remote TLS alert", &net.OpError{Op: "remote error", Err: errors.New("tls alert")}, false),
		ginkgo.Entry("protocol", errors.New("invalid response"), false),
	)

	ginkgo.DescribeTable("bootstrap degrades only for allowed HTTP failures", func(ctx ginkgo.SpecContext, status int, body string, allowed, wantFallback bool) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
			_, err := io.WriteString(w, body)
			if err != nil {
				return
			}
		}))
		defer srv.Close()
		repo := &clientRecordStorage{}
		synchronization, err := NewHttpSynchronization(ctx, SynchronizationParams{ProjectName: "project", ServerAddress: srv.URL, StagesStorage: repo, AllowFallback: allowed})
		if wantFallback {
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			manager, err := synchronization.GetStorageLockManager(ctx)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			handle, err := manager.LockStage(ctx, "project", "digest")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(handle.skipped).To(gomega.BeTrue())
			gomega.Expect(manager.Unlock(ctx, handle)).To(gomega.Succeed())
		} else {
			gomega.Expect(err).To(gomega.HaveOccurred())
		}
		gomega.Expect(repo.records).To(gomega.BeEmpty())
	},
		ginkgo.Entry("implicit unavailable", 503, `{"clientID":"must-not-register"}`, true, true),
		ginkgo.Entry("explicit unavailable", 503, `{"clientID":"must-not-register"}`, false, false),
		ginkgo.Entry("unauthorized", 401, `{"clientID":"must-not-register"}`, true, false),
		ginkgo.Entry("forbidden", 403, `{"clientID":"must-not-register"}`, true, false),
		ginkgo.Entry("malformed JSON", 200, `not json`, true, false),
		ginkgo.Entry("empty ID", 200, `{}`, true, false),
	)

	ginkgo.It("does not mistake registry DNS failure for synchronization failure", func(ctx ginkgo.SpecContext) {
		expected := &net.DNSError{Err: "no such host", Name: "registry.example"}
		_, err := NewHttpSynchronization(ctx, SynchronizationParams{ProjectName: "project", ServerAddress: "http://unused", StagesStorage: &clientRecordStorage{readErr: expected}, AllowFallback: true})
		gomega.Expect(errors.Is(err, expected)).To(gomega.BeTrue())
	})

	ginkgo.It("bounds requests and keeps parent cancellation strict", func(ctx ginkgo.SpecContext) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-r.Context().Done():
			case <-time.After(200 * time.Millisecond):
			}
		}))
		defer srv.Close()
		client := &http.Client{Timeout: 20 * time.Millisecond}
		err := performPost(ctx, client, srv.URL, struct{}{}, &struct{}{})
		var unavailable *synchronizationUnavailableError
		gomega.Expect(errors.As(err, &unavailable)).To(gomega.BeTrue())
		canceled, cancel := context.WithCancel(ctx)
		cancel()
		err = performPost(canceled, client, srv.URL, struct{}{}, &struct{}{})
		gomega.Expect(errors.Is(err, context.Canceled)).To(gomega.BeTrue())
		gomega.Expect(errors.As(err, &unavailable)).To(gomega.BeFalse())
		deadline, cancel := context.WithTimeout(ctx, time.Nanosecond)
		defer cancel()
		<-deadline.Done()
		err = performPost(deadline, client, srv.URL, struct{}{}, &struct{}{})
		gomega.Expect(errors.Is(err, context.DeadlineExceeded)).To(gomega.BeTrue())
		gomega.Expect(errors.As(err, &unavailable)).To(gomega.BeFalse())
	})

	ginkgo.It("keeps authentication failures strict even when their body stalls", func(ctx ginkgo.SpecContext) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			w.(http.Flusher).Flush()
			select {
			case <-r.Context().Done():
			case <-time.After(200 * time.Millisecond):
			}
		}))
		defer srv.Close()
		err := performPost(ctx, &http.Client{Timeout: 20 * time.Millisecond}, srv.URL, struct{}{}, &struct{}{})
		gomega.Expect(err).To(gomega.HaveOccurred())
		var unavailable *synchronizationUnavailableError
		gomega.Expect(errors.As(err, &unavailable)).To(gomega.BeFalse())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("401 Unauthorized"))
	})

	ginkgo.It("retains real handles and shares sticky fallback across managers", func(ctx ginkgo.SpecContext) {
		var outage atomic.Bool
		var acquisitions, releases atomic.Int32
		backend := distributed_locker.NewOptimisticLockingStorageBasedBackend(optimistic_locking_store.NewInMemoryStore())
		handler := http.StripPrefix("/client/locker", distributed_locker.NewHttpBackendHandler(backend))
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/acquire") {
				acquisitions.Add(1)
				if outage.Load() {
					w.WriteHeader(503)
					return
				}
			}
			if strings.HasSuffix(r.URL.Path, "/release") {
				releases.Add(1)
			}
			handler.ServeHTTP(w, r)
		}))
		defer srv.Close()
		var output bytes.Buffer
		ctxLog := logboek.NewContext(ctx, logboek.NewLogger(&output, &output))
		synchronization, err := NewHttpSynchronization(ctxLog, SynchronizationParams{ProjectName: "project", ServerAddress: srv.URL, StagesStorage: &clientRecordStorage{records: []*storage.ClientIDRecord{{ClientID: "client"}}}, AllowFallback: true})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		first, err := synchronization.GetStorageLockManager(ctxLog)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		real, err := first.LockStage(ctxLog, "project", "first")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(real.skipped).To(gomega.BeFalse())
		outage.Store(true)
		second, err := synchronization.GetStorageLockManager(ctxLog)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		var workers sync.WaitGroup
		for range 8 {
			workers.Go(func() {
				defer ginkgo.GinkgoRecover()
				handle, err := second.LockStage(ctxLog, "project", "second")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(handle.skipped).To(gomega.BeTrue())
				gomega.Expect(second.Unlock(ctxLog, handle)).To(gomega.Succeed())
			})
		}
		workers.Wait()
		count := acquisitions.Load()
		outage.Store(false)
		third, err := synchronization.GetStorageLockManager(ctxLog)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		skipped, err := third.LockStage(ctxLog, "project", "third")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(skipped.skipped).To(gomega.BeTrue())
		gomega.Expect(acquisitions.Load()).To(gomega.Equal(count))
		gomega.Expect(first.Unlock(ctxLog, real)).To(gomega.Succeed())
		gomega.Expect(releases.Load()).To(gomega.Equal(int32(1)))
		gomega.Expect(strings.Count(output.String(), "Continuing without distributed publication locks")).To(gomega.Equal(1))
		canceled, cancel := context.WithCancel(ctxLog)
		cancel()
		_, err = third.LockStage(canceled, "project", "canceled")
		gomega.Expect(errors.Is(err, context.Canceled)).To(gomega.BeTrue())
	})

	ginkgo.It("keeps lease errors strict and releases after caller cancellation", func(ctx ginkgo.SpecContext) {
		var releases atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/release") {
				releases.Add(1)
				_, err := io.WriteString(w, `{}`)
				if err != nil {
					return
				}
				return
			}
			w.WriteHeader(503)
		}))
		defer srv.Close()
		canceled, cancel := context.WithCancel(ctx)
		backend := &httpBackend{ctx: canceled, endpoint: srv.URL, client: &http.Client{Timeout: synchronizationRequestTimeout}}
		handle := lockgate.LockHandle{UUID: "real", LockName: "project.digest"}
		gomega.Expect(backend.RenewLease(handle)).To(gomega.HaveOccurred())
		cancel()
		gomega.Expect(backend.Release(handle)).To(gomega.Succeed())
		gomega.Expect(releases.Load()).To(gomega.Equal(int32(1)))
	})
	ginkgo.It("does not downgrade real TLS validation failures", func(ctx ginkgo.SpecContext) {
		srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(503) }))
		defer srv.Close()
		_, err := NewHttpSynchronization(ctx, SynchronizationParams{ProjectName: "project", ServerAddress: srv.URL, StagesStorage: &clientRecordStorage{}, AllowFallback: true})
		gomega.Expect(err).To(gomega.HaveOccurred())
		var unavailable *synchronizationUnavailableError
		gomega.Expect(errors.As(err, &unavailable)).To(gomega.BeFalse())
	})

	ginkgo.It("preserves a real acquisition that completes after fallback", func(ctx ginkgo.SpecContext) {
		firstStarted := make(chan struct{})
		finishFirst := make(chan struct{})
		var acquisitions, releases atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/release") {
				releases.Add(1)
				_, err := io.WriteString(w, `{}`)
				if err != nil {
					return
				}
				return
			}
			if acquisitions.Add(1) == 1 {
				close(firstStarted)
				<-finishFirst
				err := json.NewEncoder(w).Encode(distributed_locker.AcquireResponse{LockHandle: lockgate.LockHandle{UUID: "winner", LockName: "project.first"}})
				if err != nil {
					return
				}
				return
			}
			w.WriteHeader(503)
		}))
		defer srv.Close()
		fallback := &httpFallbackState{}
		manager := newHTTPManager(ctx, srv.URL, "client", fallback)
		handles := make(chan LockHandle, 1)
		failures := make(chan error, 1)
		go func() { handle, err := manager.LockStage(ctx, "project", "first"); handles <- handle; failures <- err }()
		gomega.Eventually(firstStarted).Should(gomega.BeClosed())
		skipped, err := manager.LockStage(ctx, "project", "second")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(skipped.skipped).To(gomega.BeTrue())
		close(finishFirst)
		var real LockHandle
		gomega.Eventually(handles).Should(gomega.Receive(&real))
		gomega.Eventually(failures).Should(gomega.Receive(gomega.BeNil()))
		gomega.Expect(real.skipped).To(gomega.BeFalse())
		gomega.Expect(real.LockgateHandle.UUID).To(gomega.Equal("winner"))
		gomega.Expect(manager.Unlock(ctx, real)).To(gomega.Succeed())
		gomega.Expect(releases.Load()).To(gomega.Equal(int32(1)))
	})
})
