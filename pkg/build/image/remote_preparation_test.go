package image

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/common-go/pkg/graceful"
	"github.com/werf/werf/v3/pkg/build/stage"
	"github.com/werf/werf/v3/pkg/config"
)

var _ = ginkgo.Describe("Remote Git preparation", func() {
	ginkgo.It("bounds concurrent clones and reuses selected repositories across images and platforms", func(ctx ginkgo.SpecContext) {
		requests := make(chan string, 20)
		release := make(chan struct{})
		var releaseOnce sync.Once
		var mutex sync.Mutex
		counts := map[string]int{}
		tree := newRemotePreparationTree(ctx, func(backend http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, "/info/refs") {
					mutex.Lock()
					counts[r.URL.Path]++
					mutex.Unlock()
					requests <- r.URL.Path
					select {
					case <-release:
					case <-r.Context().Done():
						return
					}
				}
				backend.ServeHTTP(w, r)
			})
		})
		useRemotePreparationBranches(tree)
		done := make(chan error, 1)
		runCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
		runCtx = graceful.WithTermination(runCtx)
		finished := make(chan struct{})
		defer func() {
			cancel()
			releaseOnce.Do(func() { close(release) })
			<-finished
		}()
		go func() { defer close(finished); done <- tree.Calculate(runCtx) }()
		gomega.Eventually(requests, 15*time.Second).Should(gomega.Receive())
		gomega.Eventually(requests, 15*time.Second).Should(gomega.Receive())
		gomega.Consistently(requests, 200*time.Millisecond).ShouldNot(gomega.Receive())
		releaseOnce.Do(func() { close(release) })
		gomega.Eventually(done, 30*time.Second).Should(gomega.Receive(gomega.Succeed()))
		gomega.Expect(tree.GetImages()).To(gomega.HaveLen(4))
		mutex.Lock()
		actualCounts := make(map[string]int, len(counts))
		for key, value := range counts {
			actualCounts[key] = value
		}
		mutex.Unlock()
		gomega.Expect(actualCounts).To(gomega.Equal(map[string]int{"/one.git/info/refs": 1, "/two.git/info/refs": 1, "/three.git/info/refs": 1}))
		parallelInputs := remotePreparationInputs(ctx, tree)
		serialOpts := tree.ImagesTreeOptions
		serialOpts.Conveyor = &preparationTestConveyor{}
		serialOpts.RemoteGitTasksLimit = 1
		serial := NewImagesTree(tree.werfConfig, serialOpts)
		gomega.Expect(serial.Calculate(ctx)).To(gomega.Succeed())
		gomega.Expect(remotePreparationInputs(ctx, serial)).To(gomega.Equal(parallelInputs))
	})

	ginkgo.DescribeTable("prepares one ref per storage mirror and resolves both refs", func(ctx ginkgo.SpecContext, alias bool) {
		tree := newRemotePreparationTree(ctx, func(backend http.Handler) http.Handler { return backend })
		app := tree.werfConfig.GetImage("app").(config.StapelImageInterface).ImageBaseConfig()
		original := app.Git.Remote[0]
		originalCommit := original.Commit
		otherRef := *original
		otherExport := *original.GitRemoteExport
		otherRef.GitRemoteExport = &otherExport
		otherRef.Commit = ""
		otherRef.Branch = "main"
		otherRef.RepoCacheKey += "-branch"
		if alias {
			original.Url = strings.Replace(original.Url, "http://", "http://alice@", 1)
			otherRef.Url = strings.Replace(otherRef.Url, "http://", "http://bob@", 1)
			otherRef.Name += "-alias"
		}
		gomega.Expect(original.Commit).To(gomega.Equal(originalCommit))
		gomega.Expect(original.Branch).To(gomega.BeEmpty())
		app.Git.Remote = append(app.Git.Remote, &otherRef)
		gomega.Expect(tree.prepareRemoteGitRepos(ctx, tree.werfConfig.GetImagesForProcessing(tree.ImagesToProcess))).To(gomega.Succeed())
		gomega.Expect(tree.Conveyor.GetRemoteGitRepo(original.RepoCacheKey)).NotTo(gomega.BeNil())
		gomega.Expect(tree.Conveyor.GetRemoteGitRepo(otherRef.RepoCacheKey)).To(gomega.BeNil())
		gomega.Expect(tree.Calculate(ctx)).To(gomega.Succeed())
		for _, image := range tree.GetImages() {
			if image.Name != "app" {
				continue
			}
			mappings := image.GetStage(stage.GitArchive).GetGitMappings()
			gomega.Expect(mappings).To(gomega.HaveLen(4))
			first, err := mappings[0].GetLatestCommitInfo(ctx, tree.Conveyor)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			second, err := mappings[3].GetLatestCommitInfo(ctx, tree.Conveyor)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(first.Commit).To(gomega.Equal(originalCommit))
			gomega.Expect(second.Commit).NotTo(gomega.Equal(originalCommit))
		}
	}, ginkgo.Entry("same URL", false), ginkgo.Entry("different URL usernames sharing storage", true))

	ginkgo.DescribeTable("does not retain failed preparation and permits retry", func(ctx ginkgo.SpecContext, cancelRequest bool) {
		var fail atomic.Bool
		var cancellationTimedOut atomic.Bool
		fail.Store(true)
		runCtx, cancel := context.WithCancel(ctx)
		runCtx = graceful.WithTermination(runCtx)
		defer cancel()
		tree := newRemotePreparationTree(ctx, func(backend http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if fail.Load() {
					if cancelRequest {
						cancel()
						select {
						case <-r.Context().Done():
						case <-time.After(5 * time.Second):
							cancellationTimedOut.Store(true)
							http.Error(w, "cancellation timeout", http.StatusGatewayTimeout)
						}
						return
					}
					http.Error(w, "preparation failure", http.StatusServiceUnavailable)
					return
				}
				backend.ServeHTTP(w, r)
			})
		})
		if cancelRequest {
			useRemotePreparationBranches(tree)
		}
		gomega.Expect(tree.Calculate(runCtx)).NotTo(gomega.Succeed())
		gomega.Expect(cancellationTimedOut.Load()).To(gomega.BeFalse(), "canceled Calculate must terminate its in-flight Git request")
		conveyor := tree.Conveyor.(*preparationTestConveyor)
		conveyor.remoteMutex.Lock()
		cached := len(conveyor.remotes)
		conveyor.remoteMutex.Unlock()
		gomega.Expect(cached).To(gomega.BeZero())
		gomega.Expect(tree.GetImages()).To(gomega.BeEmpty())
		fail.Store(false)
		gomega.Expect(tree.Calculate(ctx)).To(gomega.Succeed())
		gomega.Expect(tree.GetImages()).To(gomega.HaveLen(4))
	}, ginkgo.Entry("HTTP failure", false), ginkgo.Entry("canceled full clone request", true))
})
