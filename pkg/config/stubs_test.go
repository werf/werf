package config

import (
	"context"

	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/giterminism_manager"
)

type GiterminismManagerStub struct {
	giterminism_manager.Interface

	localGitRepo git_repo.GitRepo
}

func NewGiterminismManagerStub(localGitRepo git_repo.GitRepo) *GiterminismManagerStub {
	return &GiterminismManagerStub{
		localGitRepo: localGitRepo,
	}
}

func (manager *GiterminismManagerStub) RelativeToGitProjectDir() string {
	return ""
}

func (manager *GiterminismManagerStub) LocalGitRepo() git_repo.GitRepo {
	return manager.localGitRepo
}

func (manager *GiterminismManagerStub) Dev() bool {
	return false
}

func (manager *GiterminismManagerStub) HeadCommit() string {
	commit, err := manager.localGitRepo.HeadCommitHash(context.Background())
	Expect(err).To(Succeed())
	return commit
}

type LocalGitRepoStub struct {
	git_repo.GitRepo

	headCommitHash string
}

func NewLocalGitRepoStub(headCommitHash string) *LocalGitRepoStub {
	return &LocalGitRepoStub{
		headCommitHash: headCommitHash,
	}
}

func (repo *LocalGitRepoStub) HeadCommitHash(ctx context.Context) (string, error) {
	return repo.headCommitHash, nil
}

func (repo *LocalGitRepoStub) GetOrCreateArchive(ctx context.Context, opts git_repo.ArchiveOptions) (git_repo.Archive, error) {
	return NewGitRepoArchiveStub(), nil
}

type GitRepoArchiveStub struct {
	git_repo.Archive
}

func NewGitRepoArchiveStub() *GitRepoArchiveStub {
	return &GitRepoArchiveStub{}
}

func (archive *GitRepoArchiveStub) GetFilePath() string {
	return "no-such-file"
}
