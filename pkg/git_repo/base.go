package git_repo

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/flant/dapp/pkg/true_git"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

type Base struct {
	Name   string
	TmpDir string
}

func (repo *Base) getHeadCommitForRepo(repoPath string) (string, error) {
	ref, err := repo.getReferenceForRepo(repoPath)
	if err != nil {
		return "", fmt.Errorf("cannot get repo head: %s", err)
	}

	return fmt.Sprintf("%s", ref.Hash()), nil
}

func (repo *Base) getHeadBranchNameForRepo(repoPath string) (string, error) {
	ref, err := repo.getReferenceForRepo(repoPath)
	if err != nil {
		return "", fmt.Errorf("cannot get repo head: %s", err)
	}

	if ref.Name().IsBranch() {
		branchRef := ref.Name()
		return strings.Split(string(branchRef), "refs/heads/")[1], nil
	} else {
		return "", fmt.Errorf("cannot get branch name: HEAD refers to a specific revision that is not associated with a branch name")
	}
}

func (repo *Base) getReferenceForRepo(repoPath string) (*plumbing.Reference, error) {
	var err error

	repository, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open repo: %s", err)
	}

	return repository.Head()
}

func (repo *Base) String() string {
	return repo.Name
}

func (repo *Base) HeadCommit() (string, error) {
	panic("not implemented")
}

func (repo *Base) HeadBranchName() (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (repo *Base) LatestBranchCommit(branch string) (string, error) {
	panic("not implemented")
}

func (repo *Base) LatestTagCommit(branch string) (string, error) {
	panic("not implemented")
}

func (repo *Base) ArchiveType(ArchiveOptions) (ArchiveType, error) {
	panic("not implemented")
}

func (repo *Base) CreatePatch(io.Writer, PatchOptions) error {
	panic("not implemented")
}

func (repo *Base) IsAnyChanges(PatchOptions) (bool, error) {
	panic("not implemented")
}

func (repo *Base) IsAnyEntries(ArchiveOptions) (bool, error) {
	panic("not implemented")
}

func (repo *Base) CreateArchiveTar(io.Writer, ArchiveOptions) error {
	panic("not implemented")
}

func (repo *Base) ArchiveChecksum(ArchiveOptions) (string, error) {
	panic("not implemented")
}

func (repo *Base) createPatch(repoPath string, opts PatchOptions) (Patch, error) {
	repository, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open repo: %s", err)
	}

	fromHash := plumbing.NewHash(opts.FromCommit)
	_, err = repository.CommitObject(fromHash)
	if err != nil {
		return nil, fmt.Errorf("bad `from` commit `%s`: %s", opts.FromCommit, err)
	}

	toHash := plumbing.NewHash(opts.ToCommit)
	toCommit, err := repository.CommitObject(toHash)
	if err != nil {
		return nil, fmt.Errorf("bad `to` commit `%s`: %s", opts.ToCommit, err)
	}

	hasSubmodules, err := HasSubmodulesInCommit(toCommit)
	if err != nil {
		return nil, err
	}

	patch := NewTmpPatchFile()

	fileHandler, err := os.OpenFile(patch.GetFilePath(), os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, fmt.Errorf("cannot open patch file `%s`: %s", patch.GetFilePath(), err)
	}

	patchOpts := true_git.PatchOptions{
		FromCommit: opts.FromCommit,
		ToCommit:   opts.ToCommit,
		PathFilter: true_git.PathFilter{
			BasePath:     opts.BasePath,
			IncludePaths: opts.IncludePaths,
			ExcludePaths: opts.ExcludePaths,
		},
	}

	var desc *true_git.PatchDescriptor

	if hasSubmodules {
		desc, err = true_git.PatchWithSubmodules(fileHandler, repoPath, "TODO WORK TREE WITH LOCK!!!", patchOpts)
	} else {
		desc, err = true_git.Patch(fileHandler, repoPath, patchOpts)
	}

	if err != nil {
		return nil, fmt.Errorf("error creating patch between `%s` and `%s` commits: %s", opts.FromCommit, opts.ToCommit, err)
	}

	patch.Descriptor = desc

	err = fileHandler.Close()
	if err != nil {
		return nil, fmt.Errorf("error creating patch file `%s`: %s", patch.GetFilePath(), err)
	}

	return patch, nil
}

func HasSubmodulesInCommit(commit *object.Commit) (bool, error) {
	_, err := commit.File(".gitmodules")
	if err == object.ErrFileNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	// TODO: change return value when submodules support available
	return false, nil
}

func (repo *Base) createArchive(repoPath string, opts ArchiveOptions) (Archive, error) {
	repository, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open repo: %s", err)
	}

	commitHash := plumbing.NewHash(opts.Commit)
	commit, err := repository.CommitObject(commitHash)
	if err != nil {
		return nil, fmt.Errorf("bad commit `%s`: %s", opts.Commit, err)
	}

	hasSubmodules, err := HasSubmodulesInCommit(commit)
	if err != nil {
		return nil, err
	}

	archive := NewTmpArchiveFile()

	fileHandler, err := os.OpenFile(archive.GetFilePath(), os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, fmt.Errorf("cannot open archive file: %s", err)
	}

	archiveOpts := true_git.ArchiveOptions{
		Commit: opts.Commit,
		PathFilter: true_git.PathFilter{
			BasePath:     opts.BasePath,
			IncludePaths: opts.IncludePaths,
			ExcludePaths: opts.ExcludePaths,
		},
	}

	var desc *true_git.ArchiveDescriptor

	if hasSubmodules {
		desc, err = true_git.ArchiveWithSubmodules(fileHandler, repoPath, "TODO WORK TREE WITH LOCK", archiveOpts)
	} else {
		desc, err = true_git.Archive(fileHandler, repoPath, "TODO WORK TREE WITH LOCK", archiveOpts)
	}

	if err != nil {
		return nil, fmt.Errorf("error creating archive for commit `%s`: %s", opts.Commit, err)
	}

	archive.Descriptor = desc

	return archive, nil
}
