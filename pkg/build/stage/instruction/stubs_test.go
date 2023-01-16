package instruction

import (
	"context"
	"strings"

	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/container_backend/stage_builder"
	"github.com/werf/werf/pkg/dockerfile"
	"github.com/werf/werf/pkg/util"
)

func NewDockerfileStageInstructionWithDependencyStages[T dockerfile.InstructionDataInterface](data T, dependencyStages []string) *dockerfile.DockerfileStageInstruction[T] {
	i := dockerfile.NewDockerfileStageInstruction(data, dockerfile.DockerfileStageInstructionOptions{})
	for _, stageName := range dependencyStages {
		i.SetDependencyByStageRef(stageName, &dockerfile.DockerfileStage{StageName: stageName})
	}
	return i
}

type TestData struct {
	Stage          stage.Interface
	ExpectedDigest string

	Conveyor         *stage.ConveyorStub
	ContainerBackend *stage.ContainerBackendStub
	Image            *stage.LegacyImageStub
	StageBuilder     *stage_builder.StageBuilder
	StageImage       *stage.StageImage
	BuildContext     *BuildContextStub
}

type TestDataOptions struct {
	Files                                                                                       []*FileData
	LastStageImageNameByWerfImage, LastStageImageIDByWerfImage, LastStageImageDigestByWerfImage map[string]string
}

func NewTestData(stg stage.Interface, expectedDigest string, opts TestDataOptions) *TestData {
	conveyor := stage.NewConveyorStub(
		stage.NewGiterminismManagerStub(stage.NewLocalGitRepoStub("9d8059842b6fde712c58315ca0ab4713d90761c0"), stage.NewGiterminismInspectorStub()),
		opts.LastStageImageNameByWerfImage,
		opts.LastStageImageIDByWerfImage,
		opts.LastStageImageDigestByWerfImage,
	)
	containerBackend := stage.NewContainerBackendStub()

	img := stage.NewLegacyImageStub()
	stageBuilder := stage_builder.NewStageBuilder(containerBackend, "", img)
	stageImage := &stage.StageImage{
		Image:   img,
		Builder: stageBuilder,
	}

	buildContext := NewBuildContextStub(opts.Files)

	return &TestData{
		Stage:            stg,
		ExpectedDigest:   expectedDigest,
		Conveyor:         conveyor,
		ContainerBackend: containerBackend,
		Image:            img,
		StageBuilder:     stageBuilder,
		StageImage:       stageImage,
		BuildContext:     buildContext,
	}
}

type BuildContextStub struct {
	container_backend.BuildContextArchiver

	Files []*FileData
}

type FileData struct {
	Name string
	Data []byte
}

func NewBuildContextStub(files []*FileData) *BuildContextStub {
	return &BuildContextStub{Files: files}
}

func (buildContext *BuildContextStub) CalculateGlobsChecksum(ctx context.Context, globs []string, checkForArchive bool) (string, error) {
	var args []string

	for _, p := range globs {
		for _, f := range buildContext.Files {
			if f.Name == p {
				args = append(args, string(f.Data))
				break
			}
		}

		for _, f := range buildContext.Files {
			if strings.HasPrefix(f.Name, p) {
				args = append(args, string(f.Data))
				break
			}
		}
	}

	return util.Sha256Hash(args...), nil
}

func (buildContext *BuildContextStub) CalculatePathsChecksum(ctx context.Context, paths []string) (string, error) {
	var args []string

	for _, p := range paths {
		for _, f := range buildContext.Files {
			if f.Name == p {
				args = append(args, string(f.Data))
				break
			}
		}
	}

	return util.Sha256Hash(args...), nil
}
