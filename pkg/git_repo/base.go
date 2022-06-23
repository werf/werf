package git_repo

import (
	"context"
	"fmt"
	"io"
	"os"
	pathPkg "path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/types"
	"github.com/werf/werf/pkg/git_repo/repo_handle"
	"github.com/werf/werf/pkg/path_matcher"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/true_git/ls_tree"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
)

type Base struct {
	Name string

	Cache Cache

	commitRepoHandle      sync.Map
	commitRepoHandleMutex sync.Map

	initRepoHandleBackedByWorkTreeFunc func(context.Context, string) (repo_handle.Handle, error)
}

func NewBase(name string, initRepoHandleBackedByWorkTreeFunc func(context.Context, string) (repo_handle.Handle, error)) *Base {
	base := &Base{
		Name: name,
		Cache: Cache{
			Archives: make(map[string]Archive),
			Patches:  make(map[string]Patch),
		},
	}
	base.initRepoHandleBackedByWorkTreeFunc = initRepoHandleBackedByWorkTreeFunc
	return base
}

type Cache struct {
	Patches   map[string]Patch
	Archives  map[string]Archive
	Checksums sync.Map

	patchesMutex   sync.Mutex
	archivesMutex  sync.Mutex
	checksumsMutex sync.Map
}

func (repo *Base) initRepoHandleBackedByWorkTree(ctx context.Context, commit string) (repo_handle.Handle, error) {
	if repo.initRepoHandleBackedByWorkTreeFunc == nil {
		panic("initRepoHandleBackedByWorkTreeFunc is nil")
	}

	return repo.initRepoHandleBackedByWorkTreeFunc(ctx, commit)
}

func (repo *Base) HeadCommitHash(ctx context.Context) (string, error) {
	panic("not implemented")
}

func (repo *Base) HeadCommitTime(ctx context.Context) (*time.Time, error) {
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
		return "", fmt.Errorf("cannot open repo %q: %w", repoPath, err)
	}

	cfg, err := repository.Config()
	if err != nil {
		return "", fmt.Errorf("cannot access repo config: %w", err)
	}

	if originCfg, hasKey := cfg.Remotes["origin"]; hasKey {
		return originCfg.URLs[0], nil
	}

	return "", nil
}

