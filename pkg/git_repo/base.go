package git_repo

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/path_matcher"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/true_git/ls_tree"
	"github.com/werf/werf/pkg/util"
)

type Base struct {
	Name string

	Cache Cache
}

func NewBase(name string) Base {
	return Base{
		Name: name,
		Cache: Cache{
			Archives:       make(map[string]Archive),
			Patches:        make(map[string]Patch),
			Checksums:      make(map[string]Checksum),
			checksumsMutex: make(map[string]sync.Mutex),
		},
	}
}

type Cache struct {
	Patches   map[string]Patch
	Checksums map[string]Checksum
	Archives  map[string]Archive

	patchesMutex   sync.Mutex
	checksumsMutex map[string]sync.Mutex
	archivesMutex  sync.Mutex
}

func (repo *Base) HeadCommit(ctx context.Context) (string, error) {
	panic("not implemented")
}

func (repo *Base) LatestBranchCommit(ctx context.Context, branch string) (string, error) {
	panic("not implemented")
}

func (repo *Base) TagCommit(ctx context.Context, branch string) (string, error) {
	panic("not implemented")
}

func (repo *Base) remoteOriginUrl(repoPath string) (string, error) {
	repository, err := git.PlainOpenWithOptions(repoPath, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return "", fmt.Errorf("cannot open repo %q: %s", repoPath, err)
	}

	cfg, err := repository.Config()
	if err != nil {
		return "", fmt.Errorf("cannot access repo config: %s", err)
	}

	if originCfg, hasKey := cfg.Remotes["origin"]; hasKey {
		return originCfg.URLs[0], nil
	}

	return "", nil
}

