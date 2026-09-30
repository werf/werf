package image

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"go.podman.io/buildah/copier"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v3/pkg/container_backend"
	"github.com/werf/werf/v3/pkg/context_manager"
	"github.com/werf/werf/v3/pkg/git_repo"
	"github.com/werf/werf/v3/pkg/giterminism_manager"
	"github.com/werf/werf/v3/pkg/opstats"
	"github.com/werf/werf/v3/pkg/path_matcher"
	"github.com/werf/werf/v3/pkg/tmp_manager"
	"github.com/werf/werf/v3/pkg/werf"
)

// createContextPinDir is a variable so that tests can put the pin on another filesystem.
var createContextPinDir = tmp_manager.CreateContextPinDir

var _ container_backend.BuildContextArchiver = (*BuildContextArchive)(nil)

func NewBuildContextArchive(giterminismMgr giterminism_manager.Interface, extractionRootTmpDir string) *BuildContextArchive {
	return &BuildContextArchive{
		giterminismMgr:       giterminismMgr,
		extractionRootTmpDir: extractionRootTmpDir,
	}
}

type BuildContextArchive struct {
	giterminismMgr         giterminism_manager.Interface
	path                   string
	extractionRootTmpDir   string
	extractionDir          string
	contextAddFilesFromMem map[string][]byte
}

func (a *BuildContextArchive) Create(ctx context.Context, opts container_backend.BuildContextArchiveCreateOptions) error {
	contextPathRelativeToGitWorkTree := filepath.Join(a.giterminismMgr.RelativeToGitProjectDir(), opts.ContextGitSubDir)

	dockerIgnorePathMatcher, err := createDockerIgnorePathMatcher(ctx, *a.giterminismMgr.(*giterminism_manager.Manager), opts.ContextGitSubDir, opts.DockerfileRelToContextPath)
	if err != nil {
		return fmt.Errorf("unable to create dockerignore path matcher: %w", err)
	}

	lock, err := git_repo.CommonGitDataManager.LockGC(ctx, true)
	if err != nil {
		return fmt.Errorf("lock git archive cache: %w", err)
	}
	defer werf.HostLocker().ReleaseLock(lock)

	archiveOptions := git_repo.ArchiveOptions{
		PathScope: contextPathRelativeToGitWorkTree,
		PathMatcher: path_matcher.NewMultiPathMatcher(path_matcher.NewPathMatcher(
			path_matcher.PathMatcherOptions{BasePath: contextPathRelativeToGitWorkTree}),
			dockerIgnorePathMatcher,
		),
		Commit: a.giterminismMgr.HeadCommit(ctx),
	}
	if repo, ok := a.giterminismMgr.LocalGitRepo().(*git_repo.Local); ok {
		archiveOptions.ContentChecksum, err = repo.GetDockerfileContextChecksum(ctx, git_repo.ChecksumOptions{
			Commit: archiveOptions.Commit,
			LsTreeOptions: git_repo.LsTreeOptions{
				PathScope:   archiveOptions.PathScope,
				PathMatcher: archiveOptions.PathMatcher,
				AllFiles:    true,
			},
		})
		if err != nil {
			return fmt.Errorf("calculate build context checksum: %w", err)
		}
	}

	archive, err := a.giterminismMgr.LocalGitRepo().GetOrCreateArchive(ctx, archiveOptions)
	if err != nil {
		return fmt.Errorf("unable to get or create archive: %w", err)
	}

	a.path = archive.GetFilePath()
	a.contextAddFilesFromMem = nil

	addFilesFromMem := make(map[string][]byte)

	if opts.DockerfileRelToContextPath != "" && filepath.IsLocal(opts.DockerfileRelToContextPath) {
		dockerFilePath := filepath.Join(opts.ContextGitSubDir, opts.DockerfileRelToContextPath)
		gm := a.giterminismMgr.(*giterminism_manager.Manager)
		dockerFileContent, err := gm.FileManager.ReadDockerfile(ctx, dockerFilePath)
		if err != nil {
			return fmt.Errorf("unable to read dockerfile %q: %w", opts.DockerfileRelToContextPath, err)
		}
		addFilesFromMem[opts.DockerfileRelToContextPath] = dockerFileContent
	}

	if len(opts.ContextAddFiles) == 0 {
		dir, err := createContextPinDir(ctx)
		if err != nil {
			return fmt.Errorf("create context archive directory: %w", err)
		}

		// Pin the cached inode against GC without copying it. Other filesystems fall back to a private copy.
		pinnedPath := filepath.Join(dir, "archive.tar")
		if err := os.Link(a.path, pinnedPath); err == nil {
			a.path = pinnedPath
			a.contextAddFilesFromMem = addFilesFromMem
			return nil
		} else {
			logboek.Context(ctx).Debug().LogF("Unable to hard-link build context, falling back to a copy: %s\n", err)
		}
		if err := os.Remove(dir); err != nil {
			return fmt.Errorf("remove unused context archive directory: %w", err)
		}
	}

	if err := logboek.Context(ctx).Debug().LogProcess("Add contextAddFiles to build context archive %s", a.path).DoError(func() error {
		defer opstats.Observe(ctx, opstats.OperationContextAddFiles)()
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

	return nil
}

func (a *BuildContextArchive) Path() string {
	return a.path
}

// Open returns an independent context stream, including Dockerfile overrides. The caller must close it.
func (a *BuildContextArchive) Open(ctx context.Context) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	source, err := os.Open(a.path)
	if err != nil {
		return nil, fmt.Errorf("open context archive %q: %w", a.path, err)
	}
	if len(a.contextAddFilesFromMem) == 0 {
		return source, nil
	}

	reader, writer := io.Pipe()
	stop := context.AfterFunc(ctx, func() {
		writer.CloseWithError(ctx.Err())
	})
	go func() {
		defer stop()
		err := func() error {
			tw := tar.NewWriter(writer)
			if err := util.CopyTar(ctx, source, tw, util.CopyTarOptions{}); err != nil {
				return fmt.Errorf("stream context archive: %w", err)
			}
			for name, data := range a.contextAddFilesFromMem {
				if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o600, Size: int64(len(data))}); err != nil {
					return fmt.Errorf("write Dockerfile header: %w", err)
				}
				if _, err := tw.Write(data); err != nil {
					return fmt.Errorf("write Dockerfile contents: %w", err)
				}
			}
			return tw.Close()
		}()
		writer.CloseWithError(errors.Join(err, source.Close()))
	}()
	return reader, nil
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

	archiveReader, err := a.Open(ctx)
	if err != nil {
		a.CleanupExtractedDir(ctx)
		return "", fmt.Errorf("open build context: %w", err)
	}
	defer archiveReader.Close()

	if err := util.ExtractTar(archiveReader, a.extractionDir, util.ExtractTarOptions{}); err != nil {
		err = fmt.Errorf("unable to extract context tar to tmp context dir %q: %w", a.extractionDir, err)
		a.CleanupExtractedDir(ctx)
		return "", err
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
	a.extractionDir = ""
}