func (repo *Base) isEmpty(ctx context.Context, repoPath string) (bool, error) {
	repository, err := git.PlainOpenWithOptions(repoPath, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return false, fmt.Errorf("cannot open repo %q: %w", repoPath, err)
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

func getHeadCommit(ctx context.Context, repoPath string) (string, error) {
	res, err := true_git.ShowRef(ctx, repoPath)
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

	patchID := true_git.PatchOptions(opts).ID()
	if _, hasKey := repo.Cache.Patches[patchID]; !hasKey {
		patch, err := repo.CreatePatch(ctx, repoPath, gitDir, repoID, workTreeCacheDir, opts)
		if err != nil {
			return nil, err
		}
		repo.Cache.Patches[patchID] = patch
	}
	return repo.Cache.Patches[patchID], nil
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
	if lock, err := CommonGitDataManager.LockGC(ctx, true); err != nil {
		return nil, err
	} else {
		defer werf.ReleaseHostLock(lock)
	}

	if patch, err := CommonGitDataManager.GetPatchFile(ctx, repoID, opts); err != nil {
		return nil, err
	} else if patch != nil {
		return patch, err
	}

	repository, err := git.PlainOpenWithOptions(repoPath, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return nil, fmt.Errorf("cannot open repo %q: %w", repoPath, err)
	}

	fromHash, err := newHash(opts.FromCommit)
	if err != nil {
		return nil, fmt.Errorf("bad `from` commit hash %q: %w", opts.FromCommit, err)
	}

	_, err = repository.CommitObject(fromHash)
	if err != nil {
		return nil, fmt.Errorf("bad `from` commit %q: %w", opts.FromCommit, err)
	}

	toHash, err := newHash(opts.ToCommit)
	if err != nil {
		return nil, fmt.Errorf("bad `to` commit hash %q: %w", opts.ToCommit, err)
	}

	toCommit, err := repository.CommitObject(toHash)
	if err != nil {
		return nil, fmt.Errorf("bad `to` commit %q: %w", opts.ToCommit, err)
	}

	hasSubmodules, err := HasSubmodulesInCommit(toCommit)
	if err != nil {
		return nil, err
	}

	tmpFile, err := CommonGitDataManager.NewTmpFile()
	if err != nil {
		return nil, err
	}

	fileHandler, err := os.OpenFile(tmpFile, os.O_RDWR|os.O_CREATE, 0o755)
	if err != nil {
		return nil, fmt.Errorf("cannot open file %s: %w", tmpFile, err)
	}

	var desc *true_git.PatchDescriptor
	if hasSubmodules {
		var retryCount int

	TryCreatePatch:
		desc, err = true_git.PatchWithSubmodules(ctx, fileHandler, gitDir, workTreeCacheDir, true_git.PatchOptions(opts))

		if true_git.IsCommitsNotPresentError(err) && retryCount == 0 {
			logboek.Context(ctx).Default().LogF("Detected not present commits when creating patch: %s\n", err)
			logboek.Context(ctx).Default().LogF("Will switch worktree to original commit %q and retry\n", opts.FromCommit)

			if err := fileHandler.Truncate(0); err != nil {
				return nil, fmt.Errorf("unable to truncate file %s: %w", tmpFile, err)
			}
			if _, err := fileHandler.Seek(0, 0); err != nil {
				return nil, fmt.Errorf("unable to reset file %s: %w", tmpFile, err)
			}

			if err := true_git.WithWorkTree(ctx, gitDir, workTreeCacheDir, opts.FromCommit, true_git.WithWorkTreeOptions{HasSubmodules: true}, func(workTreeDir string) error {
				return nil
			}); err != nil {
				return nil, fmt.Errorf("unable to switch worktree to commit %q: %w", opts.FromCommit, err)
			}

			retryCount++
			goto TryCreatePatch
		}
	} else {
		desc, err = true_git.Patch(ctx, fileHandler, gitDir, true_git.PatchOptions(opts))
	}

	if err != nil {
		return nil, fmt.Errorf("error creating patch between %q and %q commits: %w", opts.FromCommit, opts.ToCommit, err)
	}

	err = fileHandler.Close()
	if err != nil {
		return nil, fmt.Errorf("error creating patch file %s: %w", tmpFile, err)
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

func (repo *Base) createDetachedMergeCommit(ctx context.Context, gitDir, path, workTreeCacheDir, fromCommit, toCommit string) (string, error) {
	if lock, err := CommonGitDataManager.LockGC(ctx, true); err != nil {
		return "", err
	} else {
		defer werf.ReleaseHostLock(lock)
	}

	repository, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return "", fmt.Errorf("cannot open repo at %s: %w", path, err)
	}
	commitHash, err := newHash(toCommit)
	if err != nil {
		return "", fmt.Errorf("bad commit hash %s: %w", toCommit, err)
	}
	v1MergeIntoCommitObj, err := repository.CommitObject(commitHash)
	if err != nil {
		return "", fmt.Errorf("bad commit %s: %w", toCommit, err)
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
		return nil, fmt.Errorf("cannot open repo at %s: %w", gitDir, err)
	}
	commitHash, err := newHash(commit)
	if err != nil {
		return nil, fmt.Errorf("bad commit hash %s: %w", commit, err)
	}
	commitObj, err := repository.CommitObject(commitHash)
	if err != nil {
		return nil, fmt.Errorf("bad commit %s: %w", commit, err)
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

	archiveID := true_git.ArchiveOptions(opts).ID()
	if _, hasKey := repo.Cache.Archives[archiveID]; !hasKey {
		archive, err := repo.CreateArchive(ctx, repoPath, gitDir, repoID, workTreeCacheDir, opts)
		if err != nil {
			return nil, err
		}
		repo.Cache.Archives[archiveID] = archive
	}
	return repo.Cache.Archives[archiveID], nil
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
	if lock, err := CommonGitDataManager.LockGC(ctx, true); err != nil {
		return nil, err
	} else {
		defer werf.ReleaseHostLock(lock)
	}

	if archive, err := CommonGitDataManager.GetArchiveFile(ctx, repoID, opts); err != nil {
		return nil, err
	} else if archive != nil {
		return archive, nil
	}

	repository, err := git.PlainOpenWithOptions(repoPath, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return nil, fmt.Errorf("cannot open repo %q: %w", repoPath, err)
	}

	commitHash, err := newHash(opts.Commit)
	if err != nil {
		return nil, fmt.Errorf("bad commit hash %q: %w", opts.Commit, err)
	}

	commit, err := repository.CommitObject(commitHash)
	if err != nil {
		return nil, fmt.Errorf("bad commit %q: %w", opts.Commit, err)
	}

	hasSubmodules, err := HasSubmodulesInCommit(commit)
	if err != nil {
		return nil, err
	}

	tmpPath, err := CommonGitDataManager.NewTmpFile()
	if err != nil {
		return nil, err
	}

	fileHandler, err := os.OpenFile(tmpPath, os.O_RDWR|os.O_CREATE, 0o755)
	if err != nil {
		return nil, fmt.Errorf("cannot open archive file: %w", err)
	}
	defer fileHandler.Close()

	if hasSubmodules {
		err = true_git.ArchiveWithSubmodules(ctx, fileHandler, gitDir, workTreeCacheDir, true_git.ArchiveOptions(opts))
	} else {
		err = true_git.Archive(ctx, fileHandler, gitDir, workTreeCacheDir, true_git.ArchiveOptions(opts))
	}

	if err != nil {
		return nil, fmt.Errorf("error creating archive for commit %q: %w", opts.Commit, err)
	}

	if err := fileHandler.Close(); err != nil {
		return nil, fmt.Errorf("unable to close file %s: %w", tmpPath, err)
	}

	if archive, err := CommonGitDataManager.CreateArchiveFile(ctx, repoID, opts, tmpPath); err != nil {
		return nil, err
	} else {
		return archive, nil
	}
}

func (repo *Base) isCommitExists(ctx context.Context, repoPath, gitDir, commit string) (bool, error) {
	repository, err := git.PlainOpenWithOptions(repoPath, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return false, fmt.Errorf("cannot open repo %q: %w", repoPath, err)
	}

	commitHash, err := newHash(commit)
	if err != nil {
		return false, fmt.Errorf("bad commit hash %q: %w", commit, err)
	}

	_, err = repository.CommitObject(commitHash)
	if err == plumbing.ErrObjectNotFound {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("bad commit %q: %w", commit, err)
	}

	return true, nil
}

func (repo *Base) tagsList(repoPath string) ([]string, error) {
	repository, err := git.PlainOpenWithOptions(repoPath, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return nil, fmt.Errorf("cannot open repo %q: %w", repoPath, err)
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
		return nil, fmt.Errorf("cannot open repo %q: %w", repoPath, err)
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

func (repo *Base) getOrCreateChecksum(ctx context.Context, repoHandle repo_handle.Handle, opts ChecksumOptions) (string, error) {
	checksumID := opts.ID()
	checksumMutex := util.MapLoadOrCreateMutex(&repo.Cache.checksumsMutex, checksumID)
	checksumMutex.Lock()
	defer checksumMutex.Unlock()

	if _, hasKey := repo.Cache.Checksums.Load(checksumID); !hasKey {
		checksum, err := repo.CreateChecksum(ctx, repoHandle, opts)
		if err != nil {
			return "", err
		}
		repo.Cache.Checksums.Store(checksumID, checksum)
	}

	checksum := util.MapMustLoad(&repo.Cache.Checksums, checksumID).(string)
	return checksum, nil
}

func (repo *Base) CreateChecksum(ctx context.Context, repoHandle repo_handle.Handle, opts ChecksumOptions) (checksum string, err error) {
	logboek.Context(ctx).Debug().LogProcess("Creating checksum").Do(func() {
		logboek.Context(ctx).Debug().LogFDetails("repository: %s\noptions: %+v\n", repo.Name, opts)
		logboek.Context(ctx).Debug().LogOptionalLn()
		checksum, err = repo.createChecksum(ctx, repoHandle, opts)
	})

	return
}

func (repo *Base) createChecksum(ctx context.Context, repoHandle repo_handle.Handle, opts ChecksumOptions) (checksum string, err error) {
	lsTreeResult, err := repo.lsTreeResultWithExistingHandle(ctx, repoHandle, opts.Commit, opts.LsTreeOptions)
	if err != nil {
		return "", err
	}

	return lsTreeResult.Checksum(ctx), nil
}

func (repo *Base) lsTreeResult(ctx context.Context, commit string, opts LsTreeOptions) (result *ls_tree.Result, err error) {
	err = repo.withRepoHandle(ctx, commit, func(repoHandle repo_handle.Handle) error {
		result, err = repo.lsTreeResultWithExistingHandle(ctx, repoHandle, commit, opts)
		return err
	})

	return
}

func (repo *Base) lsTreeResultWithExistingHandle(ctx context.Context, repoHandle repo_handle.Handle, commit string, opts LsTreeOptions) (result *ls_tree.Result, err error) {
	return ls_tree.LsTree(ctx, repoHandle, commit, ls_tree.LsTreeOptions(opts))
}

func (repo *Base) withRepoHandle(ctx context.Context, commit string, f func(handle repo_handle.Handle) error) error {
	mutex := util.MapLoadOrCreateMutex(&repo.commitRepoHandleMutex, commit)
	mutex.Lock()
	defer mutex.Unlock()

	attempt := 0
	retriesLimit := 1

initCommitRepoHandle:
	if _, hasKey := repo.commitRepoHandle.Load(commit); !hasKey {
		repoHandler, err := repo.initRepoHandleBackedByWorkTree(ctx, commit)
		if err != nil {
			return err
		}

		repo.commitRepoHandle.Store(commit, repoHandler)
	}

	repoHandler := util.MapMustLoad(&repo.commitRepoHandle, commit).(repo_handle.Handle)
	if err := f(repoHandler); err != nil {
		isHandledError := strings.HasSuffix(err.Error(), "object not found") || strings.HasSuffix(err.Error(), "packfile not found")
		if isHandledError && attempt < retriesLimit {
			attempt++

			logboek.Context(ctx).Warn().LogF("WARNING: Something went wrong: %s\n", err)
			logboek.Context(ctx).Warn().LogLn("WARNING: Retrying go-git operation ...")

			// reinit the commit repo handle
			repo.commitRepoHandle.Delete(commit)
			goto initCommitRepoHandle
		}

		return err
	}

	return nil
}

func (repo *Base) GetCommitTreeEntry(ctx context.Context, commit, path string) (*ls_tree.LsTreeEntry, error) {
	lsTreeResult, err := repo.lsTreeResult(ctx, commit, LsTreeOptions{
		PathScope: path,
		AllFiles:  false,
	})
	if err != nil {
		return nil, err
	}

	entry := lsTreeResult.LsTreeEntry(path)

	return entry, nil
}

func (repo *Base) IsCommitTreeEntryExist(ctx context.Context, commit, relPath string) (exist bool, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("IsCommitTreeEntryExist %q %q", commit, relPath).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debugGiterminismManager() {
				options.Mute()
			}
		}).
		Do(func() {
			exist, err = repo.isTreeEntryExist(ctx, commit, relPath)

			if debugGiterminismManager() {
				logboek.Context(ctx).Debug().LogF("exist: %v\nerr: %q\n", exist, err)
			}
		})

	return
}

func (repo *Base) isTreeEntryExist(ctx context.Context, commit, relPath string) (bool, error) {
	entry, err := repo.GetCommitTreeEntry(ctx, commit, relPath)
	if err != nil {
		return false, err
	}

	return !entry.Mode.IsMalformed(), nil
}

func (repo *Base) IsCommitTreeEntryDirectory(ctx context.Context, commit, relPath string) (isDirectory bool, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("IsCommitTreeEntryDirectory %q %q", commit, relPath).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debugGiterminismManager() {
				options.Mute()
			}
		}).
		Do(func() {
			isDirectory, err = repo.isCommitTreeEntryDirectory(ctx, commit, relPath)

			if debugGiterminismManager() {
				logboek.Context(ctx).Debug().LogF("isDirectory: %v\nerr: %q\n", isDirectory, err)
			}
		})

	return
}

func (repo *Base) isCommitTreeEntryDirectory(ctx context.Context, commit, relPath string) (bool, error) {
	entry, err := repo.GetCommitTreeEntry(ctx, commit, relPath)
	if err != nil {
		return false, err
	}

	return entry.Mode == filemode.Dir || entry.Mode == filemode.Submodule, nil
}

func (repo *Base) ReadCommitTreeEntryContent(ctx context.Context, commit, relPath string) ([]byte, error) {
	lsTreeResult, err := repo.lsTreeResult(ctx, commit, LsTreeOptions{
		PathScope: relPath,
		AllFiles:  false,
	})
	if err != nil {
		return nil, err
	}

	var content []byte
	err = repo.withRepoHandle(ctx, commit, func(repoHandle repo_handle.Handle) error {
		content, err = lsTreeResult.LsTreeEntryContent(repoHandle, relPath)
		return err
	})

	return content, err
}

// ResolveCommitFilePath follows symbolic links and returns the resolved path if there is a corresponding tree entry in the repo.
func (repo *Base) ResolveCommitFilePath(ctx context.Context, commit, path string) (resolvedPath string, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("ResolveCommitFilePath %q %q", commit, path).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debugGiterminismManager() {
				options.Mute()
			}
		}).
		Do(func() {
			resolvedPath, err = repo.resolveCommitFilePath(ctx, commit, path, 0, nil)

			if debugGiterminismManager() {
				logboek.Context(ctx).Debug().LogF("resolvedPath: %v\nerr: %q\n", resolvedPath, err)
			}
		})

	return
}

