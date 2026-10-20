package docker_registry

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("registry request body transport", func() {
	ginkgo.It("closes the unused original before opening each replacement and preserves the request", func() {
		original := &trackedRequestBody{Reader: strings.NewReader("manifest")}
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, "https://registry.test/manifest", original)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		req.GetBody = func() (io.ReadCloser, error) {
			gomega.Expect(original.closed).To(gomega.BeTrue())
			return io.NopCloser(strings.NewReader("manifest")), nil
		}
		var bodies []string
		transport := &requestBodyTransport{underlying: bearerRoundTripperFunc(func(sent *http.Request) (*http.Response, error) {
			defer sent.Body.Close()
			gomega.Expect(sent).NotTo(gomega.BeIdenticalTo(req))
			body, err := io.ReadAll(sent.Body)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			bodies = append(bodies, string(body))
			return &http.Response{StatusCode: http.StatusCreated, Body: http.NoBody}, nil
		})}
		for range 2 {
			resp, err := transport.RoundTrip(req)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(resp.Body.Close()).To(gomega.Succeed())
			gomega.Expect(req.Body).To(gomega.BeIdenticalTo(original))
		}
		gomega.Expect(bodies).To(gomega.Equal([]string{"manifest", "manifest"}))
	})

	ginkgo.It("returns a body factory error without sending and closes the original", func() {
		original := &trackedRequestBody{Reader: strings.NewReader("manifest")}
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, "https://registry.test/manifest", original)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		factoryErr := errors.New("body unavailable")
		req.GetBody = func() (io.ReadCloser, error) { return nil, factoryErr }
		transport := &requestBodyTransport{underlying: bearerRoundTripperFunc(func(*http.Request) (*http.Response, error) {
			ginkgo.Fail("sent a request without its body")
			return nil, factoryErr
		})}
		resp, err := transport.RoundTrip(req)
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring(factoryErr.Error())))
		gomega.Expect(errors.Is(err, factoryErr)).To(gomega.BeTrue())
		gomega.Expect(resp).To(gomega.BeNil())
		gomega.Expect(original.closed).To(gomega.BeTrue())
	})

	ginkgo.It("passes a non-replayable stream through without closing it early", func() {
		original := &trackedRequestBody{Reader: strings.NewReader("stream")}
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPatch, "https://registry.test/blob", original)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		transport := &requestBodyTransport{underlying: bearerRoundTripperFunc(func(sent *http.Request) (*http.Response, error) {
			defer sent.Body.Close()
			gomega.Expect(sent).To(gomega.BeIdenticalTo(req))
			gomega.Expect(original.closed).To(gomega.BeFalse())
			body, err := io.ReadAll(sent.Body)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(string(body)).To(gomega.Equal("stream"))
			return &http.Response{StatusCode: http.StatusAccepted, Body: http.NoBody}, nil
		})}
		resp, err := transport.RoundTrip(req)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(resp.StatusCode).To(gomega.Equal(http.StatusAccepted))
		gomega.Expect(resp.Body.Close()).To(gomega.Succeed())
		gomega.Expect(original.closed).To(gomega.BeTrue())
	})
})
