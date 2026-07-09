package stage

import (
	"context"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/giterminism_manager"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/path_matcher"
)

type LegacyImageStub struct {
	container_backend.LegacyImageInterface
}

func NewLegacyImageStub() *LegacyImageStub {
	return &LegacyImageStub{}
}

func (img *LegacyImageStub) GetBuildServiceLabels() map[string]string {
	return map[string]string{
		"stub": "true",
	}
}

type ConveyorStub struct {
	Conveyor

	giterminismManager              *GiterminismManagerStub
	lastStageImageNameByImageName   map[string]string
	lastStageImageDigestByImageName map[string]string
}

func NewConveyorStub(giterminismManager *GiterminismManagerStub, lastStageImageNameByImageName, lastStageImageDigestByImageName map[string]string) *ConveyorStub {
	return &ConveyorStub{
		giterminismManager:              giterminismManager,
		lastStageImageNameByImageName:   lastStageImageNameByImageName,
		lastStageImageDigestByImageName: lastStageImageDigestByImageName,
	}
}

func (c *ConveyorStub) GetImageContentTagDigest(targetPlatform, imageName string) string {
	return c.lastStageImageDigestByImageName[imageName]
}

func (c *ConveyorStub) GiterminismManager() giterminism_manager.Interface {
	return c.giterminismManager
}

func (c *ConveyorStub) GetImageContentTagStageID(targetPlatform, imageName string) string {
	return c.lastStageImageNameByImageName[imageName]
}

func (c *ConveyorStub) GetImageContentTagName(targetPlatform, imageName string) string {
	return c.lastStageImageNameByImageName[imageName]
}

type GiterminismInspectorStub struct {
	giterminism_manager.Inspector
}

func NewGiterminismInspectorStub() *GiterminismInspectorStub {
	return &GiterminismInspectorStub{}
}

func (inspector *GiterminismInspectorStub) InspectBuildContextFiles(ctx context.Context, matcher path_matcher.PathMatcher) error {
	return nil
}

type GiterminismManagerStub struct {
	giterminism_manager.Interface

	inspector    giterminism_manager.Inspector
	localGitRepo git_repo.GitRepo
}

func NewGiterminismManagerStub(localGitRepo git_repo.GitRepo, inspector giterminism_manager.Inspector) *GiterminismManagerStub {
	return &GiterminismManagerStub{
		localGitRepo: localGitRepo,
		inspector:    inspector,
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

func (manager *GiterminismManagerStub) HeadCommit(ctx context.Context) string {
	commit, err := manager.localGitRepo.HeadCommitHash(ctx)
	Expect(err).To(Succeed())
	return commit
}

func (manager *GiterminismManagerStub) Inspector() giterminism_manager.Inspector {
	return manager.inspector
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

func (repo *LocalGitRepoStub) GetOrCreateChecksum(ctx context.Context, opts git_repo.ChecksumOptions) (string, error) {
	return repo.headCommitHash, nil
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

type ContainerBackendStub struct {
	container_backend.ContainerBackend

	_PulledImages map[string]bool
}

func NewContainerBackendStub() *ContainerBackendStub {
	return &ContainerBackendStub{
		_PulledImages: make(map[string]bool),
	}
}

func (containerBackend *ContainerBackendStub) HasContainerRootMountSupport() bool {
	return false
}

func (containerBackend *ContainerBackendStub) GetImageInfo(ctx context.Context, ref string, opts container_backend.GetImageInfoOpts) (*image.Info, error) {
	return nil, nil
}

func (containerBackend *ContainerBackendStub) Pull(ctx context.Context, ref string, opts container_backend.PullOpts) error {
	containerBackend._PulledImages[ref] = true
	return nil
}

type DockerRegistryApiStub struct {
	docker_registry.GenericApiInterface
}

func NewDockerRegistryApiStub() *DockerRegistryApiStub {
	return &DockerRegistryApiStub{}
}

func (dockerRegistry *DockerRegistryApiStub) GetRepoImageConfigFile(ctx context.Context, reference string) (*v1.ConfigFile, error) {
	return &v1.ConfigFile{
		Config: v1.Config{},
	}, nil
}
