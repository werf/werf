package image

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/build/stage"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/git_repo/gitdata"
	"github.com/werf/werf/v2/pkg/giterminism_manager"
	"github.com/werf/werf/v2/pkg/true_git"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/test/pkg/utils"
)

func newProjectRepo(ctx context.Context, files map[string]string) string {
	tmpDir := ginkgo.GinkgoT().TempDir()
	homeDir := ginkgo.GinkgoT().TempDir()
	ginkgo.GinkgoT().Setenv("WERF_TMP_DIR", tmpDir)
	ginkgo.GinkgoT().Setenv("WERF_HOME", homeDir)
	gomega.Expect(werf.Init(tmpDir, homeDir)).To(gomega.Succeed())
	gomega.Expect(true_git.Init(ctx, true_git.Options{})).To(gomega.Succeed())
	gitDataManager, err := gitdata.GetHostGitDataManager(ctx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(git_repo.Init(gitDataManager)).To(gomega.Succeed())

	projectDir := filepath.Join(tmpDir, "project")
	utils.MkdirAll(projectDir)
	utils.RunSucceedCommand(ctx, projectDir, "git", "init", "--initial-branch=main")
	utils.RunSucceedCommand(ctx, projectDir, "git", "config", "user.name", "Test")
	utils.RunSucceedCommand(ctx, projectDir, "git", "config", "user.email", "test@example.com")
	utils.RunSucceedCommand(ctx, projectDir, "git", "config", "commit.gpgsign", "false")

	commitFiles(ctx, projectDir, files)

	return projectDir
}

func commitFiles(ctx context.Context, projectDir string, files map[string]string) {
	for name, data := range files {
		utils.WriteFile(filepath.Join(projectDir, name), []byte(data))
	}

	utils.RunSucceedCommand(ctx, projectDir, "git", "add", ".")
	utils.RunSucceedCommand(ctx, projectDir, "git", "commit", "-m", "test state")
}

const projectImageName = "app"

func stapelImageConfig(ctx context.Context, projectDir string) (*config.Meta, config.StapelImageInterface, CommonImageOptions) {
	repo, err := git_repo.OpenLocalRepo(ctx, "own", projectDir, git_repo.OpenLocalRepoOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	giterminismManager, err := giterminism_manager.NewManager(ctx, "", projectDir, repo, utils.GetHeadCommit(ctx, projectDir), giterminism_manager.NewManagerOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	_, werfConfig, err := config.GetWerfConfig(ctx, "", "", "", giterminismManager, config.WerfConfigOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	imageConfig, ok := werfConfig.GetImage(projectImageName).(config.StapelImageInterface)
	gomega.Expect(ok).To(gomega.BeTrue())

	opts := CommonImageOptions{
		GiterminismManager: giterminismManager,
		ProjectDir:         projectDir,
		ProjectName:        werfConfig.Meta.Project,
		ContainerWerfDir:   "/.werf",
		TmpDir:             ginkgo.GinkgoT().TempDir(),
	}

	return werfConfig.Meta, imageConfig, opts
}

func gitMappingsOf(ctx context.Context, projectDir string) []*stage.GitMapping {
	metaConfig, imageConfig, opts := stapelImageConfig(ctx, projectDir)

	gitMappings, err := generateGitMappings(ctx, metaConfig, imageConfig.ImageBaseConfig(), opts)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	return gitMappings
}

func userStageChecksums(ctx context.Context, projectDir string) map[stage.StageName][]string {
	gitMappings := gitMappingsOf(ctx, projectDir)

	checksums := map[stage.StageName][]string{}
	for _, stageName := range []stage.StageName{stage.Install, stage.BeforeSetup, stage.Setup} {
		for _, gitMapping := range gitMappings {
			checksum, err := gitMapping.StageDependenciesChecksum(ctx, nil, stageName)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			checksums[stageName] = append(checksums[stageName], checksum)
		}
	}

	return checksums
}

func stapelImage(ctx context.Context, projectDir string) (*Image, error) {
	metaConfig, imageConfig, opts := stapelImageConfig(ctx, projectDir)

	return mapStapelConfigToImage(ctx, metaConfig, imageConfig, "linux/amd64", false, opts)
}

func stageNames(image *Image) []stage.StageName {
	var names []stage.StageName
	for _, s := range image.GetStages() {
		names = append(names, s.Name())
	}

	return names
}

func beforeInstallDependencies(ctx context.Context, projectDir string) (string, error) {
	image, err := stapelImage(ctx, projectDir)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	beforeInstallStage := image.GetStage(stage.BeforeInstall)
	gomega.Expect(beforeInstallStage).NotTo(gomega.BeNil())

	return beforeInstallStage.GetDependencies(ctx, nil, nil, nil, nil, nil)
}

func stageDependenciesFixture(stageDependenciesBlock string) map[string]string {
	return map[string]string{
		"werf.yaml":           fmt.Sprintf(stageDependenciesFixtureYaml, stageDependenciesBlock),
		"src/root.txt":        "root\n",
		"src/app/main.go":     "package main\n",
		"src/cfg/values.yaml": "key: value\n",
		"outside.txt":         "outside\n",
	}
}

func changedChecksums(ctx context.Context, projectDir string, edits map[string]string) []string {
	before := userStageChecksums(ctx, projectDir)
	commitFiles(ctx, projectDir, edits)
	after := userStageChecksums(ctx, projectDir)

	var changed []string
	for stageName, checksums := range after {
		for i, checksum := range checksums {
			if checksum != before[stageName][i] {
				changed = append(changed, fmt.Sprintf("%s[%d]", stageName, i))
			}
		}
	}

	return changed
}
