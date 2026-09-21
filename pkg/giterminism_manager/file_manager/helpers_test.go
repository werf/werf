package filemanager_test

import (
	"context"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"gopkg.in/yaml.v3"

	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/git_repo/gitdata"
	"github.com/werf/werf/v2/pkg/giterminism_manager"
	filemanager "github.com/werf/werf/v2/pkg/giterminism_manager/file_manager"
	"github.com/werf/werf/v2/pkg/true_git"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/test/pkg/utils"
)

func newChartFileManager(ctx context.Context, destination string, localOverride bool) (*filemanager.FileManager, string, map[string]string) {
	tmpDir := ginkgo.GinkgoT().TempDir()
	gomega.Expect(werf.Init(tmpDir, ginkgo.GinkgoT().TempDir())).To(gomega.Succeed())
	gomega.Expect(true_git.Init(ctx, true_git.Options{})).To(gomega.Succeed())
	gitDataManager, err := gitdata.GetHostGitDataManager(ctx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(git_repo.Init(gitDataManager)).To(gomega.Succeed())

	sourceDir := filepath.Join(tmpDir, "common")
	appDir := filepath.Join(tmpDir, "app")
	projectDir := filepath.Join(appDir, "ci")
	for _, repoDir := range []string{sourceDir, appDir} {
		utils.MkdirAll(repoDir)
		utils.RunSucceedCommand(ctx, repoDir, "git", "init", "--initial-branch=main")
		utils.RunSucceedCommand(ctx, repoDir, "git", "config", "user.name", "Test")
		utils.RunSucceedCommand(ctx, repoDir, "git", "config", "user.email", "test@example.com")
		utils.RunSucceedCommand(ctx, repoDir, "git", "config", "commit.gpgsign", "false")
	}

	includedFiles := map[string]string{
		"Chart.yaml":        "apiVersion: v2\nname: repro\nversion: 0.1.0\n",
		"templates/cm.yaml": "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: included\n",
		"values.yaml":       "source: include\n",
		"werf.yaml":         "project: repro\nconfigVersion: 1\ndeploy:\n  helmChartDir: .\n",
	}
	for name, data := range includedFiles {
		utils.WriteFile(filepath.Join(sourceDir, "helm-chart", name), []byte(data))
	}
	utils.WriteFile(filepath.Join(sourceDir, "sibling", "outside.yaml"), []byte("outside chart"))
	commitChartRepo(ctx, sourceDir)

	includeConfigs := []map[string]string{
		{"git": sourceDir, "branch": "main", "add": "/helm-chart", "to": destination},
	}
	if destination != "/" {
		includeConfigs = append(includeConfigs, map[string]string{
			"git": sourceDir, "branch": "main", "add": "/sibling", "to": destination + "-other",
		})
	}
	configData, err := yaml.Marshal(map[string]interface{}{
		"apiVersion": "werf/includes/v1beta1",
		"includes":   includeConfigs,
	})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	utils.WriteFile(filepath.Join(projectDir, "werf-includes.yaml"), configData)
	if localOverride {
		utils.WriteFile(filepath.Join(projectDir, destination, "values.yaml"), []byte("source: local\n"))
	}
	commitChartRepo(ctx, appDir)

	newManager := func(createLock bool) *giterminism_manager.Manager {
		repo, err := git_repo.OpenLocalRepo(ctx, "app", appDir, git_repo.OpenLocalRepoOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		manager, err := giterminism_manager.NewManager(ctx, "", projectDir, repo, utils.GetHeadCommit(ctx, appDir), giterminism_manager.NewManagerOptions{CreateIncludesLockFile: createLock})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		return manager
	}
	newManager(true)
	commitChartRepo(ctx, appDir)
	return newManager(false).FileManager, projectDir, includedFiles
}

func commitChartRepo(ctx context.Context, repoDir string) {
	utils.RunSucceedCommand(ctx, repoDir, "git", "add", ".")
	utils.RunSucceedCommand(ctx, repoDir, "git", "commit", "-m", "test chart")
}
