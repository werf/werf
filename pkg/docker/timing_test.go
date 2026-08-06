package docker

import (
	"context"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/opstats"
)

var _ = Describe("timingRoundTripper", func() {
	It("records request duration into the collector from the request context", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		collector := opstats.NewCollector()
		ctx := opstats.NewContext(context.Background(), collector)

		httpClient := &http.Client{Transport: &timingRoundTripper{next: http.DefaultTransport}}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
		Expect(err).NotTo(HaveOccurred())

		resp, err := httpClient.Do(req)
		Expect(err).NotTo(HaveOccurred())
		defer resp.Body.Close()

		summary := collector.Summary()
		Expect(summary).To(HaveLen(1))
		Expect(summary[0].Operation).To(Equal(opstats.OperationDockerDaemon))
		Expect(summary[0].Count).To(Equal(1))
	})

	It("does not record without a collector in the request context", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		httpClient := &http.Client{Transport: &timingRoundTripper{next: http.DefaultTransport}}

		resp, err := httpClient.Get(server.URL)
		Expect(err).NotTo(HaveOccurred())
		defer resp.Body.Close()
	})
})
