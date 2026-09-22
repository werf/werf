package config

import (
	"context"

	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/giterminism_manager"
	"github.com/werf/werf/v2/pkg/path_matcher"
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

func (manager *GiterminismManagerStub) Inspector() giterminism_manager.Inspector {
	return &GiterminismInspectorStub{}
}

func (manager *GiterminismManagerStub) Dev() bool {
	return false
}

func (manager *GiterminismManagerStub) HeadCommit(ctx context.Context) string {
	commit, err := manager.localGitRepo.HeadCommitHash(ctx)
	Expect(err).To(Succeed())
	return commit
}

type GiterminismInspectorStub struct{}

var _ giterminism_manager.Inspector = (*GiterminismInspectorStub)(nil)

func (inspector *GiterminismInspectorStub) InspectCustomTags() error {
	return nil
}

func (inspector *GiterminismInspectorStub) InspectConfigGoTemplateRenderingEnv(ctx context.Context, envName string) error {
	return nil
}

func (inspector *GiterminismInspectorStub) InspectConfigStapelFromLatest() error {
	return nil
}

func (inspector *GiterminismInspectorStub) InspectConfigStapelGitBranch() error {
	return nil
}

func (inspector *GiterminismInspectorStub) InspectConfigStapelMountBuildDir() error {
	return nil
}

func (inspector *GiterminismInspectorStub) InspectConfigStapelMountFromPath(fromPath string) error {
	return nil
}

func (inspector *GiterminismInspectorStub) InspectConfigDockerfileContextAddFile(relPath string) error {
	return nil
}

func (inspector *GiterminismInspectorStub) InspectBuildContextFiles(ctx context.Context, matcher path_matcher.PathMatcher) error {
	return nil
}

func (inspector *GiterminismInspectorStub) InspectConfigSecretEnvAccepted(secret string) error {
	return nil
}

func (inspector *GiterminismInspectorStub) InspectConfigSecretSrcAccepted(secret string) error {
	return nil
}

func (inspector *GiterminismInspectorStub) InspectConfigSecretValueAccepted(secret string) error {
	return nil
}

func (inspector *GiterminismInspectorStub) InspectIncludesAllowUpdate() error {
	return nil
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

type ImageStub struct {
	ImageInterface

	name string
	deps DependsOn
}

func NewImageStub(name string, dependsOn DependsOn) *ImageStub {
	return &ImageStub{
		name: name,
		deps: dependsOn,
	}
}

func (image *ImageStub) GetName() string {
	return image.name
}

func (image *ImageStub) dependsOn() DependsOn {
	return image.deps
}
