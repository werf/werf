package docker_registry

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("registry request body retries", func() {
	ginkgo.BeforeEach(func() {
		configDir := ginkgo.GinkgoT().TempDir()
		gomega.Expect(os.WriteFile(filepath.Join(configDir, "config.json"), []byte(`{"auths":{}}`), 0o600)).To(gomega.Succeed())
		ginkgo.GinkgoT().Setenv("DOCKER_CONFIG", configDir)
	})

	ginkgo.DescribeTable("preserves the manifest and registry result after a rejected upload", func(status int, useTLS bool, finalStatus int) {
		fixture := newWritableBearerRegistryFixtureWithTLS(useTLS)
		ginkgo.DeferCleanup(fixture.server.Close)
		fixture.server.Config.SetKeepAlivesEnabled(false)
		var accepted atomic.Int64
		accepted.Store(1)
		fixture.authorize = func(req *http.Request) bool {
			return req.Header.Get("Authorization") == fmt.Sprintf("Bearer token-%d", accepted.Load())
		}
		fixture.tokenReply = func(w http.ResponseWriter, req *http.Request, count int64) {
			defer ginkgo.GinkgoRecover()
			accepted.Store(count)
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Connection", "close")
			_, err := fmt.Fprintf(w, `{"token":"token-%d","expires_in":300}`, count)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		}

		var mu sync.Mutex
		var manifests [][]byte
		backend := fixture.backend
		fixture.backend = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			defer ginkgo.GinkgoRecover()
			if req.Method != http.MethodPut || !strings.Contains(req.URL.Path, "/manifests/") {
				backend.ServeHTTP(w, req)
				return
			}

			body, err := io.ReadAll(req.Body)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			mu.Lock()
			manifests = append(manifests, body)
			first := len(manifests) == 1
			mu.Unlock()
			if first {
				w.Header().Set("Connection", "close")
				if status == http.StatusUnauthorized {
					accepted.Store(2)
					w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer realm="%s/token",service="fixture"`, fixture.server.URL))
				} else {
					w.Header().Set("Retry-After", "0")
				}
				w.WriteHeader(status)
				return
			}

			if finalStatus != http.StatusCreated {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(finalStatus)
				_, err := fmt.Fprint(w, `{"errors":[{"code":"DENIED","message":"fixture registry rejection"}]}`)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				return
			}

			req.Body = io.NopCloser(bytes.NewReader(body))
			backend.ServeHTTP(w, req)
		})

		registry := newAPI(apiOptions{InsecureRegistry: !useTLS, SkipTlsVerifyRegistry: useTLS})
		ref := strings.TrimPrefix(strings.TrimPrefix(fixture.server.URL, "http://"), "https://") + "/repo:retry"
		ctx, cancel := context.WithCancel(context.Background())
		ginkgo.DeferCleanup(cancel)
		err := registry.PushImage(ctx, ref, &PushImageOptions{Labels: map[string]string{"retry": "complete"}})
		if finalStatus == http.StatusCreated {
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		} else {
			var registryErr *transport.Error
			gomega.Expect(errors.As(err, &registryErr)).To(gomega.BeTrue(), "expected registry error, got %v", err)
			gomega.Expect(registryErr.StatusCode).To(gomega.Equal(finalStatus))
			gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("fixture registry rejection")))
		}

		mu.Lock()
		bodies := append([][]byte(nil), manifests...)
		mu.Unlock()
		gomega.Expect(bodies).To(gomega.HaveLen(2))
		gomega.Expect(bodies[0]).NotTo(gomega.BeEmpty())
		gomega.Expect(bodies[1]).To(gomega.Equal(bodies[0]))
		if status == http.StatusUnauthorized {
			gomega.Expect(fixture.exchanges.Load()).To(gomega.Equal(int64(2)))
		}
		if finalStatus != http.StatusCreated {
			return
		}
		info, err := registry.GetRepoImage(ctx, ref)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(info.Labels).To(gomega.HaveKeyWithValue("retry", "complete"))
	},
		ginkgo.Entry("renews the token", http.StatusUnauthorized, false, http.StatusCreated),
		ginkgo.Entry("renews the token over TLS", http.StatusUnauthorized, true, http.StatusCreated),
		ginkgo.Entry("recovers from rate limiting", http.StatusTooManyRequests, false, http.StatusCreated),
		ginkgo.Entry("recovers from service unavailability", http.StatusServiceUnavailable, false, http.StatusCreated),
		ginkgo.Entry("reports a repeated authorization failure", http.StatusUnauthorized, false, http.StatusUnauthorized),
		ginkgo.Entry("reports a registry rejection after token renewal", http.StatusUnauthorized, true, http.StatusForbidden),
	)
})
