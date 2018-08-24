package git_repo

import (
	"fmt"
	"io"
	"strings"

	git "github.com/flant/go-git"
	"github.com/flant/go-git/plumbing"
	"github.com/flant/go-git/plumbing/object"
	"github.com/flant/go-git/storage"
)

type Base struct {
	Name string
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

func (repo *Base) archiveType(repoPath string, opts ArchiveOptions) (ArchiveType, error) {
	archive, err := repo.createArchiveObject(repoPath, opts)
	if err != nil {
		return "", err
	}

	return archive.Type()
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

func (repo *Base) createPatchObject(repoPath string, opts PatchOptions) (*RelativeFilteredPatch, error) {
	repository, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open repo: %s", err)
	}

	fromHash := plumbing.NewHash(opts.FromCommit)
	fromCommitObj, err := repository.CommitObject(fromHash)
	if err != nil {
		return nil, fmt.Errorf("failed to  `%s`: %s", opts.FromCommit, err)
	}

	toHash := plumbing.NewHash(opts.ToCommit)
	toCommitObj, err := repository.CommitObject(toHash)
	if err != nil {
		return nil, fmt.Errorf("cannot create tree for commit `%s`: %s", opts.ToCommit, err)
	}

	originPatch, err := fromCommitObj.Patch(toCommitObj)
	if err != nil {
		return nil, fmt.Errorf("cannot create patch between `%s` and `%s`: %s", opts.FromCommit, opts.ToCommit, err)
	}

	return &RelativeFilteredPatch{
		OriginPatch: originPatch,
		PathFilter: PathFilter{
			BasePath:     opts.BasePath,
			IncludePaths: opts.IncludePaths,
			ExcludePaths: opts.ExcludePaths,
		},
	}, nil
}

func (repo *Base) createPatch(repoPath string, output io.Writer, opts PatchOptions) error {
	patchObj, err := repo.createPatchObject(repoPath, opts)
	if err != nil {
		return err
	}

	return patchObj.Encode(output)
}

func (repo *Base) isAnyChanges(repoPath string, opts PatchOptions) (bool, error) {
	patchObj, err := repo.createPatchObject(repoPath, opts)

	if err != nil {
		return false, err
	}

	return len(patchObj.FilePatches()) != 0, nil
}

func (repo *Base) createArchiveObject(repoPath string, opts ArchiveOptions) (*Archive, error) {
	repository, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open repo: %s", err)
	}

	commitHash := plumbing.NewHash(opts.Commit)
	commitObj, err := repository.CommitObject(commitHash)
	if err != nil {
		return nil, fmt.Errorf("bad commit `%s`: %s", opts.Commit, err)
	}

	tree, err := commitObj.Tree()
	if err != nil {
		return nil, err
	}

	archive := &Archive{
		PathFilter: PathFilter{
			BasePath:     opts.BasePath,
			IncludePaths: opts.IncludePaths,
			ExcludePaths: opts.ExcludePaths,
		},
		Repo: struct {
			Tree   *object.Tree
			Storer storage.Storer
		}{tree, repository.Storer},
	}

	return archive, nil
}

func (repo *Base) isAnyEntries(repoPath string, opts ArchiveOptions) (bool, error) {
	archiveObj, err := repo.createArchiveObject(repoPath, opts)
	if err != nil {
		return false, err
	}

	res, err := archiveObj.IsAnyEntries()
	if err != nil {
		return false, err
	}

	return res, nil
}

func (repo *Base) createArchiveTar(repoPath string, output io.Writer, opts ArchiveOptions) error {
	archiveObj, err := repo.createArchiveObject(repoPath, opts)
	if err != nil {
		return err
	}

	return archiveObj.CreateTar(output)
}

func (repo *Base) hasBinaryPatches(repoPath string, opts PatchOptions) (bool, error) {
	patchObj, err := repo.createPatchObject(repoPath, opts)
	if err != nil {
		return false, err
	}

	for _, fpatch := range patchObj.FilePatches() {
		if fpatch.IsBinary() {
			return true, nil
		}
	}

	return false, nil
}