func (repo *Base) resolveCommitFilePath(ctx context.Context, commit, path string, depth int, checkSymlinkTargetFunc func(resolvedPath string) error) (string, error) {
	if depth > 1000 {
		return "", fmt.Errorf("too many levels of symbolic links")
	}
	depth++

	// returns path if the corresponding tree entry is Regular, Deprecated, Executable, Dir, or Submodule.
	{
		lsTreeEntry, err := repo.GetCommitTreeEntry(ctx, commit, path)
		if err != nil {
			return "", fmt.Errorf("unable to get commit tree entry %q: %w", path, err)
		}

		if debugGiterminismManager() {
			logboek.Context(ctx).Debug().LogF("-- [*] %s %s %s\n", path, lsTreeEntry.Mode.String(), err)
		}

		if lsTreeEntry.Mode != filemode.Symlink && !lsTreeEntry.Mode.IsMalformed() {
			return path, nil
		}
	}

	pathParts := util.SplitFilepath(path)
	pathPartsLen := len(pathParts)

	var resolvedPath string
	for ind := 0; ind < pathPartsLen; ind++ {
		pathToResolve := pathPkg.Join(resolvedPath, pathParts[ind])

		lsTreeEntry, err := repo.GetCommitTreeEntry(ctx, commit, pathToResolve)
		if err != nil {
			return "", fmt.Errorf("unable to get commit tree entry %q: %w", pathToResolve, err)
		}

		if debugGiterminismManager() {
			logboek.Context(ctx).Debug().LogF("-- [%d] %s %s %s\n", ind, pathToResolve, lsTreeEntry.Mode.String(), err)
		}

		mode := lsTreeEntry.Mode
		switch {
		case mode.IsMalformed():
			return "", treeEntryNotFoundInRepoErr{fmt.Errorf("commit tree entry %q not found in the repository", pathToResolve)}
		case mode == filemode.Symlink:
			data, err := repo.ReadCommitTreeEntryContent(ctx, commit, pathToResolve)
			if err != nil {
				return "", fmt.Errorf("unable to get commit tree entry content %q: %w", pathToResolve, err)
			}

			link := string(data)
			if pathPkg.IsAbs(link) {
				return "", treeEntryNotFoundInRepoErr{fmt.Errorf("commit tree entry %q not found in the repository", link)}
			}

			resolvedLink := pathPkg.Join(pathPkg.Dir(pathToResolve), link)
			if resolvedLink == ".." || strings.HasPrefix(resolvedLink, "../") {
				return "", treeEntryNotFoundInRepoErr{fmt.Errorf("commit tree entry %q not found in the repository", link)}
			}

			if checkSymlinkTargetFunc != nil {
				if err := checkSymlinkTargetFunc(resolvedLink); err != nil {
					return "", err
				}
			}

			resolvedTarget, err := repo.resolveCommitFilePath(ctx, commit, resolvedLink, depth, checkSymlinkTargetFunc)
			if err != nil {
				return "", err
			}

			resolvedPath = resolvedTarget
		default:
			resolvedPath = pathToResolve
		}
	}

	return resolvedPath, nil
}

