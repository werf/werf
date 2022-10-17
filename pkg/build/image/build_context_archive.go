package image

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/docker/docker/builder/dockerignore"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/context_manager"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/giterminism_manager"
	"github.com/werf/werf/pkg/path_matcher"
	"github.com/werf/werf/pkg/util"
)

func NewBuildContextArchive(giterminismMgr giterminism_manager.Interface, extractionRootTmpDir string) *BuildContextArchive {
	return &BuildContextArchive{
		giterminismMgr:       giterminismMgr,
		extractionRootTmpDir: extractionRootTmpDir,
	}
}

type BuildContextArchive struct {
	giterminismMgr       giterminism_manager.Interface
	path                 string
	extractionRootTmpDir string
	extractionDir        string
}

func (a *BuildContextArchive) Create(ctx context.Context, opts container_backend.BuildContextArchiveCreateOptions) error {
	contextPathRelativeToGitWorkTree := filepath.Join(a.giterminismMgr.RelativeToGitProjectDir(), opts.ContextGitSubDir)

	pathMatcher := path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{BasePath: contextPathRelativeToGitWorkTree})
	if dockerIgnorePathMatcher, err := createDockerIgnorePathMatcher(ctx, a.giterminismMgr, opts.ContextGitSubDir, opts.DockerfileRelToContextPath); err != nil {
		return fmt.Errorf("unable to create dockerignore path matcher: %w", err)
	} else if dockerIgnorePathMatcher != nil {
		pathMatcher = path_matcher.NewMultiPathMatcher(pathMatcher, dockerIgnorePathMatcher)
	}

	archive, err := a.giterminismMgr.LocalGitRepo().GetOrCreateArchive(ctx, git_repo.ArchiveOptions{
		PathScope:   contextPathRelativeToGitWorkTree,
		PathMatcher: pathMatcher,
		Commit:      a.giterminismMgr.HeadCommit(),
	})
	if err != nil {
		return fmt.Errorf("unable to get or create archive: %w", err)
	}

	a.path = archive.GetFilePath()

	if len(opts.ContextAddFiles) > 0 {
		if err := logboek.Context(ctx).Debug().LogProcess("Add contextAddFiles to build context archive %s", a.path).DoError(func() error {
			a.path, err = context_manager.AddContextAddFilesToContextArchive(ctx, a.path, a.giterminismMgr.ProjectDir(), opts.ContextGitSubDir, opts.ContextAddFiles)
			return err
		}); err != nil {
			return fmt.Errorf("unable to add contextAddFiles to build context archive %s: %w", a.path, err)
		}
	}

	return nil
}

func (a *BuildContextArchive) Path() string {
	return a.path
}

func (a *BuildContextArchive) ExtractOrGetExtractedDir(ctx context.Context) (string, error) {
	if a.extractionDir != "" {
		return a.extractionDir, nil
	}

	if err := os.MkdirAll(a.extractionRootTmpDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("unable to create extraction root tmp dir %q: %w", a.extractionRootTmpDir, err)
	}

	var err error
	a.extractionDir, err = ioutil.TempDir(a.extractionRootTmpDir, "context")
	if err != nil {
		return "", fmt.Errorf("unable to create context tmp dir: %w", err)
	}

	archiveReader, err := os.Open(a.path)
	if err != nil {
		return "", fmt.Errorf("unable to open context archive %q: %w", a.path, err)
	}
	defer archiveReader.Close()

	if err := util.ExtractTar(archiveReader, a.extractionDir, util.ExtractTarOptions{}); err != nil {
		return "", fmt.Errorf("unable to extract context tar to tmp context dir %q: %w", a.extractionDir, err)
	}

	return a.extractionDir, nil
}

func (a *BuildContextArchive) CleanupExtractedDir(ctx context.Context) {
	if a.extractionDir == "" {
		return
	}

	if err := os.RemoveAll(a.extractionDir); err != nil {
		logboek.Context(ctx).Warn().LogF("WARNING: unable to remove extracted context dir %q: %s", a.extractionDir, err)
	}
}

// Might return nil.
func createDockerIgnorePathMatcher(ctx context.Context, giterminismMgr giterminism_manager.Interface, contextGitSubDir, dockerfileRelToContextPath string) (path_matcher.PathMatcher, error) {
	dockerfileRelToGitPath := filepath.Join(contextGitSubDir, dockerfileRelToContextPath)

	var dockerIgnorePatterns []string
	for _, dockerIgnoreRelToContextPath := range []string{
		dockerfileRelToContextPath + ".dockerignore",
		".dockerignore",
	} {
		dockerIgnoreRelToGitPath := filepath.Join(contextGitSubDir, dockerIgnoreRelToContextPath)
		if exist, err := giterminismMgr.FileReader().IsDockerignoreExistAnywhere(ctx, dockerIgnoreRelToGitPath); err != nil {
			return nil, err
		} else if !exist {
			continue
		}

		dockerIgnore, err := giterminismMgr.FileReader().ReadDockerignore(ctx, dockerIgnoreRelToGitPath)
		if err != nil {
			return nil, err
		}

		r := bytes.NewReader(dockerIgnore)
		dockerIgnorePatterns, err = dockerignore.ReadAll(r)
		if err != nil {
			return nil, fmt.Errorf("unable to read %q file: %w", dockerIgnoreRelToContextPath, err)
		}

		break
	}

	if dockerIgnorePatterns == nil {
		return nil, nil
	}

	dockerIgnorePathMatcher := path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{
		BasePath:             filepath.Join(giterminismMgr.RelativeToGitProjectDir(), contextGitSubDir),
		DockerignorePatterns: dockerIgnorePatterns,
	})

	if !dockerIgnorePathMatcher.IsPathMatched(dockerfileRelToGitPath) {
		logboek.Context(ctx).Warn().LogLn("WARNING: There is no way to ignore the Dockerfile due to docker limitation when building an image for a compressed context that reads from STDIN.")
		logboek.Context(ctx).Warn().LogF("WARNING: To hide this message, remove the Dockerfile ignore rule or add an exception rule.\n")

		exceptionRule := "!" + dockerfileRelToContextPath
		dockerIgnorePatterns = append(dockerIgnorePatterns, exceptionRule)
		dockerIgnorePathMatcher = path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{
			BasePath:             filepath.Join(giterminismMgr.RelativeToGitProjectDir(), contextGitSubDir),
			DockerignorePatterns: dockerIgnorePatterns,
		})
	}

	return dockerIgnorePathMatcher, nil
}