func (repo *Base) isEmpty(ctx context.Context, repoPath string) (bool, error) {
	repository, err := git.PlainOpenWithOptions(repoPath, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return false, fmt.Errorf("cannot open repo %q: %s", repoPath, err)
	}

	commitIter, err := repository.CommitObjects()
	if err != nil {
		return false, err
	}

	_, err = commitIter.Next()
	if err == io.EOF {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return false, nil
}

func getHeadCommit(repoPath string) (string, error) {
	res, err := true_git.ShowRef(repoPath)
	if err != nil {
		return "", err
	}

	for _, ref := range res.Refs {
		if ref.IsHEAD {
			return ref.Commit, nil
		}
	}

	return "", err
}

func (repo *Base) String() string {
	return repo.GetName()
}

func (repo *Base) GetName() string {
	return repo.Name
}

func (repo *Base) getOrCreatePatch(ctx context.Context, repoPath, gitDir, repoID, workTreeCacheDir string, opts PatchOptions) (Patch, error) {
	repo.Cache.patchesMutex.Lock()
	defer repo.Cache.patchesMutex.Unlock()

	if _, hasKey := repo.Cache.Patches[util.ObjectToHashKey(opts)]; !hasKey {
		patch, err := repo.CreatePatch(ctx, repoPath, gitDir, repoID, workTreeCacheDir, opts)
		if err != nil {
			return nil, err
		}
		repo.Cache.Patches[util.ObjectToHashKey(opts)] = patch
	}
	return repo.Cache.Patches[util.ObjectToHashKey(opts)], nil
}

func (repo *Base) CreatePatch(ctx context.Context, repoPath, gitDir, repoID, workTreeCacheDir string, opts PatchOptions) (patch Patch, err error) {
	logboek.Context(ctx).Debug().LogProcess("Creating patch").Do(func() {
		logboek.Context(ctx).Debug().LogFDetails("repository: %s\noptions: %+v\n", repo.Name, opts)
		logboek.Context(ctx).Debug().LogOptionalLn()
		patch, err = repo.createPatch(ctx, repoPath, gitDir, repoID, workTreeCacheDir, opts)
	})

	return
}

func (repo *Base) createPatch(ctx context.Context, repoPath, gitDir, repoID, workTreeCacheDir string, opts PatchOptions) (Patch, error) {
	if patch, err := CommonGitDataManager.GetPatchFile(ctx, repoID, opts); err != nil {
		return nil, err
	} else if patch != nil {
		return patch, err
	}

	repository, err := git.PlainOpenWithOptions(repoPath, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return nil, fmt.Errorf("cannot open repo %q: %s", repoPath, err)
	}

	fromHash, err := newHash(opts.FromCommit)
	if err != nil {
		return nil, fmt.Errorf("bad `from` commit hash %q: %s", opts.FromCommit, err)
	}

	_, err = repository.CommitObject(fromHash)
	if err != nil {
		return nil, fmt.Errorf("bad `from` commit %q: %s", opts.FromCommit, err)
	}

	toHash, err := newHash(opts.ToCommit)
	if err != nil {
		return nil, fmt.Errorf("bad `to` commit hash %q: %s", opts.ToCommit, err)
	}

	toCommit, err := repository.CommitObject(toHash)
	if err != nil {
		return nil, fmt.Errorf("bad `to` commit %q: %s", opts.ToCommit, err)
	}

	hasSubmodules, err := HasSubmodulesInCommit(toCommit)
	if err != nil {
		return nil, err
	}

	tmpFile, err := CommonGitDataManager.NewTmpFile()
	if err != nil {
		return nil, err
	}

	fileHandler, err := os.OpenFile(tmpFile, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, fmt.Errorf("cannot open file %s: %s", tmpFile, err)
	}

	patchOpts := true_git.PatchOptions{
		FromCommit: opts.FromCommit,
		ToCommit:   opts.ToCommit,
		PathMatcher: path_matcher.NewGitMappingPathMatcher(
			opts.BasePath,
			opts.IncludePaths,
			opts.ExcludePaths,
			false,
		),
		WithEntireFileContext: opts.WithEntireFileContext,
		WithBinary:            opts.WithBinary,
	}

	var desc *true_git.PatchDescriptor
	if hasSubmodules {
		desc, err = true_git.PatchWithSubmodules(ctx, fileHandler, gitDir, workTreeCacheDir, patchOpts)
	} else {
		desc, err = true_git.Patch(ctx, fileHandler, gitDir, patchOpts)
	}

	if err != nil {
		return nil, fmt.Errorf("error creating patch between %q and %q commits: %s", opts.FromCommit, opts.ToCommit, err)
	}

	err = fileHandler.Close()
	if err != nil {
		return nil, fmt.Errorf("error creating patch file %s: %s", tmpFile, err)
	}

	if patch, err := CommonGitDataManager.CreatePatchFile(ctx, repoID, opts, tmpFile, desc); err != nil {
		return nil, err
	} else {
		return patch, nil
	}
}

func HasSubmodulesInCommit(commit *object.Commit) (bool, error) {
	_, err := commit.File(".gitmodules")
	if err == object.ErrFileNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (repo *Base) createDetachedMergeCommit(ctx context.Context, gitDir, path, workTreeCacheDir string, fromCommit, toCommit string) (string, error) {
	repository, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return "", fmt.Errorf("cannot open repo at %s: %s", path, err)
	}
	commitHash, err := newHash(toCommit)
	if err != nil {
		return "", fmt.Errorf("bad commit hash %s: %s", toCommit, err)
	}
	v1MergeIntoCommitObj, err := repository.CommitObject(commitHash)
	if err != nil {
		return "", fmt.Errorf("bad commit %s: %s", toCommit, err)
	}
	hasSubmodules, err := HasSubmodulesInCommit(v1MergeIntoCommitObj)
	if err != nil {
		return "", err
	}

	return true_git.CreateDetachedMergeCommit(ctx, gitDir, workTreeCacheDir, fromCommit, toCommit, true_git.CreateDetachedMergeCommitOptions{HasSubmodules: hasSubmodules})
}

func (repo *Base) getMergeCommitParents(gitDir, commit string) ([]string, error) {
	repository, err := git.PlainOpenWithOptions(gitDir, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return nil, fmt.Errorf("cannot open repo at %s: %s", gitDir, err)
	}
	commitHash, err := newHash(commit)
	if err != nil {
		return nil, fmt.Errorf("bad commit hash %s: %s", commit, err)
	}
	commitObj, err := repository.CommitObject(commitHash)
	if err != nil {
		return nil, fmt.Errorf("bad commit %s: %s", commit, err)
	}

	var res []string

	for _, parent := range commitObj.ParentHashes {
		res = append(res, parent.String())
	}

	return res, nil
}

func (repo *Base) getOrCreateArchive(ctx context.Context, repoPath, gitDir, repoID, workTreeCacheDir string, opts ArchiveOptions) (Archive, error) {
	repo.Cache.archivesMutex.Lock()
	defer repo.Cache.archivesMutex.Unlock()

	if _, hasKey := repo.Cache.Archives[util.ObjectToHashKey(opts)]; !hasKey {
		archive, err := repo.CreateArchive(ctx, repoPath, gitDir, repoID, workTreeCacheDir, opts)
		if err != nil {
			return nil, err
		}
		repo.Cache.Archives[util.ObjectToHashKey(opts)] = archive
	}
	return repo.Cache.Archives[util.ObjectToHashKey(opts)], nil
}

func (repo *Base) CreateArchive(ctx context.Context, repoPath, gitDir, repoID, workTreeCacheDir string, opts ArchiveOptions) (archive Archive, err error) {
	logboek.Context(ctx).Debug().LogProcess("Creating archive").Do(func() {
		logboek.Context(ctx).Debug().LogFDetails("repository: %s\noptions: %+v\n", repo.Name, opts)
		logboek.Context(ctx).Debug().LogOptionalLn()
		archive, err = repo.createArchive(ctx, repoPath, gitDir, repoID, workTreeCacheDir, opts)
	})

	return
}

func (repo *Base) createArchive(ctx context.Context, repoPath, gitDir, repoID, workTreeCacheDir string, opts ArchiveOptions) (Archive, error) {
	if archive, err := CommonGitDataManager.GetArchiveFile(ctx, repoID, opts); err != nil {
		return nil, err
	} else if archive != nil {
		return archive, nil
	}

	repository, err := git.PlainOpenWithOptions(repoPath, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return nil, fmt.Errorf("cannot open repo %q: %s", repoPath, err)
	}

	commitHash, err := newHash(opts.Commit)
	if err != nil {
		return nil, fmt.Errorf("bad commit hash %q: %s", opts.Commit, err)
	}

	commit, err := repository.CommitObject(commitHash)
	if err != nil {
		return nil, fmt.Errorf("bad commit %q: %s", opts.Commit, err)
	}

	hasSubmodules, err := HasSubmodulesInCommit(commit)
	if err != nil {
		return nil, err
	}

	tmpPath, err := CommonGitDataManager.NewTmpFile()
	if err != nil {
		return nil, err
	}

	fileHandler, err := os.OpenFile(tmpPath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, fmt.Errorf("cannot open archive file: %s", err)
	}
	defer fileHandler.Close()

	archiveOpts := true_git.ArchiveOptions{
		Commit: opts.Commit,
		PathMatcher: path_matcher.NewGitMappingPathMatcher(
			opts.BasePath,
			opts.IncludePaths,
			opts.ExcludePaths,
			true,
		),
	}

	var desc *true_git.ArchiveDescriptor
	if hasSubmodules {
		desc, err = true_git.ArchiveWithSubmodules(ctx, fileHandler, gitDir, workTreeCacheDir, archiveOpts)
	} else {
		desc, err = true_git.Archive(ctx, fileHandler, gitDir, workTreeCacheDir, archiveOpts)
	}

	if err != nil {
		return nil, fmt.Errorf("error creating archive for commit %q: %s", opts.Commit, err)
	}

	if err := fileHandler.Close(); err != nil {
		return nil, fmt.Errorf("unable to close file %s: %s", tmpPath, err)
	}

	if archive, err := CommonGitDataManager.CreateArchiveFile(ctx, repoID, opts, tmpPath, desc); err != nil {
		return nil, err
	} else {
		return archive, nil
	}
}

func (repo *Base) isCommitExists(ctx context.Context, repoPath, gitDir string, commit string) (bool, error) {
	repository, err := git.PlainOpenWithOptions(repoPath, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return false, fmt.Errorf("cannot open repo %q: %s", repoPath, err)
	}

	commitHash, err := newHash(commit)
	if err != nil {
		return false, fmt.Errorf("bad commit hash %q: %s", commit, err)
	}

	_, err = repository.CommitObject(commitHash)
	if err == plumbing.ErrObjectNotFound {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("bad commit %q: %s", commit, err)
	}

	return true, nil
}

func (repo *Base) tagsList(repoPath string) ([]string, error) {
	repository, err := git.PlainOpenWithOptions(repoPath, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return nil, fmt.Errorf("cannot open repo %q: %s", repoPath, err)
	}

	tags, err := repository.Tags()
	if err != nil {
		return nil, err
	}

	res := make([]string, 0)

	if err := tags.ForEach(func(ref *plumbing.Reference) error {
		obj, err := repository.TagObject(ref.Hash())
		switch err {
		case nil:
			res = append(res, obj.Name)
		case plumbing.ErrObjectNotFound:
			res = append(res, strings.TrimPrefix(ref.Name().String(), "refs/tags/"))
		default:
			// Some other error
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return res, nil
}

func (repo *Base) remoteBranchesList(repoPath string) ([]string, error) {
	repository, err := git.PlainOpenWithOptions(repoPath, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return nil, fmt.Errorf("cannot open repo %q: %s", repoPath, err)
	}

	branches, err := repository.References()
	if err != nil {
		return nil, err
	}

	remoteBranchPrefix := "refs/remotes/origin/"

	res := make([]string, 0)
	err = branches.ForEach(func(r *plumbing.Reference) error {
		refName := r.Name().String()
		if strings.HasPrefix(refName, remoteBranchPrefix) {
			value := strings.TrimPrefix(refName, remoteBranchPrefix)
			if value != "HEAD" {
				res = append(res, value)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (repo *Base) getOrCreateChecksum(ctx context.Context, yieldRepositoryFunc func(ctx context.Context, commit string, doFunc func(*git.Repository) error) error, opts ChecksumOptions) (Checksum, error) {
	checksumMutex, ok := repo.Cache.checksumsMutex[util.ObjectToHashKey(opts)]
	if !ok {
		checksumMutex = sync.Mutex{}
	}

	checksumMutex.Lock()
	defer checksumMutex.Unlock()

	if _, hasKey := repo.Cache.Checksums[util.ObjectToHashKey(opts)]; !hasKey {
		checksum, err := repo.CreateChecksum(ctx, yieldRepositoryFunc, opts)
		if err != nil {
			return nil, err
		}
		repo.Cache.Checksums[util.ObjectToHashKey(opts)] = checksum
	}
	return repo.Cache.Checksums[util.ObjectToHashKey(opts)], nil
}

func (repo *Base) CreateChecksum(ctx context.Context, yieldRepositoryFunc func(ctx context.Context, commit string, doFunc func(*git.Repository) error) error, opts ChecksumOptions) (checksum Checksum, err error) {
	logboek.Context(ctx).Debug().LogProcess("Creating checksum").Do(func() {
		logboek.Context(ctx).Debug().LogFDetails("repository: %s\noptions: %+v\n", repo.Name, opts)
		logboek.Context(ctx).Debug().LogOptionalLn()
		checksum, err = repo.createChecksum(ctx, yieldRepositoryFunc, opts)
	})

	return
}

func (repo *Base) createChecksum(ctx context.Context, yieldRepositoryFunc func(ctx context.Context, commit string, doFunc func(*git.Repository) error) error, opts ChecksumOptions) (cs Checksum, err error) {
	_ = yieldRepositoryFunc(ctx, opts.Commit, func(repository *git.Repository) error {
		cs, err = repo.checksumWithLsTree(ctx, repository, opts)
		return nil
	})

	return
}

func (repo *Base) checksumWithLsTree(ctx context.Context, repository *git.Repository, opts ChecksumOptions) (Checksum, error) {
	checksum := &ChecksumDescriptor{
		NoMatchPaths: make([]string, 0),
		Hash:         sha256.New(),
	}

	pathMatcher := path_matcher.NewGitMappingPathMatcher(
		opts.BasePath,
		opts.IncludePaths,
		opts.ExcludePaths,
		false,
	)

	var mainLsTreeResult *ls_tree.Result
	if err := logboek.Context(ctx).Debug().LogProcess("ls-tree (%s)", pathMatcher.String()).DoError(func() error {
		var err error
		mainLsTreeResult, err = ls_tree.LsTree(ctx, repository, opts.Commit, pathMatcher, true)
		return err
	}); err != nil {
		return nil, err
	}

	for _, p := range opts.Paths {
		var pathLsTreeResult *ls_tree.Result
		pathMatcher := path_matcher.NewSimplePathMatcher(
			opts.BasePath,
			[]string{p},
			false,
		)

		logProcess := logboek.Context(ctx).Debug().LogProcess("ls-tree (%s)", pathMatcher.String())
		logProcess.Start()
		var err error
		pathLsTreeResult, err = mainLsTreeResult.LsTree(ctx, pathMatcher)
		if err != nil {
			logProcess.Fail()
			return nil, err
		}
		logProcess.End()

		var pathChecksum string
		if !pathLsTreeResult.IsEmpty() {
			logboek.Context(ctx).Debug().LogBlock("ls-tree result checksum (%s)", pathMatcher.String()).Do(func() {
				pathChecksum = pathLsTreeResult.Checksum(ctx)
				logboek.Context(ctx).Debug().LogLn()
				logboek.Context(ctx).Debug().LogLn(pathChecksum)
			})
		}

		if pathChecksum != "" {
			checksum.Hash.Write([]byte(pathChecksum))
		} else {
			checksum.NoMatchPaths = append(checksum.NoMatchPaths, p)
		}
	}

	return checksum, nil
}
