package git_repo

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/lockgate"
	"github.com/werf/werf/v2/pkg/true_git"
	"github.com/werf/werf/v2/pkg/werf"
)

func TestGitRepoSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Git Repo Suite")
}

type fakeGitDataManager struct{}

var _ GitDataManager = (*fakeGitDataManager)(nil)

func (m *fakeGitDataManager) CreateArchiveFile(ctx context.Context, repoID string, opts ArchiveOptions, tmpPath string) (*ArchiveFile, error) {
	panic("not implemented")
}

func (m *fakeGitDataManager) GetArchiveFile(ctx context.Context, repoID string, opts ArchiveOptions) (*ArchiveFile, error) {
	panic("not implemented")
}

func (m *fakeGitDataManager) CreatePatchFile(ctx context.Context, repoID string, opts PatchOptions, tmpPath string, desc *true_git.PatchDescriptor) (*PatchFile, error) {
	panic("not implemented")
}

func (m *fakeGitDataManager) GetPatchFile(ctx context.Context, repoID string, opts PatchOptions) (*PatchFile, error) {
	panic("not implemented")
}

func (m *fakeGitDataManager) NewTmpFile() (string, error) {
	panic("not implemented")
}

func (m *fakeGitDataManager) LockGC(ctx context.Context, shared bool) (lockgate.LockHandle, error) {
	_, handle, err := werf.HostLocker().AcquireLock(ctx, "gc", lockgate.AcquireOptions{Shared: shared})
	return handle, err
}

func (m *fakeGitDataManager) GetArchivesCacheDir() string {
	panic("not implemented")
}

func (m *fakeGitDataManager) GetPatchesCacheDir() string {
	panic("not implemented")
}