func (a *BuildContextArchive) CalculateGlobsChecksum(ctx context.Context, globs []string, opts container_backend.CalculateGlobsChecksumOptions) (string, error) {
	contextDir, err := a.ExtractOrGetExtractedDir(ctx)
	if err != nil {
		return "", fmt.Errorf("unable to get build context dir: %w", err)
	}

	var contextGlobs []string
	for _, glob := range globs {
		contextGlobs = append(contextGlobs, filepath.Join(contextDir, glob))
	}
	logboek.Context(ctx).Debug().LogF("Calculating checksum for globs %v in context dir %q: will scan following dirs globs: %v\n", globs, contextDir, contextGlobs)

	globStats, err := copier.Stat(contextDir, contextDir, copier.StatOptions{CheckForArchives: opts.CheckForArchives}, contextGlobs)
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
			if dockerfileStageDependenciesDebug() {
				logboek.Context(ctx).Debug().LogF("Calculating checksum for globs %v in context dir %q: matched path %q\n", globs, contextDir, relMatch)
			}
			matches = append(matches, relMatch)
		}
	}

	sort.Strings(matches)
	matches = util.UniqStrings(matches)

	pathsChecksum, err := a.CalculatePathsChecksum(ctx, matches)
	if err != nil {
		return "", fmt.Errorf("unable to calculate build context paths checksum: %w", err)
	}

	if !opts.IncludeMatchedPaths {
		return pathsChecksum, nil
	}

	return util.Sha256Hash(append(matches, pathsChecksum)...), nil
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

func dockerfileStageDependenciesDebug() bool {
	return os.Getenv("WERF_DEBUG_DOCKERFILE_STAGE_DEPENDENCIES") == "1"
}
