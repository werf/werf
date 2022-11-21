package image

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/containers/buildah/copier"

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

	dockerIgnorePathMatcher, err := createDockerIgnorePathMatcher(ctx, a.giterminismMgr, opts.ContextGitSubDir, opts.DockerfileRelToContextPath)
	if err != nil {
		return fmt.Errorf("unable to create dockerignore path matcher: %w", err)
	}

	archive, err := a.giterminismMgr.LocalGitRepo().GetOrCreateArchive(ctx, git_repo.ArchiveOptions{
		PathScope: contextPathRelativeToGitWorkTree,
		PathMatcher: path_matcher.NewMultiPathMatcher(path_matcher.NewPathMatcher(
			path_matcher.PathMatcherOptions{BasePath: contextPathRelativeToGitWorkTree}),
			dockerIgnorePathMatcher,
		),
		Commit: a.giterminismMgr.HeadCommit(),
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
	if a.path == "" {
		panic("extract should not be called before create")
	}

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

func (a *BuildContextArchive) CalculateGlobsChecksum(ctx context.Context, globs []string, checkForArchives bool) (string, error) {
	contextDir, err := a.ExtractOrGetExtractedDir(ctx)
	if err != nil {
		return "", fmt.Errorf("unable to get build context dir: %w", err)
	}

	globStats, err := copier.Stat(contextDir, contextDir, copier.StatOptions{CheckForArchives: checkForArchives}, globs)
	if err != nil {
		return "", fmt.Errorf("unable to stat globs: %w", err)
	}
	if len(globStats) == 0 {
		return "", fmt.Errorf("no glob matches for globs: %v", globs)
	}

	var matches []string
	for _, globStat := range globStats {
		if globStat.Error != "" {
			return "", fmt.Errorf("unable to stat glob %q: %s", globStat.Glob, globStat.Error)
		}

		for _, match := range globStat.Globbed {
			matches = append(matches, match)
		}
	}

	pathsChecksum, err := a.CalculatePathsChecksum(ctx, matches)
	if err != nil {
		return "", fmt.Errorf("unable to calculate build context paths checksum: %w", err)
	}

	return pathsChecksum, nil
}

func (a *BuildContextArchive) CalculatePathsChecksum(ctx context.Context, paths []string) (string, error) {
	sort.Strings(paths)
	paths = util.UniqStrings(paths)

	dir, err := a.ExtractOrGetExtractedDir(ctx)
	if err != nil {
		return "", fmt.Errorf("unable to access context directory: %w", err)
	}

	var pathsHashes []string
	for _, path := range paths {
		p := filepath.Join(dir, path)

		hash, err := util.HashContentsAndPathsRecurse(p)
		if err != nil {
			return "", fmt.Errorf("unable to calculate hash: %w", err)
		}

		pathsHashes = append(pathsHashes, hash)
	}

	return util.Sha256Hash(pathsHashes...), nil
}
