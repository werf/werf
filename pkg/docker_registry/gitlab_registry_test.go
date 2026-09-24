package docker_registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/image"
)

var _ = ginkgo.DescribeTable("GitLab tag deletion", func(modernStatus int, modernBody string, legacyStatus int, expectedPaths []string, expectedError string) {
	const digest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer ginkgo.GinkgoRecover()
		if request.Method == http.MethodGet && request.URL.Path == "/v2/" {
			writer.WriteHeader(http.StatusOK)
			return
		}

		requests = append(requests, request.Method+" "+request.URL.Path)
		responseStatus := http.StatusNotFound
		responseBody := ""
		switch request.URL.Path {
		case "/v2/project/image/manifests/stage", "/v2/project/image/manifests/next":
			responseStatus, responseBody = modernStatus, modernBody
		case "/v2/project/image/tags/reference/stage", "/v2/project/image/tags/reference/next":
			responseStatus = legacyStatus
		case "/v2/project/image/manifests/" + digest:
			responseStatus = http.StatusAccepted
		}
		if responseBody == "" && (responseStatus == http.StatusNotFound || responseStatus == http.StatusMethodNotAllowed) {
			connection, buffer, err := writer.(http.Hijacker).Hijack()
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			defer connection.Close()
			_, err = fmt.Fprintf(buffer, "HTTP/1.1 %d Custom Reason\r\nContent-Length: 0\r\nConnection: close\r\n\r\n", responseStatus)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(buffer.Flush()).To(gomega.Succeed())
			return
		}
		writer.WriteHeader(responseStatus)
		_, err := writer.Write([]byte(responseBody))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}))
	ginkgo.DeferCleanup(server.Close)

	registry, err := newGitLabRegistry(gitLabRegistryOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	repository := strings.TrimPrefix(server.URL, "http://") + "/project/image"
	info := &image.Info{
		Name:       repository + ":stage",
		Repository: repository,
		Tag:        "stage",
		RepoDigest: repository + "@" + digest,
	}
	err = registry.DeleteRepoImage(context.Background(), info)
	if expectedError != "" {
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring(expectedError)))
		var transportErr *transport.Error
		gomega.Expect(errors.As(err, &transportErr)).To(gomega.BeTrue())
		gomega.Expect(transportErr.StatusCode).To(gomega.Equal(modernStatus))
		if modernBody == "" {
			gomega.Expect(strings.Count(err.Error(), expectedError)).To(gomega.Equal(1))
		}
	} else {
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		info.Name = repository + ":next"
		info.Tag = "next"
		gomega.Expect(registry.DeleteRepoImage(context.Background(), info)).To(gomega.Succeed())
	}
	gomega.Expect(requests).To(gomega.Equal(expectedPaths))
},
	ginkgo.Entry("prefers the manifest tag endpoint", http.StatusAccepted, "", http.StatusAccepted,
		[]string{"DELETE /v2/project/image/manifests/stage", "DELETE /v2/project/image/manifests/next"}, ""),
	ginkgo.Entry("accepts an OK response", http.StatusOK, "", http.StatusAccepted,
		[]string{"DELETE /v2/project/image/manifests/stage", "DELETE /v2/project/image/manifests/next"}, ""),
	ginkgo.Entry("falls back to the legacy tag endpoint on 404 and reuses it", http.StatusNotFound, "", http.StatusAccepted,
		[]string{"DELETE /v2/project/image/manifests/stage", "DELETE /v2/project/image/tags/reference/stage", "DELETE /v2/project/image/tags/reference/next"}, ""),
	ginkgo.Entry("falls back to the legacy tag endpoint on 405", http.StatusMethodNotAllowed, "", http.StatusAccepted,
		[]string{"DELETE /v2/project/image/manifests/stage", "DELETE /v2/project/image/tags/reference/stage", "DELETE /v2/project/image/tags/reference/next"}, ""),
	ginkgo.Entry("falls back when a registry requires a digest", http.StatusBadRequest, `{"errors":[{"code":"DIGEST_INVALID","message":"provided digest did not match uploaded content"}]}`, http.StatusAccepted,
		[]string{"DELETE /v2/project/image/manifests/stage", "DELETE /v2/project/image/tags/reference/stage", "DELETE /v2/project/image/tags/reference/next"}, ""),
	ginkgo.Entry("retains digest fallback when both tag APIs are unavailable", http.StatusBadRequest, `{"errors":[{"code":"DIGEST_INVALID"}]}`, http.StatusNotFound,
		[]string{"DELETE /v2/project/image/manifests/stage", "DELETE /v2/project/image/tags/reference/stage", "DELETE /v2/project/image/manifests/sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "DELETE /v2/project/image/manifests/sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}, ""),
	ginkgo.Entry("does not fall back on forbidden deletion", http.StatusForbidden, `{"errors":[{"code":"DENIED"}]}`, http.StatusAccepted,
		[]string{"DELETE /v2/project/image/manifests/stage"}, "DENIED"),
	ginkgo.Entry("does not delete a shared digest when the tag is missing", http.StatusNotFound, `{"errors":[{"code":"MANIFEST_UNKNOWN"}]}`, http.StatusNotFound,
		[]string{"DELETE /v2/project/image/manifests/stage", "DELETE /v2/project/image/manifests/next"}, ""),
	ginkgo.Entry("does not fall back when the repository is missing", http.StatusNotFound, `{"errors":[{"code":"NAME_UNKNOWN"}]}`, http.StatusNotFound,
		[]string{"DELETE /v2/project/image/manifests/stage", "DELETE /v2/project/image/manifests/next"}, ""),
	ginkgo.Entry("does not fall back on a server error", http.StatusInternalServerError, "", http.StatusAccepted,
		[]string{"DELETE /v2/project/image/manifests/stage"}, "500 Internal Server Error"),
	ginkgo.Entry("does not treat diagnostic message text as an error code", http.StatusInternalServerError, `{"errors":[{"code":"UNKNOWN","message":"DIGEST_INVALID UNAUTHORIZED 404 Not Found"}]}`, http.StatusAccepted,
		[]string{"DELETE /v2/project/image/manifests/stage"}, "UNKNOWN"),
	ginkgo.Entry("does not suppress server errors carrying MANIFEST_UNKNOWN", http.StatusInternalServerError, `{"errors":[{"code":"MANIFEST_UNKNOWN"}]}`, http.StatusAccepted,
		[]string{"DELETE /v2/project/image/manifests/stage"}, "MANIFEST_UNKNOWN"),
	ginkgo.Entry("does not suppress server errors carrying NAME_UNKNOWN", http.StatusInternalServerError, `{"errors":[{"code":"NAME_UNKNOWN"}]}`, http.StatusAccepted,
		[]string{"DELETE /v2/project/image/manifests/stage"}, "NAME_UNKNOWN"),
)

