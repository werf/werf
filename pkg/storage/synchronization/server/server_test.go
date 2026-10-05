package server

import (
	"net/http/httptest"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func TestServer(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Synchronization server")
}

var _ = ginkgo.Describe("Landing page", func() {
	ginkgo.DescribeTable("writes the configured address literally", func(address string) {
		original := DefaultAddress
		DefaultAddress = address
		ginkgo.DeferCleanup(func() { DefaultAddress = original })
		response := httptest.NewRecorder()
		server := &handler{}
		server.handleLanding(response, httptest.NewRequest("GET", "/", nil))
		gomega.Expect(response.Code).To(gomega.Equal(200))
		gomega.Expect(response.Body.String()).To(gomega.ContainSubstring("--synchronization=" + address))
		gomega.Expect(response.Body.String()).NotTo(gomega.ContainSubstring("%!"))
	},
		ginkgo.Entry("default", "https://synchronization.werf.io"),
		ginkgo.Entry("percent-encoded path", "https://example.com/path%20segment"),
	)
})
