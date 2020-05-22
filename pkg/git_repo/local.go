package git_repo

import (
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"

	"github.com/flant/logboek"

	"github.com/flant/werf/pkg/git_repo/check_ignore"
	"github.com/flant/werf/pkg/git_repo/ls_tree"
	"github.com/flant/werf/pkg/git_repo/status"
	"github.com/flant/werf/pkg/path_matcher"
	"github.com/flant/werf/pkg/true_git"
	"github.com/flant/werf/pkg/util"
)

type Local struct {
	Base
	Path   string
	GitDir string
}

func OpenLocalRepo(name string, path string) (*Local, error) {
	_, err := git.PlainOpen(path)
	if err != nil {
		if err == git.ErrRepositoryNotExists {
			return nil, nil
		}

		return nil, err
	}

	gitDir, err := true_git.GetRealRepoDir(filepath.Join(path, ".git"))
	if err != nil {
		return nil, fmt.Errorf("unable to get real git repo dir for %s: %s", path, err)
	}

	return &Local{Base: Base{Name: name}, Path: path, GitDir: gitDir}, nil
}

func (repo *Local) CreateVirtualMergeCommit(fromCommit, toCommit string) (string, error) {
	return repo.createVirtualMergeCommit(repo.GitDir, repo.Path, repo.getRepoWorkTreeCacheDir(), fromCommit, toCommit)
}

func (repo *Local) LsTree(pathMatcher path_matcher.PathMatcher) (*ls_tree.Result, error) {
	repository, err := git.PlainOpen(repo.Path)
	if err != nil {
		return nil, fmt.Errorf("cannot open repo `%s`: %s", repo.Path, err)
	}

	return ls_tree.LsTree(repository, pathMatcher)
}

func (repo *Local) Status(pathMatcher path_matcher.PathMatcher) (*status.Result, error) {
	repository, err := git.PlainOpen(repo.Path)
	if err != nil {
		return nil, fmt.Errorf("cannot open repo `%s`: %s", repo.Path, err)
	}

	return status.Status(repository, repo.Path, pathMatcher)
}

func (repo *Local) CheckIgnore(paths []string) (*check_ignore.Result, error) {
	repository, err := git.PlainOpen(repo.Path)
	if err != nil {
		return nil, fmt.Errorf("cannot open repo `%s`: %s", repo.Path, err)
	}

	return check_ignore.CheckIgnore(repository, repo.Path, paths)
}

func (repo *Local) FindCommitIdByMessage(regex string) (string, error) {
	head, err := repo.HeadCommit()
	if err != nil {
		return "", fmt.Errorf("error getting head commit: %s", err)
	}
	return repo.findCommitIdByMessage(repo.Path, regex, head)
}

func (repo *Local) IsEmpty() (bool, error) {
	return repo.isEmpty(repo.Path)
}

func (repo *Local) IsAncestor(ancestorCommit, descendantCommit string) (bool, error) {
	return true_git.IsAncestor(ancestorCommit, descendantCommit, repo.GitDir)
}

func (repo *Local) RemoteOriginUrl() (string, error) {
	return repo.remoteOriginUrl(repo.Path)
}

func (repo *Local) HeadCommit() (string, error) {
	ref, err := repo.getReferenceForRepo(repo.Path)
	if err != nil {
		return "", fmt.Errorf("cannot get repo `%s` head ref: %s", repo.Path, err)
	}
	return fmt.Sprintf("%s", ref.Hash()), nil
}

func (repo *Local) IsHeadReferenceExist() (bool, error) {
	_, err := repo.getReferenceForRepo(repo.Path)
	if err != nil {
		if err == plumbing.ErrReferenceNotFound {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (repo *Local) HeadBranchName() (string, error) {
	return repo.getHeadBranchName(repo.Path)
}

func (repo *Local) CreatePatch(opts PatchOptions) (Patch, error) {
	return repo.createPatch(repo.Path, repo.GitDir, repo.getRepoWorkTreeCacheDir(), opts)
}

func (repo *Local) CreateArchive(opts ArchiveOptions) (Archive, error) {
	return repo.createArchive(repo.Path, repo.GitDir, repo.getRepoWorkTreeCacheDir(), opts)
}

func (repo *Local) Checksum(opts ChecksumOptions) (checksum Checksum, err error) {
	_ = logboek.Debug.LogProcess(
		"Calculating checksum",
		logboek.LevelLogProcessOptions{},
		func() error {
			checksum, err = repo.checksumWithLsTree(repo.Path, repo.GitDir, repo.getRepoWorkTreeCacheDir(), opts)
			return nil
		},
	)

	return
}

func (repo *Local) IsCommitExists(commit string) (bool, error) {
	return repo.isCommitExists(repo.Path, repo.GitDir, commit)
}

func (repo *Local) TagsList() ([]string, error) {
	return repo.tagsList(repo.Path)
}

func (repo *Local) RemoteBranchesList() ([]string, error) {
	return repo.remoteBranchesList(repo.Path)
}

func (repo *Local) getRepoWorkTreeCacheDir() string {
	absPath, err := filepath.Abs(repo.Path)
	if err != nil {
		panic(err) // stupid interface of filepath.Abs
	}

	fullPath := filepath.Clean(absPath)
	repoId := util.Sha256Hash(fullPath)

	return filepath.Join(GetWorkTreeCacheDir(), "local", repoId)
}

func (repo *Local) IsBranchState() bool {
	_, err := repo.HeadBranchName()
	if err == errNotABranch {
		return false
	} else if err != nil {
		logboek.LogWarnF("ERROR: Getting branch of local git: %s\n", err)
		return false
	}
	return true
}

func (repo *Local) GetCurrentBranchName() string {
	name, err := repo.HeadBranchName()
	if err != nil {
		logboek.LogWarnF("ERROR: Getting branch of local git: %s\n", err)
		return ""
	}
	return name
}

func (repo *Local) IsTagState() bool {
	return repo.GetCurrentTagName() != ""
}

func (repo *Local) findTagByCommitID(repoPath string, commitID plumbing.Hash) (string, error) {
	repository, err := git.PlainOpen(repoPath)
	if err != nil {
		return "", fmt.Errorf("cannot open repo `%s`: %s", repoPath, err)
	}

	references, err := repository.References()
	if err != nil {
		return "", err
	}

	tagPrefix := "refs/tags/"

	var res *plumbing.Reference

	err = references.ForEach(func(r *plumbing.Reference) error {
		refName := r.Name().String()
		if strings.HasPrefix(refName, tagPrefix) {
			if r.Hash() == commitID {
				res = r
				return storer.ErrStop
			}
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	if res != nil {
		return strings.TrimPrefix(res.Name().String(), tagPrefix), nil
	}
	return "", nil
}

func (repo *Local) GetCurrentTagName() string {
	ref, err := repo.getReferenceForRepo(repo.Path)
	if err != nil {
		logboek.LogWarnF("ERROR: Cannot get local git repo head ref: %s\n", err)
		return ""
	}

	tag, err := repo.findTagByCommitID(repo.Path, ref.Hash())
	if err != nil {
		logboek.LogWarnF("ERROR: Cannot get local git repo tag: %s\n", err)
		return ""
	}
	return tag
}

func (repo *Local) GetHeadCommit() string {
	ref, err := repo.getReferenceForRepo(repo.Path)
	if err != nil {
		logboek.LogWarnF("ERROR: Getting HEAD commit id of local git repo: %s\n", err)
		return ""
	}
	return fmt.Sprintf("%s", ref.Hash())
}
