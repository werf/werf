package stage_test

import (
	"context"

	"github.com/werf/werf/v3/pkg/git_repo"
)

var _ git_repo.GitRepo = (*gitArchiveRepoStub)(nil)

type gitArchiveRepoStub struct {
	*GitRepoStub
	checksum string
}

func (repo *gitArchiveRepoStub) GetOrCreateChecksum(_ context.Context, _ git_repo.ChecksumOptions) (string, error) {
	return repo.checksum, nil
}