// ResolveAndCheckCommitFilePath does ResolveCommitFilePath with an additional check for each resolved link target.
func (repo *Base) ResolveAndCheckCommitFilePath(ctx context.Context, commit, path string, checkSymlinkTargetFunc func(resolvedPath string) error) (resolvedPath string, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("ResolveAndCheckCommitFilePath %q %q", commit, path).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debugGiterminismManager() {
				options.Mute()
			}
		}).
		Do(func() {
			checkWithDebugFunc := func(resolvedPath string) error {
				return logboek.Context(ctx).Debug().
					LogBlock("-- check %s", resolvedPath).
					Options(func(options types.LogBlockOptionsInterface) {
						if !debugGiterminismManager() {
							options.Mute()
						}
					}).
					DoError(func() error {
						err := checkSymlinkTargetFunc(resolvedPath)

						if debugGiterminismManager() {
							logboek.Context(ctx).Debug().LogF("err: %q\n", err)
						}

						return err
					})
			}

			resolvedPath, err = repo.resolveAndCheckCommitFilePath(ctx, commit, path, checkWithDebugFunc)

			if debugGiterminismManager() {
				logboek.Context(ctx).Debug().LogF("resolvedPath: %v\nerr: %q\n", resolvedPath, err)
			}
		})

	return
}

