package true_git

import (
	"os"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/werf"
)

var _ = ginkgo.DescribeTable("read reused nested submodules",
	func(ctx ginkgo.SpecContext, legacyCache bool) {
		isolateGitConfig()
		baseDir := SuiteData.TestDirPath
		setEnvForSpec("WERF_TMP_DIR", baseDir)
		setEnvForSpec("WERF_HOME", filepath.Join(baseDir, "werf-home"))
		gomega.Expect(werf.Init(baseDir, filepath.Join(baseDir, "werf-home"))).To(gomega.Succeed())
		gomega.Expect(Init(ctx, Options{})).To(gomega.Succeed())
		leafRemote := filepath.Join(baseDir, "leaf-remote")
		gitInitRepoWithFile(ctx, leafRemote, "leaf.txt", "leaf content")
		midRemote := filepath.Join(baseDir, "mid-remote")
		gitInitRepo(ctx, midRemote)
		gitAddSubmoduleSucceed(ctx, midRemote, "../leaf-remote", ".ai-context")
		gitSucceed(ctx, midRemote, "commit", "-m", "add leaf")

		superRepo := filepath.Join(baseDir, "super")
		superGitDir := filepath.Join(superRepo, ".git")
		gitInitRepo(ctx, superRepo)
		gitAddSubmoduleSucceed(ctx, superRepo, midRemote, "installer")
		gitUpdateSubmodulesSucceed(ctx, superRepo, "--init", "--recursive")
		gitSucceed(ctx, superRepo, "commit", "-m", "add installer")
		commit := gitSucceedTrimmed(ctx, superRepo, "rev-parse", "HEAD")
		cacheDir := filepath.Join(baseDir, "cache")

		gomega.Expect(os.RemoveAll(leafRemote)).To(gomega.Succeed())
		gomega.Expect(os.RemoveAll(midRemote)).To(gomega.Succeed())
		workTreeDir, err := prepareWorkTree(ctx, superGitDir, cacheDir, commit, true)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		if legacyCache {
			gitSucceed(ctx, workTreeDir, "config", "submodule.installer.url", midRemote)
			gitSucceed(ctx, workTreeDir, "config", "--remove-section", "submodule.installer")
			gitSucceed(ctx, filepath.Join(workTreeDir, "installer"), "config", "submodule..ai-context.url", leafRemote)
			gitSucceed(ctx, filepath.Join(workTreeDir, "installer"), "config", "--remove-section", "submodule..ai-context")
			gitSucceed(ctx, filepath.Join(workTreeDir, "installer"), "remote", "set-url", "origin", filepath.Join(superGitDir, "modules", "installer"))
			gitSucceed(ctx, filepath.Join(workTreeDir, "installer", ".ai-context"), "remote", "set-url", "origin", filepath.Join(superGitDir, "modules", "installer", "modules", ".ai-context"))
			workTreeDir, err = prepareWorkTree(ctx, superGitDir, cacheDir, commit, true)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		}

		expectSubmoduleFile(superGitDir, workTreeDir, "leaf.txt", "leaf content", "installer", ".ai-context")
		gomega.Expect(gitSucceedTrimmed(ctx, filepath.Join(workTreeDir, "installer"), "config", "submodule..ai-context.url")).To(gomega.Equal(leafRemote))
		gomega.Expect(gitSucceedTrimmed(ctx, filepath.Join(workTreeDir, "installer", ".ai-context"), "remote", "get-url", "origin")).To(gomega.Equal(leafRemote))
	},
	ginkgo.Entry("cold cache with relative URLs", false),
	ginkgo.Entry("same commit cached without persistent registration", true),
)
