package git_repo

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/config"

	"github.com/werf/werf/pkg/util"

	"github.com/go-git/go-git/v5/plumbing/filemode"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/path_matcher"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/true_git/ls_tree"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Base struct {
	Name string
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
		return "", fmt.Errorf("cannot open repo `%s`: %s", repoPath, err)
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
		return false, fmt.Errorf("cannot open repo `%s`: %s", repoPath, err)
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

func (repo *Base) createPatch(ctx context.Context, repoPath, gitDir, repoID, workTreeCacheDir string, opts PatchOptions) (Patch, error) {
	if patch, err := CommonGitDataManager.GetPatchFile(ctx, repoID, opts); err != nil {
		return nil, err
	} else if patch != nil {
		return patch, err
	}

	repository, err := git.PlainOpenWithOptions(repoPath, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return nil, fmt.Errorf("cannot open repo `%s`: %s", repoPath, err)
	}

	fromHash, err := newHash(opts.FromCommit)
	if err != nil {
		return nil, fmt.Errorf("bad `from` commit hash `%s`: %s", opts.FromCommit, err)
	}

	_, err = repository.CommitObject(fromHash)
	if err != nil {
		return nil, fmt.Errorf("bad `from` commit `%s`: %s", opts.FromCommit, err)
	}

	toHash, err := newHash(opts.ToCommit)
	if err != nil {
		return nil, fmt.Errorf("bad `to` commit hash `%s`: %s", opts.ToCommit, err)
	}

	toCommit, err := repository.CommitObject(toHash)
	if err != nil {
		return nil, fmt.Errorf("bad `to` commit `%s`: %s", opts.ToCommit, err)
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
		return nil, fmt.Errorf("error creating patch between `%s` and `%s` commits: %s", opts.FromCommit, opts.ToCommit, err)
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

func (repo *Base) createArchive(ctx context.Context, repoPath, gitDir, repoID, workTreeCacheDir string, opts ArchiveOptions) (Archive, error) {
	if archive, err := CommonGitDataManager.GetArchiveFile(ctx, repoID, opts); err != nil {
		return nil, err
	} else if archive != nil {
		return archive, nil
	}

	repository, err := git.PlainOpenWithOptions(repoPath, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return nil, fmt.Errorf("cannot open repo `%s`: %s", repoPath, err)
	}

	commitHash, err := newHash(opts.Commit)
	if err != nil {
		return nil, fmt.Errorf("bad commit hash `%s`: %s", opts.Commit, err)
	}

	commit, err := repository.CommitObject(commitHash)
	if err != nil {
		return nil, fmt.Errorf("bad commit `%s`: %s", opts.Commit, err)
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
		return nil, fmt.Errorf("error creating archive for commit `%s`: %s", opts.Commit, err)
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
		return false, fmt.Errorf("cannot open repo `%s`: %s", repoPath, err)
	}

	commitHash, err := newHash(commit)
	if err != nil {
		return false, fmt.Errorf("bad commit hash `%s`: %s", commit, err)
	}

	_, err = repository.CommitObject(commitHash)
	if err == plumbing.ErrObjectNotFound {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("bad commit `%s`: %s", commit, err)
	}

	return true, nil
}

func (repo *Base) doCheckAndReadCommitSymlinkInRepository(ctx context.Context, repository *git.Repository, commit, path string) (bool, []byte, error) {
	var result *struct {
		IsSymlink bool
		LinkDest  []byte
		Err       error
	}

	if err := YieldEachSubmoduleRepository(ctx, repository, func(ctx context.Context, submoduleRepository *git.Repository, submoduleConfig *config.Submodule, submoduleStatus *git.SubmoduleStatus) error {
		if !util.IsSubpathOfBasePath(submoduleConfig.Path, path) {
			return nil
		}

		pathInsideSubmodule := util.GetRelativeToBaseFilepath(submoduleConfig.Path, path)

		result = &struct {
			IsSymlink bool
			LinkDest  []byte
			Err       error
		}{}

		isSymlink, linkDest, err := repo.doCheckAndReadCommitSymlinkInRepository(ctx, submoduleRepository, submoduleStatus.Current.String(), pathInsideSubmodule)
		if isSymlink {
			linkDest = []byte(filepath.Clean(filepath.Join(submoduleConfig.Path, string(linkDest))))
		}

		result.IsSymlink, result.LinkDest, result.Err = isSymlink, linkDest, err

		return stopYield
	}); err != nil {
		return false, nil, err
	} else if result != nil {
		return result.IsSymlink, result.LinkDest, result.Err
	}

	commitHash, err := newHash(commit)
	if err != nil {
		return false, nil, fmt.Errorf("bad commit hash %q: %s", commit, err)
	}

	commitObj, err := repository.CommitObject(commitHash)
	if err != nil {
		return false, nil, fmt.Errorf("cannot get commit %q object: %s", commit, err)
	}

	if f, err := commitObj.File(path); err == object.ErrFileNotFound {
		return false, nil, nil
	} else if err != nil {
		return false, nil, err
	} else if f.Mode == filemode.Symlink {
		if content, err := f.Contents(); err != nil {
			return false, nil, err
		} else {
			return true, []byte(content), nil
		}
	}

	return false, nil, nil
}

func (repo *Base) checkAndReadSymlink(ctx context.Context, repoPath, gitDir, commit, path string) (bool, []byte, error) {
	repository, err := git.PlainOpenWithOptions(repoPath, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return false, nil, fmt.Errorf("cannot open repo %s: %s", repoPath, err)
	}
	return repo.checkAndReadSymlinkInRepository(ctx, repository, commit, path)
}

var (
	stopYield = errors.New("stop yield")
)

func YieldEachSubmoduleRepository(ctx context.Context, repository *git.Repository, f func(ctx context.Context, submoduleRepository *git.Repository, submoduleConfig *config.Submodule, submoduleStatus *git.SubmoduleStatus) error) error {
	worktree, err := repository.Worktree()
	if err != nil {
		return fmt.Errorf("error getting worktree of repository: %s", err)
	}

	submodules, err := worktree.Submodules()
	if err != nil {
		return fmt.Errorf("error getting submodules of repository worktree: %s", err)
	}
	if len(submodules) == 0 {
		return nil
	}

	for _, submodule := range submodules {
		submoduleStatus, err := submodule.Status()
		if err != nil {
			return fmt.Errorf("error getting submodule %q status: %s", submodule.Config().Name, err)
		}

		if submoduleRepository, err := submodule.Repository(); err != nil {
			return fmt.Errorf("error getting submodule %q repository handle: %s", submodule.Config().Name, err)
		} else {
			if err := f(ctx, submoduleRepository, submodule.Config(), submoduleStatus); err == stopYield {
				return nil
			} else if err != nil {
				return err
			}
		}
	}

	return nil
}

func (repo *Base) checkAndReadSymlinkInRepository(ctx context.Context, repository *git.Repository, commit, path string) (bool, []byte, error) {
	var symlinkFound bool
	var queue []string = []string{path}

	for len(queue) > 0 {
		var p string
		p, queue = queue[0], queue[1:]

		if isSymlink, linkDest, err := repo.doCheckAndReadCommitSymlinkInRepository(ctx, repository, commit, p); err != nil {
			return false, nil, fmt.Errorf("error checking %q: %s", p, err)
		} else if isSymlink {
			symlinkFound = true
			queue = append(queue, string(linkDest))
		} else {
			return symlinkFound, []byte(p), nil
		}
	}

	panic("unexpected condition")
}

func (repo *Base) readFile(ctx context.Context, workTreeCacheDir, repoPath, gitDir, commit, path string) ([]byte, error) {
	repository, err := git.PlainOpenWithOptions(repoPath, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return nil, fmt.Errorf("cannot open repo %s: %s", repoPath, err)
	}

	commitHash, err := newHash(commit)
	if err != nil {
		return nil, fmt.Errorf("bad commit hash %q: %s", commit, err)
	}

	commitObj, err := repository.CommitObject(commitHash)
	if err != nil {
		return nil, fmt.Errorf("bad commit %s: %s", commit, err)
	}

	hasSubmodules, err := HasSubmodulesInCommit(commitObj)
	if err != nil {
		return nil, err
	}

	if hasSubmodules {
		var res []byte

		err = true_git.WithWorkTree(ctx, gitDir, workTreeCacheDir, commit, true_git.WithWorkTreeOptions{HasSubmodules: true}, func(worktreeDir string) error {
			if repositoryWithPreparedWorktree, err := true_git.GitOpenWithCustomWorktreeDir(gitDir, worktreeDir); err != nil {
				return err
			} else {
				res, err = repo.readFileInRepository(ctx, repositoryWithPreparedWorktree, commit, path)
				return err
			}
		})

		return res, err
	} else {
		return repo.readFileInRepository(ctx, repository, commit, path)
	}
}

func (repo *Base) readFileInRepository(ctx context.Context, repository *git.Repository, commit, path string) ([]byte, error) {
	commitHash, err := newHash(commit)
	if err != nil {
		return nil, fmt.Errorf("bad commit hash %q: %s", commit, err)
	}

	commitObj, err := repository.CommitObject(commitHash)
	if err != nil {
		return nil, fmt.Errorf("cannot get commit %q object: %s", commit, err)
	}

	realpath, err := repo.realpathInRepository(ctx, repository, commit, path)
	if err != nil {
		return nil, fmt.Errorf("error getting realpath for %q path: %s", path, err)
	}

	var result *struct {
		Data []byte
		Err  error
	}

	if err := YieldEachSubmoduleRepository(ctx, repository, func(ctx context.Context, submoduleRepository *git.Repository, submoduleConfig *config.Submodule, submoduleStatus *git.SubmoduleStatus) error {
		if !util.IsSubpathOfBasePath(submoduleConfig.Path, realpath) {
			return nil
		}

		pathInsideSubmodule := util.GetRelativeToBaseFilepath(submoduleConfig.Path, realpath)

		result = &struct {
			Data []byte
			Err  error
		}{}

		result.Data, result.Err = repo.readFileInRepository(ctx, submoduleRepository, submoduleStatus.Current.String(), pathInsideSubmodule)

		return stopYield
	}); err != nil {
		return nil, err
	} else if result != nil {
		return result.Data, result.Err
	}

	file, err := commitObj.File(realpath)
	if err != nil {
		return nil, fmt.Errorf("error getting repo file handle %q from commit %q: %s", realpath, commit, err)
	}

	content, err := file.Contents()
	if err != nil {
		return nil, err
	}

	return []byte(content), nil
}

func (repo *Base) realpathInRepository(ctx context.Context, repository *git.Repository, commit, filePath string) (string, error) {
	parts := util.SplitPath(filePath)

	var resolvedBasePath string

	for _, part := range parts {
		pathToResolve := path.Join(resolvedBasePath, part)

		if _, resolvedPath, err := repo.checkAndReadSymlinkInRepository(ctx, repository, commit, pathToResolve); err != nil {
			return "", fmt.Errorf("error reading link %q: %s", pathToResolve, err)
		} else {
			resolvedBasePath = string(resolvedPath)
		}
	}

	return resolvedBasePath, nil
}

func (repo *Base) isCommitFileExists(ctx context.Context, workTreeCacheDir, repoPath, gitDir, commit, path string) (bool, error) {
	repository, err := git.PlainOpenWithOptions(repoPath, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return false, fmt.Errorf("cannot open repo %s: %s", repoPath, err)
	}

	commitHash, err := newHash(commit)
	if err != nil {
		return false, fmt.Errorf("bad commit hash %q: %s", commit, err)
	}

	commitObj, err := repository.CommitObject(commitHash)
	if err != nil {
		return false, fmt.Errorf("bad commit %s: %s", commit, err)
	}

	hasSubmodules, err := HasSubmodulesInCommit(commitObj)
	if err != nil {
		return false, err
	}

	if hasSubmodules {
		var res bool

		err = true_git.WithWorkTree(ctx, gitDir, workTreeCacheDir, commit, true_git.WithWorkTreeOptions{HasSubmodules: true}, func(worktreeDir string) error {
			if repositoryWithPreparedWorktree, err := true_git.GitOpenWithCustomWorktreeDir(gitDir, worktreeDir); err != nil {
				return err
			} else {
				res, err = repo.isCommitFileExistsInRepository(ctx, repositoryWithPreparedWorktree, commit, path)
				return err
			}
		})

		return res, err
	} else {
		return repo.isCommitFileExistsInRepository(ctx, repository, commit, path)
	}
}

func (repo *Base) isCommitFileExistsInRepository(ctx context.Context, repository *git.Repository, commit, path string) (bool, error) {
	commitHash, err := newHash(commit)
	if err != nil {
		return false, fmt.Errorf("bad commit hash %q: %s", commit, err)
	}

	commitObj, err := repository.CommitObject(commitHash)
	if err != nil {
		return false, fmt.Errorf("cannot get commit %q object: %s", commit, err)
	}

	realpath, err := repo.realpathInRepository(ctx, repository, commit, path)
	if err != nil {
		return false, fmt.Errorf("error getting realpath for %q path: %s", path, err)
	}

	var result *struct {
		IsExists bool
		Err      error
	}

	if err := YieldEachSubmoduleRepository(ctx, repository, func(ctx context.Context, submoduleRepository *git.Repository, submoduleConfig *config.Submodule, submoduleStatus *git.SubmoduleStatus) error {
		if !util.IsSubpathOfBasePath(submoduleConfig.Path, realpath) {
			return nil
		}

		pathInsideSubmodule := util.GetRelativeToBaseFilepath(submoduleConfig.Path, realpath)

		result = &struct {
			IsExists bool
			Err      error
		}{}

		result.IsExists, result.Err = repo.isCommitFileExistsInRepository(ctx, submoduleRepository, submoduleStatus.Current.String(), pathInsideSubmodule)

		return stopYield
	}); err != nil {
		return false, err
	} else if result != nil {
		return result.IsExists, result.Err
	}

	if _, err := commitObj.File(realpath); err == object.ErrFileNotFound {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("error getting repo file %q from commit %q: %s", realpath, commit, err)
	}
	return true, nil
}

func (repo *Base) tagsList(repoPath string) ([]string, error) {
	repository, err := git.PlainOpenWithOptions(repoPath, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return nil, fmt.Errorf("cannot open repo `%s`: %s", repoPath, err)
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
		return nil, fmt.Errorf("cannot open repo `%s`: %s", repoPath, err)
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

func (repo *Base) checksumWithLsTree(ctx context.Context, repoPath, gitDir, workTreeCacheDir string, opts ChecksumOptions) (Checksum, error) {
	repository, err := git.PlainOpenWithOptions(repoPath, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return nil, fmt.Errorf("cannot open repo `%s`: %s", repoPath, err)
	}

	commitHash, err := newHash(opts.Commit)
	if err != nil {
		return nil, fmt.Errorf("bad commit hash `%s`: %s", opts.Commit, err)
	}

	commit, err := repository.CommitObject(commitHash)
	if err != nil {
		return nil, fmt.Errorf("bad commit `%s`: %s", opts.Commit, err)
	}

	hasSubmodules, err := HasSubmodulesInCommit(commit)
	if err != nil {
		return nil, err
	}

	checksum := &ChecksumDescriptor{
		NoMatchPaths: make([]string, 0),
		Hash:         sha256.New(),
	}

	err = true_git.WithWorkTree(ctx, gitDir, workTreeCacheDir, opts.Commit, true_git.WithWorkTreeOptions{HasSubmodules: hasSubmodules}, func(worktreeDir string) error {
		repositoryWithPreparedWorktree, err := true_git.GitOpenWithCustomWorktreeDir(gitDir, worktreeDir)
		if err != nil {
			return err
		}

		pathMatcher := path_matcher.NewGitMappingPathMatcher(
			opts.BasePath,
			opts.IncludePaths,
			opts.ExcludePaths,
			false,
		)

		var mainLsTreeResult *ls_tree.Result
		if err := logboek.Context(ctx).Debug().LogProcess("ls-tree (%s)", pathMatcher.String()).DoError(func() error {
			mainLsTreeResult, err = ls_tree.LsTree(ctx, repositoryWithPreparedWorktree, opts.Commit, pathMatcher, true)
			return err
		}); err != nil {
			return err
		}

		for _, path := range opts.Paths {
			var pathLsTreeResult *ls_tree.Result
			pathMatcher := path_matcher.NewSimplePathMatcher(
				opts.BasePath,
				[]string{path},
				false,
			)

			logProcess := logboek.Context(ctx).Debug().LogProcess("ls-tree (%s)", pathMatcher.String())
			logProcess.Start()
			pathLsTreeResult, err = mainLsTreeResult.LsTree(ctx, pathMatcher)
			if err != nil {
				logProcess.Fail()
				return err
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
				checksum.NoMatchPaths = append(checksum.NoMatchPaths, path)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return checksum, nil
}
