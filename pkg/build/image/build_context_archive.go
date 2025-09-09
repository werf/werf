package image

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/containers/buildah/copier"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/context_manager"
	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/giterminism_manager"
	"github.com/werf/werf/v2/pkg/path_matcher"
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

	dockerIgnorePathMatcher, err := createDockerIgnorePathMatcher(ctx, *a.giterminismMgr.(*giterminism_manager.Manager), opts.ContextGitSubDir, opts.DockerfileRelToContextPath)
	if err != nil {
		return fmt.Errorf("unable to create dockerignore path matcher: %w", err)
	}

	archive, err := a.giterminismMgr.LocalGitRepo().GetOrCreateArchive(ctx, git_repo.ArchiveOptions{
		PathScope: contextPathRelativeToGitWorkTree,
		PathMatcher: path_matcher.NewMultiPathMatcher(path_matcher.NewPathMatcher(
			path_matcher.PathMatcherOptions{BasePath: contextPathRelativeToGitWorkTree}),
			dockerIgnorePathMatcher,
		),
		Commit: a.giterminismMgr.HeadCommit(ctx),
	})
	if err != nil {
		return fmt.Errorf("unable to get or create archive: %w", err)
	}

	a.path = archive.GetFilePath()

	addFilesFromMem := make(map[string][]byte)

	if opts.DockerfileRelToContextPath != "" {
		dockerFilePath := filepath.Join(opts.ContextGitSubDir, opts.DockerfileRelToContextPath)
		gm := a.giterminismMgr.(*giterminism_manager.Manager)
		dockerFileContent, err := gm.FileManager.ReadDockerfile(ctx, dockerFilePath)
		if err != nil {
			return fmt.Errorf("unable to read dockerfile %q: %w", opts.DockerfileRelToContextPath, err)
		}
		addFilesFromMem[opts.DockerfileRelToContextPath] = dockerFileContent
	}

	if len(opts.ContextAddFiles) > 0 || len(addFilesFromMem) > 0 {
		if err := logboek.Context(ctx).Debug().LogProcess("Add contextAddFiles to build context archive %s", a.path).DoError(func() error {
			a.path, err = context_manager.AddContextAddFilesToContextArchive(ctx, &context_manager.AddContextAddFilesToContextArchiveOpts{
				OriginalArchivePath:    a.path,
				ProjectDir:             a.giterminismMgr.ProjectDir(),
				ContextDir:             opts.ContextGitSubDir,
				ContextAddFiles:        opts.ContextAddFiles,
				ContextAddFilesFromMem: addFilesFromMem,
			})
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

	var contextGlobs []string
	for _, glob := range globs {
		contextGlobs = append(contextGlobs, filepath.Join(contextDir, glob))
	}
	logboek.Context(ctx).Debug().LogF("Calculating checksum for globs %v in context dir %q: will scan following dirs globs: %v\n", globs, contextDir, contextGlobs)

	globStats, err := copier.Stat(contextDir, contextDir, copier.StatOptions{CheckForArchives: checkForArchives}, contextGlobs)
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
			relMatch := util.GetRelativeToBaseFilepath(contextDir, match)
			logboek.Context(ctx).Debug().LogF("Calculating checksum for globs %v in context dir %q: matched path %q\n", globs, contextDir, relMatch)
			matches = append(matches, relMatch)
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
