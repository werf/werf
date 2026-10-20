package transport

import (
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("registry rate-limit transport", func() {
	ginkgo.DescribeTable("returns a rejected non-replayable upload without retrying", func(status int) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPatch, "https://registry.test/blob", io.NopCloser(strings.NewReader("stream")))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		body := &trackedResponseBody{Reader: strings.NewReader("rejected")}
		want := &http.Response{StatusCode: status, Body: body, Header: http.Header{"Retry-After": []string{"0"}}}
		calls := 0
		transport := NewTransport(roundTripperFunc(func(sent *http.Request) (*http.Response, error) {
			defer sent.Body.Close()
			calls++
			gomega.Expect(calls).To(gomega.Equal(1))
			content, err := io.ReadAll(sent.Body)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(string(content)).To(gomega.Equal("stream"))
			return want, nil
		}))
		resp, err := transport.RoundTrip(req)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(resp).To(gomega.BeIdenticalTo(want))
		gomega.Expect(body.closed).To(gomega.BeFalse())
		gomega.Expect(resp.Body.Close()).To(gomega.Succeed())
	},
		ginkgo.Entry("429", http.StatusTooManyRequests),
		ginkgo.Entry("503", http.StatusServiceUnavailable),
	)

	ginkgo.DescribeTable("closes rejected responses before retrying", func(status int) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://registry.test/tags", nil)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		body := &trackedResponseBody{Reader: strings.NewReader("retry later")}
		calls := 0
		transport := NewTransport(roundTripperFunc(func(*http.Request) (*http.Response, error) {
			calls++
			if calls == 1 {
				return &http.Response{StatusCode: status, Body: body, Header: http.Header{"Retry-After": []string{"0"}}}, nil
			}
			gomega.Expect(body.closed).To(gomega.BeTrue())
			return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
		}))
		resp, err := transport.RoundTrip(req)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(resp.StatusCode).To(gomega.Equal(http.StatusOK))
		gomega.Expect(resp.Body.Close()).To(gomega.Succeed())
		gomega.Expect(calls).To(gomega.Equal(2))
	},
		ginkgo.Entry("429", http.StatusTooManyRequests),
		ginkgo.Entry("503", http.StatusServiceUnavailable),
	)

	ginkgo.DescribeTable("closes rejected responses when no retry follows", func(retryAfter string, cancelRequest bool) {
		ctx, cancel := context.WithCancel(context.Background())
		ginkgo.DeferCleanup(cancel)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://registry.test/tags", nil)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		body := &trackedResponseBody{Reader: strings.NewReader("retry later")}
		transport := NewTransport(roundTripperFunc(func(*http.Request) (*http.Response, error) {
			if cancelRequest {
				cancel()
			}
			return &http.Response{StatusCode: http.StatusServiceUnavailable, Body: body, Header: http.Header{"Retry-After": []string{retryAfter}}}, nil
		}))
		resp, err := transport.RoundTrip(req)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(resp).To(gomega.BeNil())
		gomega.Expect(body.closed).To(gomega.BeTrue())
		if cancelRequest {
			gomega.Expect(err).To(gomega.MatchError(context.Canceled))
		}
	},
		ginkgo.Entry("invalid Retry-After", "invalid", false),
		ginkgo.Entry("canceled request", "30", true),
	)
})
