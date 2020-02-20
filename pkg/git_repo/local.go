package git_repo

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/flant/werf/pkg/true_git"

	"github.com/flant/werf/pkg/util"

	"github.com/flant/logboek"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
)

type Local struct {
	Base
	Path   string
	GitDir string
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