func (repo *Base) resolveAndCheckCommitFilePath(ctx context.Context, commit, path string, checkSymlinkTargetFunc func(relPath string) error) (resolvedPath string, err error) {
	return repo.resolveCommitFilePath(ctx, commit, path, 0, checkSymlinkTargetFunc)
}

func (repo *Base) checkCommitFileMode(ctx context.Context, commit, path string, expectedFileModeList []filemode.FileMode) (bool, error) {
	resolvedPath, err := repo.ResolveCommitFilePath(ctx, commit, path)
	if err != nil {
		if IsTreeEntryNotFoundInRepoErr(err) {
			return false, nil
		}

		return false, fmt.Errorf("unable to resolve commit file %q: %w", path, err)
	}

	lsTreeEntry, err := repo.GetCommitTreeEntry(ctx, commit, resolvedPath)
	if err != nil {
		return false, fmt.Errorf("unable to get commit tree entry %q: %w", resolvedPath, err)
	}

	for _, mode := range expectedFileModeList {
		if mode == lsTreeEntry.Mode {
			return true, nil
		}
	}

	return false, nil
}

// IsCommitDirectoryExist resolves symlinks and returns true if the resolved commit tree entry is Dir or Submodule.
func (repo *Base) IsCommitDirectoryExist(ctx context.Context, commit, path string) (exist bool, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("IsCommitDirectoryExist %q %q", commit, path).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debugGiterminismManager() {
				options.Mute()
			}
		}).
		Do(func() {
			exist, err = repo.isCommitDirectoryExist(ctx, commit, path)

			if debugGiterminismManager() {
				logboek.Context(ctx).Debug().LogF("exist: %v\nerr: %q\n", exist, err)
			}
		})

	return
}

