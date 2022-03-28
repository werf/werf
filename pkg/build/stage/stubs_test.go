package stage

import (
	"context"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/giterminism_manager"
	"github.com/werf/werf/pkg/image"
)

type LegacyImageStub struct {
	container_runtime.LegacyImageInterface

	_Container *LegacyContainerStub
}

func NewLegacyImageStub() *LegacyImageStub {
	return &LegacyImageStub{
		_Container: NewLegacyContainerStub(),
	}
}

func (img *LegacyImageStub) Container() container_runtime.LegacyContainer {
	return img._Container
}

type LegacyContainerStub struct {
	container_runtime.LegacyContainer

	_ServiceCommitChangeOptions *LegacyContainerOptionsStub
}

func NewLegacyContainerStub() *LegacyContainerStub {
	return &LegacyContainerStub{
		_ServiceCommitChangeOptions: NewLegacyContainerOptionsStub(),
	}
}

func (c *LegacyContainerStub) ServiceCommitChangeOptions() container_runtime.LegacyContainerOptions {
	return c._ServiceCommitChangeOptions
}

type LegacyContainerOptionsStub struct {
	container_runtime.LegacyContainerOptions

	Env map[string]string
}

func NewLegacyContainerOptionsStub() *LegacyContainerOptionsStub {
	return &LegacyContainerOptionsStub{Env: make(map[string]string)}
}

func (opts *LegacyContainerOptionsStub) AddEnv(envs map[string]string) {
	for k, v := range envs {
		opts.Env[k] = v
	}
}

type ConveyorStub struct {
	Conveyor

	giterminismManager            *GiterminismManagerStub
	lastStageImageNameByImageName map[string]string
	lastStageImageIDByImageName   map[string]string
}

func NewConveyorStub(giterminismManager *GiterminismManagerStub, lastStageImageNameByImageName, lastStageImageIDByImageName map[string]string) *ConveyorStub {
	return &ConveyorStub{
		giterminismManager:            giterminismManager,
		lastStageImageNameByImageName: lastStageImageNameByImageName,
		lastStageImageIDByImageName:   lastStageImageIDByImageName,
	}
}

func NewConveyorStubForDependencies(giterminismManager *GiterminismManagerStub, dependencies []*TestDependency) *ConveyorStub {
	lastStageImageNameByImageName := make(map[string]string)
	lastStageImageIDByImageName := make(map[string]string)

	for _, dep := range dependencies {
		lastStageImageNameByImageName[dep.ImageName] = dep.GetDockerImageName()
		lastStageImageIDByImageName[dep.ImageName] = dep.DockerImageID
	}

	return NewConveyorStub(giterminismManager, lastStageImageNameByImageName, lastStageImageIDByImageName)
}

func (c *ConveyorStub) GetImageNameForLastImageStage(imageName string) string {
	return c.lastStageImageNameByImageName[imageName]
}

func (c *ConveyorStub) GetImageIDForLastImageStage(imageName string) string {
	return c.lastStageImageIDByImageName[imageName]
}

func (c *ConveyorStub) GiterminismManager() giterminism_manager.Interface {
	return c.giterminismManager
}

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

type ContainerRuntimeMock struct {
	container_runtime.ContainerRuntime

	_PulledImages map[string]bool
}

func NewContainerRuntimeMock() *ContainerRuntimeMock {
	return &ContainerRuntimeMock{
		_PulledImages: make(map[string]bool),
	}
}

func (containerRuntime *ContainerRuntimeMock) HasContainerRootMountSupport() bool {
	return false
}

func (containerRuntime *ContainerRuntimeMock) GetImageInfo(ctx context.Context, ref string, opts container_runtime.GetImageInfoOpts) (*image.Info, error) {
	return nil, nil
}

func (containerRuntime *ContainerRuntimeMock) Pull(ctx context.Context, ref string, opts container_runtime.PullOpts) error {
	containerRuntime._PulledImages[ref] = true
	return nil
}

type DockerRegistryApiStub struct {
	docker_registry.ApiInterface
}

func NewDockerRegistryApiStub() *DockerRegistryApiStub {
	return &DockerRegistryApiStub{}
}

func (dockerRegistry *DockerRegistryApiStub) GetRepoImageConfigFile(ctx context.Context, reference string) (*v1.ConfigFile, error) {
	return &v1.ConfigFile{
		Config: v1.Config{},
	}, nil
}
