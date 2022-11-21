package instruction

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/dockerfile"
)

type Base[T dockerfile.InstructionDataInterface, BT container_backend.InstructionInterface] struct {
	*stage.BaseStage

	instruction        *dockerfile.DockerfileStageInstruction[T]
	backendInstruction BT
	dependencies       []*config.Dependency
	hasPrevStage       bool
}

func NewBase[T dockerfile.InstructionDataInterface, BT container_backend.InstructionInterface](name stage.StageName, instruction *dockerfile.DockerfileStageInstruction[T], backendInstruction BT, dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Base[T, BT] {
	return &Base[T, BT]{
		BaseStage:          stage.NewBaseStage(name, opts),
		instruction:        instruction,
		backendInstruction: backendInstruction,
		dependencies:       dependencies,
		hasPrevStage:       hasPrevStage,
	}
}

func (stg *Base[T, BT]) HasPrevStage() bool {
	return stg.hasPrevStage
}

func (stg *Base[T, BT]) IsStapelStage() bool {
	return false
}

func (stg *Base[T, BT]) UsesBuildContext() bool {
	return stg.backendInstruction.UsesBuildContext()
}

func (stg *Base[T, BT]) getDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver, expander InstructionExpander) ([]string, error) {
	if err := expander.ExpandInstruction(ctx, c, cb, prevImage, prevBuiltImage, buildContextArchive); err != nil {
		return nil, fmt.Errorf("unable to expand instruction %q: %w", stg.instruction.Data.Name(), err)
	}
	return nil, nil
}

func (stg *Base[T, BT]) PrepareImage(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevBuiltImage, stageImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) error {
	stageImage.Builder.DockerfileStageBuilder().AppendInstruction(stg.backendInstruction)
	return nil
}

func (stg *Base[T, BT]) ExpandInstruction(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevBuiltImage, stageImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) error {
	dependenciesArgs := stage.ResolveDependenciesArgs(stg.dependencies, c)
	// NOTE: do not skip unset envs during second stage expansion
	return stg.instruction.Expand(dependenciesArgs, dockerfile.ExpandOptions{SkipUnsetEnv: false})
}

type InstructionExpander interface {
	ExpandInstruction(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevBuiltImage, stageImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) error
}