var _ = ginkgo.DescribeTable("GitLab missing legacy tag", func(code string, warmCache bool) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer ginkgo.GinkgoRecover()
		if request.Method == http.MethodGet && request.URL.Path == "/v2/" {
			writer.WriteHeader(http.StatusOK)
			return
		}
		requests = append(requests, request.Method+" "+request.URL.Path)
		switch request.URL.Path {
		case "/v2/project/image/manifests/stage":
			writer.WriteHeader(http.StatusMethodNotAllowed)
		case "/v2/project/image/tags/reference/stage":
			if warmCache {
				writer.WriteHeader(http.StatusAccepted)
				return
			}
			fallthrough
		case "/v2/project/image/tags/reference/missing":
			writer.WriteHeader(http.StatusNotFound)
			gomega.Expect(json.NewEncoder(writer).Encode(map[string]interface{}{
				"errors": []map[string]string{{"code": code}},
			})).To(gomega.Succeed())
		default:
			writer.WriteHeader(http.StatusAccepted)
		}
	}))
	ginkgo.DeferCleanup(server.Close)
	registry, err := newGitLabRegistry(gitLabRegistryOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	repository := strings.TrimPrefix(server.URL, "http://") + "/project/image"
	for _, tag := range []string{"stage", "missing", "next"} {
		gomega.Expect(registry.DeleteRepoImage(context.Background(), &image.Info{
			Name: repository + ":" + tag, Repository: repository, Tag: tag,
			RepoDigest: repository + "@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		})).To(gomega.Succeed())
	}
	gomega.Expect(requests).To(gomega.Equal([]string{
		"DELETE /v2/project/image/manifests/stage",
		"DELETE /v2/project/image/tags/reference/stage",
		"DELETE /v2/project/image/tags/reference/missing",
		"DELETE /v2/project/image/tags/reference/next",
	}))
},
	ginkgo.Entry("ignores MANIFEST_UNKNOWN during endpoint selection", "MANIFEST_UNKNOWN", false),
	ginkgo.Entry("ignores NAME_UNKNOWN during endpoint selection", "NAME_UNKNOWN", false),
	ginkgo.Entry("ignores MANIFEST_UNKNOWN with a cached endpoint", "MANIFEST_UNKNOWN", true),
	ginkgo.Entry("ignores NAME_UNKNOWN with a cached endpoint", "NAME_UNKNOWN", true),
)

var _ = ginkgo.It("GitLab deletes tags concurrently with cold and warm caches", func() {
	const workers = 8
	var arrivals atomic.Int32
	allArrived := make(chan struct{})
	requests := make(chan string, workers*2)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodGet && request.URL.Path == "/v2/" {
			writer.WriteHeader(http.StatusOK)
			return
		}
		requests <- request.Method + " " + request.URL.Path
		if arrivals.Add(1) == workers {
			close(allArrived)
		}
		select {
		case <-allArrived:
			writer.WriteHeader(http.StatusAccepted)
		case <-time.After(10 * time.Second):
			writer.WriteHeader(http.StatusGatewayTimeout)
		}
	}))
	ginkgo.DeferCleanup(server.Close)
	registry, err := newGitLabRegistry(gitLabRegistryOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	repository := strings.TrimPrefix(server.URL, "http://") + "/project/image"
	results := make(chan error, workers)
	var expectedRequests []string
	for wave := 0; wave < 2; wave++ {
		for worker := 0; worker < workers; worker++ {
			tag := fmt.Sprintf("stage-%d-%d", wave, worker)
			expectedRequests = append(expectedRequests, "DELETE /v2/project/image/manifests/"+tag)
			go func() {
				results <- registry.DeleteRepoImage(context.Background(), &image.Info{
					Name: repository + ":" + tag, Repository: repository, Tag: tag,
				})
			}()
		}
		for worker := 0; worker < workers; worker++ {
			gomega.Eventually(results, 20*time.Second).Should(gomega.Receive(gomega.BeNil()))
		}
	}
	var actualRequests []string
	for range expectedRequests {
		actualRequests = append(actualRequests, <-requests)
	}
	gomega.Expect(actualRequests).To(gomega.ConsistOf(expectedRequests))
})

