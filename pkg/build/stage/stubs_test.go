package stage

import (
	"context"

	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/giterminism_manager"
)

type LegacyImageStub struct {
	container_runtime.LegacyImageInterface

	_Container              *LegacyContainerStub
	_DockerfileImageBuilder *container_runtime.DockerfileImageBuilder
}

func NewLegacyImageStub() *LegacyImageStub {
	return &LegacyImageStub{
		_Container: NewLegacyContainerStub(),
	}
}

func (img *LegacyImageStub) Container() container_runtime.LegacyContainer {
	return img._Container
}

func (img *LegacyImageStub) DockerfileImageBuilder() *container_runtime.DockerfileImageBuilder {
	if img._DockerfileImageBuilder == nil {
		img._DockerfileImageBuilder = container_runtime.NewDockerfileImageBuilder(nil)
	}
	return img._DockerfileImageBuilder
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
