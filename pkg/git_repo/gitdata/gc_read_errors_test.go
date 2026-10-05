package gitdata

import (
	"os"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/git_repo"
	"github.com/werf/werf/v3/pkg/werf"
)

var _ = ginkgo.Describe("Git cache access errors", func() {
	ginkgo.BeforeEach(func() {
		ginkgo.GinkgoT().Setenv("WERF_HOME", ginkgo.GinkgoT().TempDir())
		ginkgo.GinkgoT().Setenv("WERF_TMP_DIR", ginkgo.GinkgoT().TempDir())
		gomega.Expect(werf.Init("", "")).To(gomega.Succeed())
		gcFixture()
	})
	ginkgo.DescribeTable("preserves a repository whose access marker cannot be read",
		func(ctx ginkgo.SpecContext, relative string) {
			dir := filepath.Join(werf.GetLocalCacheDir(), relative)
			marker := filepath.Join(dir, "last_access_at")
			gomega.Expect(os.Remove(marker)).To(gomega.Succeed())
			gomega.Expect(os.Symlink("last_access_at", marker)).To(gomega.Succeed())
			gomega.Expect(RunGC(ctx, RunGCOptions{})).NotTo(gomega.Succeed())
			gomega.Expect(dir).To(gomega.BeADirectory())
			target, err := os.Readlink(marker)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(target).To(gomega.Equal("last_access_at"))
		},
		ginkgo.Entry("full", filepath.Join("git_repos", git_repo.GitReposCacheVersion, "valid")),
		ginkgo.Entry("shallow", filepath.Join("git_mirrors", git_repo.GitMirrorsCacheVersion, "valid", "shallow")),
	)
})
