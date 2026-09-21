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

func newFileURLChartFileManager(ctx context.Context, withIncludes bool) *filemanager.FileManager {
	tmpDir := ginkgo.GinkgoT().TempDir()
	gomega.Expect(werf.Init(tmpDir, ginkgo.GinkgoT().TempDir())).To(gomega.Succeed())
	gomega.Expect(true_git.Init(ctx, true_git.Options{})).To(gomega.Succeed())
	gitDataManager, err := gitdata.GetHostGitDataManager(ctx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(git_repo.Init(gitDataManager)).To(gomega.Succeed())

	sourceDir := filepath.Join(tmpDir, "common")
	appDir := filepath.Join(tmpDir, "app")
	projectDir := filepath.Join(appDir, "ci")
	repoDirs := []string{appDir}
	if withIncludes {
		repoDirs = append(repoDirs, sourceDir)
	}
	for _, repoDir := range repoDirs {
		utils.MkdirAll(repoDir)
		utils.RunSucceedCommand(ctx, repoDir, "git", "init", "--initial-branch=main")
		utils.RunSucceedCommand(ctx, repoDir, "git", "config", "user.name", "Test")
		utils.RunSucceedCommand(ctx, repoDir, "git", "config", "user.email", "test@example.com")
		utils.RunSucceedCommand(ctx, repoDir, "git", "config", "commit.gpgsign", "false")
	}

	projectFiles := map[string]string{
		".helm/.helmignore":                      "root-dropped.yaml\nsubchart-kept.yaml\n",
		".helm/Chart.yaml":                       "apiVersion: v2\nname: root\nversion: 0.1.0\ndependencies:\n- name: sub\n  version: 0.1.0\n  repository: file://../sub-chart\n",
		".helm/Chart.lock":                       "dependencies:\n- name: sub\n  version: 0.1.0\n  repository: file://../sub-chart\n",
		".helm/templates/root-kept.yaml":         "root kept",
		".helm/templates/root-dropped.yaml":      "root dropped",
		"sub-chart/.helmignore":                  "notes.txt\ndropped-by-dep.txt\n",
		"sub-chart/Chart.yaml":                   "apiVersion: v2\nname: sub\nversion: 0.1.0\n",
		"sub-chart/templates/subchart-kept.yaml": "subchart kept",
		"sub-chart/notes.txt":                    "subchart dropped",
	}
	for name, data := range projectFiles {
		utils.WriteFile(filepath.Join(projectDir, name), []byte(data))
	}

	newManager := func(createLock bool) *filemanager.FileManager {
		repo, err := git_repo.OpenLocalRepo(ctx, "app", appDir, git_repo.OpenLocalRepoOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		manager, err := giterminism_manager.NewManager(ctx, "", projectDir, repo, utils.GetHeadCommit(ctx, appDir), giterminism_manager.NewManagerOptions{CreateIncludesLockFile: createLock})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		return manager.FileManager
	}

	if !withIncludes {
		commitChartRepo(ctx, appDir)
		return newManager(false)
	}

	includedFiles := map[string]string{
		"templates/from-include.yaml": "subchart from include",
		"dropped-by-dep.txt":          "subchart include dropped",
	}
	for name, data := range includedFiles {
		utils.WriteFile(filepath.Join(sourceDir, "dep-extra", name), []byte(data))
	}
	commitChartRepo(ctx, sourceDir)

	configData, err := yaml.Marshal(map[string]interface{}{
		"apiVersion": "werf/includes/v1beta1",
		"includes": []map[string]string{
			{"git": sourceDir, "branch": "main", "add": "/dep-extra", "to": "/sub-chart"},
		},
	})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	utils.WriteFile(filepath.Join(projectDir, "werf-includes.yaml"), configData)
	commitChartRepo(ctx, appDir)

	newManager(true)
	commitChartRepo(ctx, appDir)
	return newManager(false)
}