var _ = ginkgo.DescribeTable("GitLab digest deletion", func(tag string, warmCache bool) {
	const digest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodGet && request.URL.Path == "/v2/" {
			writer.WriteHeader(http.StatusOK)
			return
		}
		requests = append(requests, request.Method+" "+request.URL.Path)
		writer.WriteHeader(http.StatusAccepted)
	}))
	ginkgo.DeferCleanup(server.Close)
	registry, err := newGitLabRegistry(gitLabRegistryOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	repository := strings.TrimPrefix(server.URL, "http://") + "/project/image"
	var expectedPaths []string
	if warmCache {
		gomega.Expect(registry.DeleteRepoImage(context.Background(), &image.Info{
			Name: repository + ":stage", Repository: repository, Tag: "stage", RepoDigest: repository + "@" + digest,
		})).To(gomega.Succeed())
		expectedPaths = append(expectedPaths, "DELETE /v2/project/image/manifests/stage")
	}
	gomega.Expect(registry.DeleteRepoImage(context.Background(), &image.Info{
		Name: repository + "@" + digest, Repository: repository, Tag: tag, RepoDigest: repository + "@" + digest,
	})).To(gomega.Succeed())
	expectedPaths = append(expectedPaths, "DELETE /v2/project/image/manifests/"+digest)
	gomega.Expect(requests).To(gomega.Equal(expectedPaths))
},
	ginkgo.Entry("does not interpret an empty tag as latest", "", false),
	ginkgo.Entry("does not delete a synthesized latest tag", "latest", false),
	ginkgo.Entry("does not reuse cached tag deletion for an empty tag", "", true),
	ginkgo.Entry("does not reuse cached tag deletion for a digest reference", "latest", true),
)

