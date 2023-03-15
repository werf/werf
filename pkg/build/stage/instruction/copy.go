package instruction

import (
	"context"
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"

	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_backend"
	backend_instruction "github.com/werf/werf/pkg/container_backend/instruction"
	"github.com/werf/werf/pkg/dockerfile"
	"github.com/werf/werf/pkg/util"
)

type Copy struct {
	*Base[*instructions.CopyCommand, *backend_instruction.Copy]
}

func NewCopy(i *dockerfile.DockerfileStageInstruction[*instructions.CopyCommand], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Copy {
	return &Copy{Base: NewBase(i, backend_instruction.NewCopy(*i.Data), dependencies, hasPrevStage, opts)}
}

func (stg *Copy) ExpandDependencies(ctx context.Context, c stage.Conveyor, baseEnv map[string]string) error {
	return stg.doExpandDependencies(ctx, c, baseEnv, stg)
}

func (stg *Copy) ExpandInstruction(c stage.Conveyor, env map[string]string) error {
	if err := stg.Base.ExpandInstruction(c, env); err != nil {
		return err
	}

	if stg.instruction.Data.From != "" {
		if ds := stg.instruction.GetDependencyByStageRef(stg.instruction.Data.From); ds != nil {
			depStageImageName := c.GetImageNameForLastImageStage(stg.TargetPlatform(), ds.GetWerfImageName())
			stg.backendInstruction.From = depStageImageName
		}
	}

	return nil
}

func (stg *Copy) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	var args []string

	args = append(args, "From", stg.instruction.Data.From)
	args = append(args, append([]string{"Sources"}, stg.instruction.Data.SourcePaths...)...)
	args = append(args, "Dest", stg.instruction.Data.DestPath)
	args = append(args, "Chown", stg.instruction.Data.Chown)
	args = append(args, "Chmod", stg.instruction.Data.Chmod)
	args = append(args, "ExpandedFrom", stg.backendInstruction.From)

	if stg.UsesBuildContext() {
		if srcChecksum, err := buildContextArchive.CalculateGlobsChecksum(ctx, stg.instruction.Data.SourcePaths, false); err != nil {
			return "", fmt.Errorf("unable to calculate build context globs checksum: %w", err)
		} else {
			args = append(args, "SourcesChecksum", srcChecksum)
		}
	}

	// TODO(ilya-lesikov): should checksum of files from other image be calculated if --from specified?

	// TODO(staged-dockerfile): support --link option: https://docs.docker.com/engine/reference/builder/#copy---link

	return util.Sha256Hash(args...), nil
}
