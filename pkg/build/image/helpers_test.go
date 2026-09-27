package image

import (
	"context"
	"fmt"
	"net/http"
	"net/http/cgi"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/build/stage"
	"github.com/werf/werf/v3/pkg/config"
	"github.com/werf/werf/v3/pkg/git_repo"
	"github.com/werf/werf/v3/pkg/git_repo/gitdata"
	"github.com/werf/werf/v3/pkg/giterminism_manager"
	"github.com/werf/werf/v3/pkg/true_git"
	"github.com/werf/werf/v3/pkg/werf"
	"github.com/werf/werf/v3/test/pkg/utils"
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

var _ Conveyor = (*preparationTestConveyor)(nil)

type preparationTestConveyor struct {
	Conveyor
	remoteMutex sync.Mutex
	remotes     map[string]*git_repo.Remote
}

func (c *preparationTestConveyor) GetForcedTargetPlatforms() []string { return nil }
func (c *preparationTestConveyor) GetTargetPlatforms() ([]string, error) {
	return []string{"linux/amd64", "linux/arm64"}, nil
}

func (c *preparationTestConveyor) GetImageTargetPlatforms(string) ([]string, error) {
	return nil, nil
}

func (c *preparationTestConveyor) GetRemoteGitRepo(key string) *git_repo.Remote {
	c.remoteMutex.Lock()
	defer c.remoteMutex.Unlock()
	return c.remotes[key]
}

func (c *preparationTestConveyor) SetRemoteGitRepo(key string, repo *git_repo.Remote) {
	c.remoteMutex.Lock()
	defer c.remoteMutex.Unlock()
	if c.remotes == nil {
		c.remotes = make(map[string]*git_repo.Remote)
	}
	c.remotes[key] = repo
}

func newRemotePreparationTree(ctx context.Context, wrap func(http.Handler) http.Handler) *ImagesTree {
	origin := newProjectRepo(ctx, map[string]string{"data": "remote data"})
	firstCommit := utils.GetHeadCommit(ctx, origin)
	commitFiles(ctx, origin, map[string]string{"data": "updated remote data"})
	root := ginkgo.GinkgoT().TempDir()
	for _, name := range []string{"one", "two", "three"} {
		utils.RunSucceedCommand(ctx, origin, "git", "clone", "--bare", origin, filepath.Join(root, name+".git"))
	}
	git, err := exec.LookPath("git")
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	backend := &cgi.Handler{Path: git, Args: []string{"http-backend"}, Env: []string{"GIT_PROJECT_ROOT=" + root, "GIT_HTTP_EXPORT_ALL=1"}}

	server := httptest.NewServer(wrap(backend))
	ginkgo.DeferCleanup(server.Close)
	data, err := os.ReadFile("testdata/remote-preparation/werf.yaml.tmpl")
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	commitFiles(ctx, origin, map[string]string{"werf.yaml": fmt.Sprintf(string(data), server.URL, firstCommit)})
	_, _, opts := stapelImageConfig(ctx, origin)
	_, cfg, err := config.GetWerfConfig(ctx, "", "", "", opts.GiterminismManager, config.WerfConfigOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	opts.Conveyor = &preparationTestConveyor{}
	selected, err := config.NewImagesToProcess(cfg, []string{"app", "other"}, false, false)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return NewImagesTree(cfg, ImagesTreeOptions{CommonImageOptions: opts, ImagesToProcess: selected, RemoteGitTasksLimit: 2})
}

func remotePreparationInputs(ctx context.Context, tree *ImagesTree) map[string][]string {
	inputs := make(map[string][]string)
	for _, image := range tree.GetImages() {
		gitStage := image.GetStage(stage.GitArchive)
		gomega.Expect(gitStage).NotTo(gomega.BeNil())
		mappings := gitStage.GetGitMappings()
		gomega.Expect(mappings).To(gomega.HaveLen(3))
		key := image.Name + "/" + image.TargetPlatform
		var values []string
		for _, mapping := range mappings {
			commit, err := mapping.GetLatestCommitInfo(ctx, tree.Conveyor)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			values = append(values, mapping.To+":"+commit.Commit+":"+mapping.GetParamshash())
		}
		content, err := gitStage.GetContentDependencies(ctx, tree.Conveyor, nil)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		inputs[key] = append(values, content)
	}
	gomega.Expect(inputs).To(gomega.HaveLen(4))
	for _, key := range []string{"app/linux/amd64", "app/linux/arm64", "other/linux/amd64", "other/linux/arm64"} {
		gomega.Expect(inputs).To(gomega.HaveKey(key))
	}
	return inputs
}

func useRemotePreparationBranches(tree *ImagesTree) {
	for _, image := range tree.werfConfig.GetImagesForProcessing(tree.ImagesToProcess) {
		for _, remote := range image.(config.StapelImageInterface).ImageBaseConfig().Git.Remote {
			remote.Commit = ""
			remote.Branch = "main"
		}
	}
}
