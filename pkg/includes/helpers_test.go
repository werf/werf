package includes

import (
	"context"
	"os"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/git_repo/gitdata"
	"github.com/werf/werf/v2/pkg/true_git"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/test/pkg/utils"
)

func setupRemoteInclude(ctx context.Context) (Config, *gitRepositoriesWithCache, string) {
	home := ginkgo.GinkgoT().TempDir()
	gomega.Expect(werf.Init(home, ginkgo.GinkgoT().TempDir())).To(gomega.Succeed())
	gomega.Expect(true_git.Init(ctx, true_git.Options{})).To(gomega.Succeed())
	manager, err := gitdata.GetHostGitDataManager(ctx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	oldManager := git_repo.CommonGitDataManager
	ginkgo.DeferCleanup(func() { git_repo.CommonGitDataManager = oldManager })
	gomega.Expect(git_repo.Init(manager)).To(gomega.Succeed())
	source := filepath.Join(home, "source")
	gomega.Expect(os.Mkdir(source, 0o700)).To(gomega.Succeed())
	utils.RunSucceedCommand(ctx, source, "git", "-c", "init.defaultBranch=main", "init")
	utils.RunSucceedCommand(ctx, source, "git", "config", "uploadpack.allowAnySHA1InWant", "true")
	utils.WriteFile(filepath.Join(source, "data.txt"), []byte("v1"))
	utils.RunSucceedCommand(ctx, source, "git", "add", "data.txt")
	utils.RunSucceedCommand(ctx, source, "git", "-c", "user.name=Test", "-c", "user.email=test@example.com", "-c", "commit.gpgsign=false", "commit", "-m", "v1")
	commit := utils.GetHeadCommit(ctx, source)
	cfg := Config{Includes: []includeConf{{Git: "file://" + source, Branch: "main", Add: "/", To: "/"}}}
	repos, err := initRemoteRepos(ctx, cfg)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return cfg, repos, commit
}
