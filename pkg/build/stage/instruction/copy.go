package instruction

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_backend"
	backend_instruction "github.com/werf/werf/pkg/container_backend/instruction"
	"github.com/werf/werf/pkg/dockerfile"
	dockerfile_instruction "github.com/werf/werf/pkg/dockerfile/instruction"
	"github.com/werf/werf/pkg/util"
)

type Copy struct {
	*Base[*dockerfile_instruction.Copy, *backend_instruction.Copy]
}

func NewCopy(name stage.StageName, i *dockerfile.DockerfileStageInstruction[*dockerfile_instruction.Copy], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Copy {
	return &Copy{Base: NewBase(name, i, backend_instruction.NewCopy(*i.Data), dependencies, hasPrevStage, opts)}
}

func (stg *Copy) ExpandInstruction(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevBuiltImage, stageImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) error {
	if stg.instruction.Data.From != "" {
		if ds := stg.instruction.GetDependencyByStageRef(stg.instruction.Data.From); ds != nil {
			depStageImageName := c.GetImageNameForLastImageStage(ds.WerfImageName())
			stg.backendInstruction.From = depStageImageName
		}
	}

	return nil
}

func (stg *Copy) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	args, err := stg.getDependencies(ctx, c, cb, prevImage, prevBuiltImage, buildContextArchive, stg)
	if err != nil {
		return "", err
	}

	args = append(args, "Instruction", stg.instruction.Data.Name())
	args = append(args, "From", stg.instruction.Data.From)
	args = append(args, append([]string{"Src"}, stg.instruction.Data.Src...)...)
	args = append(args, "Dst", stg.instruction.Data.Dst)
	args = append(args, "Chown", stg.instruction.Data.Chown)
	args = append(args, "Chmod", stg.instruction.Data.Chmod)
	args = append(args, "ExpandedFrom", stg.backendInstruction.From)

	if stg.UsesBuildContext() {
		if srcChecksum, err := buildContextArchive.CalculateGlobsChecksum(ctx, stg.instruction.Data.Src, false); err != nil {
			return "", fmt.Errorf("unable to calculate build context globs checksum: %w", err)
		} else {
			args = append(args, "SrcChecksum", srcChecksum)
		}
	}

	// TODO(ilya-lesikov): should checksum of files from other image be calculated if --from specified?

	// TODO(staged-dockerfile): support --link option: https://docs.docker.com/engine/reference/builder/#copy---link

	return util.Sha256Hash(args...), nil
}