func (repo *Base) isCommitDirectoryExist(ctx context.Context, commit, path string) (bool, error) {
	return repo.checkCommitFileMode(ctx, commit, path, []filemode.FileMode{filemode.Dir, filemode.Submodule})
}

// IsCommitFileExist resolves symlinks and returns true if the resolved commit tree entry is Regular, Deprecated, or Executable.
func (repo *Base) IsCommitFileExist(ctx context.Context, commit, path string) (exist bool, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("IsCommitFileExist %q %q", commit, path).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debugGiterminismManager() {
				options.Mute()
			}
		}).
		Do(func() {
			exist, err = repo.isCommitFileExist(ctx, commit, path)

			if debugGiterminismManager() {
				logboek.Context(ctx).Debug().LogF("exist: %v\nerr: %q\n", exist, err)
			}
		})

	return
}

func (repo *Base) isCommitFileExist(ctx context.Context, commit, path string) (bool, error) {
	return repo.checkCommitFileMode(ctx, commit, path, []filemode.FileMode{filemode.Regular, filemode.Deprecated, filemode.Executable})
}

// ReadCommitFile resolves symlinks and returns commit tree entry content.
func (repo *Base) ReadCommitFile(ctx context.Context, commit, path string) (data []byte, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("ReadCommitFile %q %q", commit, path).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debugGiterminismManager() {
				options.Mute()
			}
		}).
		Do(func() {
			data, err = repo.readCommitFile(ctx, commit, path)

			if debugGiterminismManager() {
				logboek.Context(ctx).Debug().LogF("dataLength: %v\nerr: %q\n", len(data), err)
			}
		})

	return
}

