package docker_registry

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/image"
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
		switch request.URL.Path {
		case "/v2/project/image/manifests/stage", "/v2/project/image/manifests/next":
			writer.WriteHeader(modernStatus)
			_, err := writer.Write([]byte(modernBody))
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		case "/v2/project/image/tags/reference/stage", "/v2/project/image/tags/reference/next":
			writer.WriteHeader(legacyStatus)
		case "/v2/project/image/manifests/" + digest:
			writer.WriteHeader(http.StatusAccepted)
		default:
			writer.WriteHeader(http.StatusNotFound)
		}
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
	} else {
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		info.Name = repository + ":next"
		info.Tag = "next"
		gomega.Expect(registry.DeleteRepoImage(context.Background(), info)).To(gomega.Succeed())
	}
	gomega.Expect(requests).To(gomega.Equal(expectedPaths))
},
	ginkgo.Entry("prefers the manifest tag endpoint and reuses it", http.StatusAccepted, "", http.StatusAccepted,
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
		[]string{"DELETE /v2/project/image/manifests/stage"}, "403 Forbidden"),
	ginkgo.Entry("does not delete a shared digest when the tag is missing", http.StatusNotFound, `{"errors":[{"code":"MANIFEST_UNKNOWN"}]}`, http.StatusNotFound,
		[]string{"DELETE /v2/project/image/manifests/stage"}, "MANIFEST_UNKNOWN"),
	ginkgo.Entry("does not fall back when the repository is missing", http.StatusNotFound, `{"errors":[{"code":"NAME_UNKNOWN"}]}`, http.StatusNotFound,
		[]string{"DELETE /v2/project/image/manifests/stage"}, "NAME_UNKNOWN"),
	ginkgo.Entry("does not fall back on a server error", http.StatusInternalServerError, "", http.StatusAccepted,
		[]string{"DELETE /v2/project/image/manifests/stage"}, "500 Internal Server Error"),
)

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
