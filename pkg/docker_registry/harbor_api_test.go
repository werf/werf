package docker_registry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = DescribeTable("Harbor repository deletion", func(repository string, v2Status int, expectedRequests []string) {
	var requests []string
	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer GinkgoRecover()

		username, password, ok := request.BasicAuth()
		Expect(ok).To(BeTrue())
		Expect(username).To(Equal("user"))
		Expect(password).To(Equal("password"))
		requests = append(requests, request.Method+" "+request.RequestURI)

		if strings.HasPrefix(request.URL.Path, "/api/v2.0/") {
			writer.WriteHeader(v2Status)
			return
		}

		writer.WriteHeader(http.StatusAccepted)
	}))
	DeferCleanup(server.Close)

	registryAPI := &harborApi{httpClient: server.Client()}
	hostname := strings.TrimPrefix(server.URL, "https://")
	_, err := registryAPI.DeleteRepository(context.Background(), hostname, repository, "user", "password")
	Expect(err).NotTo(HaveOccurred())
	Expect(requests).To(Equal(expectedRequests))
},
	Entry("uses the v2 API", "project/image", http.StatusAccepted, []string{
		"DELETE /api/v2.0/projects/project/repositories/image",
	}),
	Entry("double-encodes a nested repository name", "project/nested/image", http.StatusAccepted, []string{
		"DELETE /api/v2.0/projects/project/repositories/nested%252Fimage",
	}),
	Entry("falls back to the v1 API", "project/image", http.StatusNotFound, []string{
		"DELETE /api/v2.0/projects/project/repositories/image",
		"DELETE /api/repositories/project/image",
	}),
)

var _ = It("rejects a Harbor repository without a project", func() {
	registryAPI := newHarborApi()
	_, err := registryAPI.DeleteRepository(context.Background(), "harbor.example.com", "image", "", "")
	Expect(err).To(MatchError(`invalid Harbor repository "image": expected project/name`))
})
