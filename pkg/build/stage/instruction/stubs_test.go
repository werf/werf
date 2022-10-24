package instruction

import (
	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/container_backend/stage_builder"
)

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

func NewTestData(stg stage.Interface, expectedDigest string, files []*FileData) *TestData {
	conveyor := stage.NewConveyorStub(stage.NewGiterminismManagerStub(stage.NewLocalGitRepoStub("9d8059842b6fde712c58315ca0ab4713d90761c0"), stage.NewGiterminismInspectorStub()), nil, nil)
	containerBackend := stage.NewContainerBackendStub()

	img := stage.NewLegacyImageStub()
	stageBuilder := stage_builder.NewStageBuilder(containerBackend, "", img)
	stageImage := &stage.StageImage{
		Image:   img,
		Builder: stageBuilder,
	}

	buildContext := NewBuildContextStub(files)

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