var _ = ginkgo.DescribeTable("GitLab deletion authentication", func(allowWildcard bool) {
	var scopes []string
	var requests []string
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer ginkgo.GinkgoRecover()
		switch request.URL.Path {
		case "/v2/":
			writer.Header().Set("Www-Authenticate", fmt.Sprintf(`Bearer realm="%s/token",service="registry"`, server.URL))
			writer.WriteHeader(http.StatusUnauthorized)
		case "/token":
			scope := strings.Join(request.URL.Query()["scope"], " ")
			scopes = append(scopes, scope)
			token := "full"
			if strings.Contains(scope, ":*") {
				token = "wildcard"
			}
			gomega.Expect(json.NewEncoder(writer).Encode(map[string]string{"token": token})).To(gomega.Succeed())
		default:
			requests = append(requests, request.Method+" "+request.URL.Path+" "+request.Header.Get("Authorization"))
			if allowWildcard && request.Header.Get("Authorization") == "Bearer wildcard" {
				writer.WriteHeader(http.StatusAccepted)
				return
			}
			writer.WriteHeader(http.StatusUnauthorized)
			gomega.Expect(json.NewEncoder(writer).Encode(map[string]interface{}{
				"errors": []map[string]string{{"code": "UNAUTHORIZED"}},
			})).To(gomega.Succeed())
		}
	}))
	ginkgo.DeferCleanup(server.Close)
	registry, err := newGitLabRegistry(gitLabRegistryOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	repository := strings.TrimPrefix(server.URL, "http://") + "/project/image"
	info := &image.Info{
		Name: repository + ":stage", Repository: repository, Tag: "stage",
		RepoDigest: repository + "@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	err = registry.DeleteRepoImage(context.Background(), info)
	expectedRequests := []string{
		"DELETE /v2/project/image/manifests/stage Bearer full",
		"DELETE /v2/project/image/manifests/stage Bearer wildcard",
	}
	if allowWildcard {
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		info.Name = repository + ":next"
		info.Tag = "next"
		gomega.Expect(registry.DeleteRepoImage(context.Background(), info)).To(gomega.Succeed())
		expectedRequests = append(expectedRequests, "DELETE /v2/project/image/manifests/next Bearer wildcard")
	} else {
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("UNAUTHORIZED")))
	}
	gomega.Expect(requests).To(gomega.Equal(expectedRequests))
	gomega.Expect(scopes).To(gomega.HaveLen(len(expectedRequests)))
	gomega.Expect(scopes[0]).To(gomega.And(gomega.ContainSubstring("push"), gomega.ContainSubstring("pull"), gomega.ContainSubstring("delete")))
	gomega.Expect(scopes[1:]).To(gomega.HaveEach("repository:project/image:*"))
},
	ginkgo.Entry("retries with wildcard scope and caches the successful scope", true),
	ginkgo.Entry("does not try other deletion APIs when both scopes are denied", false),
)