func (repo *Base) readCommitFile(ctx context.Context, commit, path string) ([]byte, error) {
	resolvedPath, err := repo.ResolveCommitFilePath(ctx, commit, path)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve commit file %q: %w", path, err)
	}

	return repo.ReadCommitTreeEntryContent(ctx, commit, resolvedPath)
}

func (repo *Base) IsAnyCommitTreeEntriesMatched(ctx context.Context, commit, pathScope string, pathMatcher path_matcher.PathMatcher, allFiles bool) (bool, error) {
	result, err := repo.lsTreeResult(ctx, commit, LsTreeOptions{
		PathScope:   pathScope,
		PathMatcher: pathMatcher,
		AllFiles:    allFiles,
	})
	if err != nil {
		return false, err
	}

	return !result.IsEmpty(), nil
}

func (repo *Base) WalkCommitFiles(ctx context.Context, commit, dir string, pathMatcher path_matcher.PathMatcher, fileFunc func(notResolvedPath string) error) error {
	if !pathMatcher.IsDirOrSubmodulePathMatched(dir) {
		return nil
	}

	exist, err := repo.IsCommitDirectoryExist(ctx, commit, dir)
	if err != nil {
		return err
	}

	if !exist {
		return nil
	}

	resolvedDir, err := repo.ResolveCommitFilePath(ctx, commit, dir)
	if err != nil {
		return fmt.Errorf("unable to resolve commit file %q: %w", dir, err)
	}

	result, err := repo.lsTreeResult(ctx, commit, LsTreeOptions{
		PathScope: resolvedDir,
		AllFiles:  true,
	})
	if err != nil {
		return err
	}

	return result.Walk(func(lsTreeEntry *ls_tree.LsTreeEntry) error {
		notResolvedPath := strings.Replace(filepath.ToSlash(lsTreeEntry.FullFilepath), resolvedDir, dir, 1)

		if debugGiterminismManager() {
			logboek.Context(ctx).Debug().LogF("-- %q %q\n", notResolvedPath, lsTreeEntry.Mode.String())
		}

		if lsTreeEntry.Mode.IsMalformed() {
			panic(fmt.Sprintf("unexpected condition: %+v", lsTreeEntry))
		}

		if !pathMatcher.IsDirOrSubmodulePathMatched(notResolvedPath) {
			return nil
		}

		if lsTreeEntry.Mode == filemode.Symlink {
			isDir, err := repo.IsCommitDirectoryExist(ctx, commit, notResolvedPath)
			if err != nil {
				return err
			}

			if isDir {
				err := repo.WalkCommitFiles(ctx, commit, notResolvedPath, pathMatcher, fileFunc)
				if err != nil {
					return err
				}
			} else {
				if err := fileFunc(notResolvedPath); err != nil {
					return err
				}
			}
		} else {
			if err := fileFunc(notResolvedPath); err != nil {
				return err
			}
		}

		return nil
	})
}

