package image

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/cgi"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/build/stage"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/context_manager"
	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/git_repo/gitdata"
	"github.com/werf/werf/v2/pkg/giterminism_manager"
	"github.com/werf/werf/v2/pkg/true_git"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/test/pkg/utils"
)

func requireGitAttributeSource(ctx context.Context, projectDir string) {
	for _, args := range [][]string{
		{"check-attr", "--source=HEAD", "--all", "--", "app/included.txt"},
		{"var", "GIT_ATTR_SYSTEM"},
		{"var", "GIT_ATTR_GLOBAL"},
	} {
		cmd := true_git.NewGitCmd(ctx, &true_git.GitCmdOptions{RepoDir: projectDir}, args...)
		err := cmd.Run(ctx)
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 129 {
			ginkgo.Skip("requires Git attribute input discovery: git " + strings.Join(args, " "))
		}
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}
}

// The attribute query only runs for repositories with a configured filter driver, so tests about
// that query must bring their own driver instead of depending on the host's global Git LFS setup.
func configureProbeFilter(ctx context.Context, projectDir string) {
	utils.RunSucceedCommand(ctx, projectDir, "git", "config", "filter.probe.smudge", "cat")
}

func interceptGitCheckAttributes(response string) {
	interceptGitSubcommand("check-attr", response)
}

func interceptGitSubcommand(subcommand, response string) {
	git, err := exec.LookPath("git")
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	dir := ginkgo.GinkgoT().TempDir()
	script := fmt.Sprintf("#!/bin/sh\nfor arg do\n  if [ \"$arg\" = %s ]; then\n%s\n  fi\ndone\nexec %q \"$@\"\n", subcommand, response, git)
	gomega.Expect(os.WriteFile(filepath.Join(dir, "git"), []byte(script), 0o755)).To(gomega.Succeed())
	ginkgo.GinkgoT().Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func taintCachedContextAfterReset(command string) string {
	git, err := exec.LookPath("git")
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	marker := filepath.Join(ginkgo.GinkgoT().TempDir(), "worktree")
	interceptGitSubcommand("reset", fmt.Sprintf("real_git=%q\n\"$real_git\" \"$@\" || exit $?\n%s\npwd > %q\n\"$real_git\" rev-parse --git-path index > %q\ncp \"$(\"$real_git\" rev-parse --git-path index)\" %q || exit $?\nexit 0", git, command, marker, marker+".indexpath", marker+".index"))
	return marker
}

func newContentCachingRepo(ctx context.Context) string {
	projectDir := newProjectRepo(ctx, dockerfileProjectFiles("\nimage: "+projectImageName+"\ncontext: app\ndockerfile: Dockerfile\n",
		map[string]string{
			"app/Dockerfile":    "FROM scratch\nCOPY included.txt /\n",
			"app/included.txt":  "included\n",
			"app/ignored.txt":   "ignored\n",
			"app/.dockerignore": ".dockerignore\nignored.txt\n",
			"outside.txt":       "outside\n",
		}))
	utils.RunSucceedCommand(ctx, projectDir, "ln", "-s", "included.txt", "app/link.txt")
	commitFiles(ctx, projectDir, nil)

	return projectDir
}

func cachedArchive(ctx context.Context, projectDir string) (string, os.FileInfo) {
	return cachedArchiveOf(ctx, giterminismManagerOf(ctx, projectDir))
}

func cachedArchiveOf(ctx context.Context, giterminismManager *giterminism_manager.Manager) (string, os.FileInfo) {
	path := contextArchivesFor(ctx, giterminismManager, projectImageName)[0].path
	info, err := os.Stat(path)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	return path, info
}

func entryNamed(entries []tarEntry, name string) tarEntry {
	for _, entry := range entries {
		if entry.Name == name {
			return entry
		}
	}
	ginkgo.Fail("no tar entry named " + name)

	return tarEntry{}
}

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

const contextStreamingBlobSize = 1 << 20

func contextStreamingProjectFiles() map[string]string {
	files := map[string]string{"blob.bin": strings.Repeat("x", contextStreamingBlobSize)}
	var imageBlocks []string
	for _, name := range []string{"one", "two", "three"} {
		files[name+".Dockerfile"] = fmt.Sprintf("FROM scratch\nCOPY blob.bin /blob-%s\n", name)
		imageBlocks = append(imageBlocks, fmt.Sprintf("\nimage: %s\ndockerfile: %s.Dockerfile\n", name, name))
	}

	return dockerfileProjectFiles(strings.Join(imageBlocks, "---"), files)
}

func dockerfileProjectFiles(imageBlocks string, files map[string]string) map[string]string {
	projectFiles := map[string]string{"werf.yaml": "project: context-streaming\nconfigVersion: 1\n---" + imageBlocks}
	for name, data := range files {
		projectFiles[name] = data
	}

	return projectFiles
}

func giterminismDockerfileConfig(allowances map[string][]string) string {
	config := "giterminismConfigVersion: \"1\"\nconfig:\n  dockerfile:\n"
	for key, paths := range allowances {
		config += fmt.Sprintf("    %s: [%s]\n", key, strings.Join(paths, ", "))
	}

	return config
}

func giterminismManagerOf(ctx context.Context, projectDir string) *giterminism_manager.Manager {
	repo, err := git_repo.OpenLocalRepo(ctx, "own", projectDir, git_repo.OpenLocalRepoOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	manager, err := giterminism_manager.NewManager(ctx, "werf-giterminism.yaml", projectDir, repo, utils.GetHeadCommit(ctx, projectDir), giterminism_manager.NewManagerOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	return manager
}

func dockerfileContextArchives(ctx context.Context, projectDir string, imageNames ...string) []*BuildContextArchive {
	return contextArchivesFor(ctx, giterminismManagerOf(ctx, projectDir), imageNames...)
}

func contextArchivesFor(ctx context.Context, giterminismManager *giterminism_manager.Manager, imageNames ...string) []*BuildContextArchive {
	_, werfConfig, err := config.GetWerfConfig(ctx, "", "", "", giterminismManager, config.WerfConfigOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	conveyorTmpDir := filepath.Join(werf.GetTmpDir(), "conveyor")

	var archives []*BuildContextArchive
	for _, imageName := range imageNames {
		imageConfig, ok := werfConfig.GetImage(imageName).(*config.ImageFromDockerfile)
		gomega.Expect(ok).To(gomega.BeTrue(), "image %q must be a dockerfile image", imageName)

		archive := NewBuildContextArchive(giterminismManager, filepath.Join(conveyorTmpDir, "image", imageName))
		gomega.Expect(archive.Create(ctx, container_backend.BuildContextArchiveCreateOptions{
			DockerfileRelToContextPath: imageConfig.Dockerfile,
			ContextGitSubDir:           imageConfig.Context,
			ContextAddFiles:            imageConfig.ContextAddFiles,
		})).To(gomega.Succeed())

		archives = append(archives, archive)
	}

	return archives
}

func uniqueFileBytes(root string) (int64, error) {
	var counted []os.FileInfo
	var total int64

	if err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !entry.Type().IsRegular() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("stat %q: %w", path, err)
		}
		// ponytail: O(n²) SameFile scan, fine for the few dozen files of a test tree
		for _, seen := range counted {
			if os.SameFile(seen, info) {
				return nil
			}
		}
		counted = append(counted, info)
		total += info.Size()
		return nil
	}); err != nil {
		return 0, fmt.Errorf("walk %q: %w", root, err)
	}

	return total, nil
}

func filesMatching(root, pattern string) ([]string, error) {
	var matched []string

	if err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		ok, err := filepath.Match(pattern, entry.Name())
		if err != nil {
			return fmt.Errorf("match %q: %w", pattern, err)
		}
		if ok {
			matched = append(matched, path)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("walk %q: %w", root, err)
	}

	return matched, nil
}

type tarEntry struct {
	tar.Header
	Content string
}

func tarEntries(reader io.Reader) []tarEntry {
	var entries []tarEntry
	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			return entries
		}
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		content, err := io.ReadAll(tarReader)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		entries = append(entries, tarEntry{Header: *header, Content: string(content)})
	}
}

