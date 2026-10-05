package gitdata

import (
	"math"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/common-go/pkg/util/timestamps"
	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/werf"
)

var _ = ginkgo.Describe("RunGC dry run", func() {
	var localCache string

	ginkgo.BeforeEach(func() {
		ginkgo.GinkgoT().Setenv("WERF_HOME", ginkgo.GinkgoT().TempDir())
		ginkgo.GinkgoT().Setenv("WERF_TMP_DIR", ginkgo.GinkgoT().TempDir())
		gomega.Expect(werf.Init("", "")).To(gomega.Succeed())

		localCache = werf.GetLocalCacheDir()
		gcFixture()
	})

	ginkgo.It("touches nothing under maximum volume pressure", func(ctx ginkgo.SpecContext) {
		before := snapshotTree(localCache)

		gomega.Expect(RunGC(ctx, RunGCOptions{DryRun: true})).To(gomega.Succeed())

		expectTreeUnchanged(localCache, before)
	})

	ginkgo.It("touches nothing when the volume usage is below the allowed level", func(ctx ginkgo.SpecContext) {
		before := snapshotTree(localCache)

		gomega.Expect(RunGC(ctx, RunGCOptions{
			AllowedLocalCacheVolumeUsageBytes: math.MaxUint64,
			DryRun:                            true,
		})).To(gomega.Succeed())

		expectTreeUnchanged(localCache, before)
	})

	ginkgo.It("removes the same fixture for real when dry run is off", func(ctx ginkgo.SpecContext) {
		staleVersionDir := filepath.Join(localCache, "git_repos", "stale-version")
		strayFile := filepath.Join(localCache, "git_repos", git_repo.GitReposCacheVersion, "stray-file")
		danglingMeta := filepath.Join(localCache, "git_patches", GitPatchesCacheVersion, "repo", "ab", "dangling.meta.json")
		unrecognized := filepath.Join(localCache, "git_archives", GitArchivesCacheVersion, "repo", "ab", "unrecognized.txt")
		validRepo := filepath.Join(localCache, "git_repos", git_repo.GitReposCacheVersion, "valid")

		gomega.Expect(RunGC(ctx, RunGCOptions{})).To(gomega.Succeed())

		gomega.Expect(staleVersionDir).NotTo(gomega.BeADirectory())
		gomega.Expect(strayFile).NotTo(gomega.BeAnExistingFile())
		gomega.Expect(danglingMeta).NotTo(gomega.BeAnExistingFile())
		gomega.Expect(unrecognized).NotTo(gomega.BeAnExistingFile())
		gomega.Expect(validRepo).NotTo(gomega.BeADirectory())
	})

	ginkgo.It("does not refresh last access timestamps", func(ctx ginkgo.SpecContext) {
		lastAccessAtPath := filepath.Join(localCache, "git_repos", git_repo.GitReposCacheVersion, "valid", "last_access_at")

		before, err := timestamps.ReadTimestampFile(lastAccessAtPath)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		gomega.Expect(RunGC(ctx, RunGCOptions{DryRun: true})).To(gomega.Succeed())

		after, err := timestamps.ReadTimestampFile(lastAccessAtPath)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(after).To(gomega.Equal(before))
	})
})