// ListCommitFilesWithGlob returns the list of files by the glob, follows symlinks.
// The result paths are relative to the passed directory, the method does reverse resolving for symlinks.
func (repo *Base) ListCommitFilesWithGlob(ctx context.Context, commit, dir, glob string) (files []string, err error) {
	var prefixWithoutPatterns string
	prefixWithoutPatterns, glob = util.GlobPrefixWithoutPatterns(glob)
	dirOrFileWithGlobPrefix := filepath.Join(dir, prefixWithoutPatterns)

	pathMatcher := path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{
		BasePath:     dirOrFileWithGlobPrefix,
		IncludeGlobs: []string{glob},
	})
	if debugGiterminismManager() {
		logboek.Context(ctx).Debug().LogLn("pathMatcher:", pathMatcher.String())
	}

	var result []string
	fileFunc := func(notResolvedPath string) error {
		if pathMatcher.IsPathMatched(notResolvedPath) {
			result = append(result, notResolvedPath)
		}

		return nil
	}

	isRegularFile, err := repo.IsCommitFileExist(ctx, commit, dirOrFileWithGlobPrefix)
	if err != nil {
		return nil, err
	}

	if isRegularFile {
		if err := fileFunc(dirOrFileWithGlobPrefix); err != nil {
			return nil, err
		}

		return result, nil
	}

	if err := repo.WalkCommitFiles(ctx, commit, dirOrFileWithGlobPrefix, pathMatcher, fileFunc); err != nil {
		return nil, err
	}

	return result, nil
}

func baseHeadCommitTime(repo gitRepo, ctx context.Context) (*time.Time, error) {
	headCommitHash, err := repo.HeadCommitHash(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get HEAD hash: %w", err)
	}

	var time *time.Time
	if err := repo.withRepoHandle(ctx, headCommitHash, func(repoHandle repo_handle.Handle) error {
		headHash, err := newHash(headCommitHash)
		if err != nil {
			return fmt.Errorf("unable to create new Hash object from commit SHA %q: %w", headCommitHash, err)
		}

		repo := repoHandle.Repository()
		if repo == nil {
			return fmt.Errorf("unable to get repository from repoHandle")
		}

		commit, err := repo.CommitObject(headHash)
		if err != nil {
			return fmt.Errorf("unable to get commit object for ref %q: %w", headCommitHash, err)
		}

		time = &commit.Author.When

		return nil
	}); err != nil {
		return nil, err
	}

	return time, nil
}