func tarFileEntries(path string) []tarEntry {
	file, err := os.Open(path)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	defer file.Close()

	return tarEntries(file)
}

func openedContextEntries(ctx context.Context, archive *BuildContextArchive) []tarEntry {
	reader, err := archive.Open(ctx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	defer reader.Close()

	return tarEntries(reader)
}

func lastEntryContents(entries []tarEntry) map[string]string {
	contents := map[string]string{}
	for _, entry := range entries {
		contents[entry.Name] = entry.Content
	}

	return contents
}

func materializedContextArchive(ctx context.Context, archive *BuildContextArchive, projectDir, contextGitSubDir string) string {
	path, err := context_manager.AddContextAddFilesToContextArchive(ctx, &context_manager.AddContextAddFilesToContextArchiveOpts{
		OriginalArchivePath:    archive.path,
		ProjectDir:             projectDir,
		ContextDir:             contextGitSubDir,
		ContextAddFilesFromMem: archive.contextAddFilesFromMem,
	})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	return path
}

func dirEntries(root string) map[string]string {
	entries := map[string]string{}

	gomega.Expect(filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if entry.Type()&fs.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			entries[relPath] = "symlink:" + target
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		entries[relPath] = fmt.Sprintf("%04o:%s", info.Mode().Perm(), data)
		return err
	})).To(gomega.Succeed())

	return entries
}

var _ Conveyor = (*preparationTestConveyor)(nil)

type preparationTestConveyor struct {
	Conveyor
	remoteMutex sync.Mutex
	remotes     map[string]*git_repo.Remote
}

var _ Conveyor = (*preparationTestConveyor)(nil)

func (c *preparationTestConveyor) GetLocalGitRepoVirtualMergeOptions() stage.VirtualMergeOptions {
	return stage.VirtualMergeOptions{}
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
		content, err := gitStage.GetNextStageDependencies(ctx, tree.Conveyor)
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
	for _, image := range remotePreparationImages(tree) {
		for _, remote := range image.(config.StapelImageInterface).ImageBaseConfig().Git.Remote {
			remote.Commit = ""
			remote.Branch = "main"
		}
	}
}

func remotePreparationImages(tree *ImagesTree) []config.ImageInterface {
	sets, err := tree.werfConfig.GroupImagesByIndependentSets(tree.ImagesToProcess)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	var images []config.ImageInterface
	for _, set := range sets {
		images = append(images, set...)
	}
	return images
}
