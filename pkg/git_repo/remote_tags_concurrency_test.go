package git_repo

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/true_git"
	"github.com/werf/werf/v3/pkg/werf"
)

var _ = ginkgo.Describe("Remote tag concurrency", func() {
	ginkgo.It("resolves independent remotes while another request is blocked and preserves cache refresh", func(ctx ginkgo.SpecContext) {
		gomega.Expect(werf.Init(ginkgo.GinkgoT().TempDir(), ginkgo.GinkgoT().TempDir())).To(gomega.Succeed())
		gomega.Expect(true_git.Init(ctx, true_git.Options{})).To(gomega.Succeed())
		resetLsRemoteTagsCache()

		const firstSHA = "1111111111111111111111111111111111111111"
		const secondSHA = "2222222222222222222222222222222222222222"
		entered := make(chan struct{}, 2)
		release := make(chan struct{})
		var releaseOnce sync.Once
		var requests atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasSuffix(r.URL.Path, "/info/refs") {
				http.NotFound(w, r)
				return
			}
			request := requests.Add(1)
			if request <= 2 {
				entered <- struct{}{}
				select {
				case <-release:
				case <-r.Context().Done():
					return
				}
			}
			sha := firstSHA
			if request > 2 {
				sha = secondSHA
			}
			w.Header().Set("Content-Type", "text/plain")
			if _, err := fmt.Fprintf(w, "%s\trefs/tags/v1\n", sha); err != nil {
				return
			}
		}))
		defer server.Close()
		requestCtx := context.WithoutCancel(ctx)
		var workers sync.WaitGroup
		defer func() {
			releaseOnce.Do(func() { close(release) })
			workers.Wait()
		}()
		repos := []*Remote{{Url: server.URL + "/first", Tag: "v1"}, {Url: server.URL + "/second", Tag: "v1"}}
		for range 14 {
			repos = append(repos, repos[0])
		}
		results := make(chan string, len(repos))
		errs := make(chan error, len(repos))
		for _, repo := range repos {
			workers.Add(1)
			go func() {
				defer workers.Done()
				sha, err := repo.lsRemoteTag(requestCtx, false)
				results <- sha
				errs <- err
			}()
		}
		gomega.Eventually(entered, 10*time.Second).Should(gomega.Receive())
		gomega.Eventually(entered, 10*time.Second).Should(gomega.Receive(), "independent ls-remote requests must reach the server before either is released")
		releaseOnce.Do(func() { close(release) })
		for range repos {
			gomega.Eventually(results, 10*time.Second).Should(gomega.Receive(gomega.Equal(firstSHA)))
			gomega.Eventually(errs, 10*time.Second).Should(gomega.Receive(gomega.BeNil()))
		}

		sha, err := repos[0].lsRemoteTag(ctx, false)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(sha).To(gomega.Equal(firstSHA))
		gomega.Expect(requests.Load()).To(gomega.Equal(int32(2)))
		sha, err = repos[0].lsRemoteTag(ctx, true)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(sha).To(gomega.Equal(secondSHA))
		gomega.Expect(requests.Load()).To(gomega.Equal(int32(3)))
		sha, err = repos[0].lsRemoteTag(ctx, false)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(sha).To(gomega.Equal(secondSHA))
		gomega.Expect(requests.Load()).To(gomega.Equal(int32(3)))
		authenticated := &Remote{Url: repos[0].Url, Tag: "v1", BasicAuth: &BasicAuth{Username: "another-user", Password: "another-password"}}
		sha, err = authenticated.lsRemoteTag(ctx, false)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(sha).To(gomega.Equal(secondSHA))
		gomega.Expect(requests.Load()).To(gomega.Equal(int32(4)))
	})
})
