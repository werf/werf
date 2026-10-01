package lock_manager

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/storage"
)

var _ = ginkgo.Describe("Synchronization client identity", func() {
	ginkgo.It("reuses the oldest registered namespace without contacting the server", func(ctx ginkgo.SpecContext) {
		repo := &clientRecordStorage{records: []*storage.ClientIDRecord{
			{ClientID: "new", TimestampMillisec: 20},
			{ClientID: "old", TimestampMillisec: 10},
		}}
		id, err := GetHttpClientID(ctx, "project", ":invalid-url", repo)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(id).To(gomega.Equal("old"))
	})

	ginkgo.It("propagates registry read errors without registering another namespace", func(ctx ginkgo.SpecContext) {
		expected := errors.New("registry unavailable")
		id, err := GetHttpClientID(ctx, "project", ":invalid-url", &clientRecordStorage{readErr: expected})
		gomega.Expect(err).To(gomega.MatchError(expected))
		gomega.Expect(id).To(gomega.BeEmpty())
	})

	ginkgo.It("converges on one registered namespace after concurrent registration", func(ctx ginkgo.SpecContext) {
		var requests atomic.Int32
		bothRegistered := make(chan struct{})
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer ginkgo.GinkgoRecover()
			n := requests.Add(1)
			if n == 2 {
				close(bothRegistered)
			}
			<-bothRegistered
			id := "first"
			if n == 2 {
				id = "second"
			}
			_, err := w.Write([]byte(`{"clientID":"` + id + `"}`))
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		}))
		defer server.Close()
		repo := &clientRecordStorage{}
		type result struct {
			id  string
			err error
		}
		results := make(chan result, 2)
		for range 2 {
			go func() {
				defer ginkgo.GinkgoRecover()
				id, err := GetHttpClientID(ctx, "project", server.URL, repo)
				results <- result{id, err}
			}()
		}
		var first, second result
		gomega.Eventually(results, "5s").Should(gomega.Receive(&first))
		gomega.Eventually(results, "5s").Should(gomega.Receive(&second))
		gomega.Expect(first.err).NotTo(gomega.HaveOccurred())
		gomega.Expect(second.err).NotTo(gomega.HaveOccurred())
		gomega.Expect(first.id).NotTo(gomega.BeEmpty())
		gomega.Expect(second.id).To(gomega.Equal(first.id))
		gomega.Expect(requests.Load()).To(gomega.Equal(int32(2)))
	})

	ginkgo.It("does not use a namespace whose registry publication failed", func(ctx ginkgo.SpecContext) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer ginkgo.GinkgoRecover()
			_, err := w.Write([]byte(`{"clientID":"unpublished"}`))
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		}))
		defer server.Close()
		expected := errors.New("registry write denied")
		id, err := GetHttpClientID(ctx, "project", server.URL, &clientRecordStorage{writeErr: expected})
		gomega.Expect(err).To(gomega.MatchError(expected))
		gomega.Expect(id).To(gomega.BeEmpty())
	})

	ginkgo.DescribeTable("rejects invalid synchronization responses",
		func(ctx ginkgo.SpecContext, status int, body string) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer ginkgo.GinkgoRecover()
				w.WriteHeader(status)
				_, err := w.Write([]byte(body))
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			}))
			defer server.Close()
			repo := &clientRecordStorage{}
			id, err := GetHttpClientID(ctx, "project", server.URL, repo)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(id).To(gomega.BeEmpty())
			gomega.Expect(repo.records).To(gomega.BeEmpty())
		},
		ginkgo.Entry("HTTP failure", http.StatusServiceUnavailable, "unavailable"),
		ginkgo.Entry("malformed JSON", http.StatusOK, "not json"),
	)
})
