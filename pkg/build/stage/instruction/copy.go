package instruction

import (
	"context"
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v3/pkg/build/stage"
	"github.com/werf/werf/v3/pkg/config"
	"github.com/werf/werf/v3/pkg/container_backend"
	backend_instruction "github.com/werf/werf/v3/pkg/container_backend/instruction"
	"github.com/werf/werf/v3/pkg/dockerfile"
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

	stg.backendInstruction.CopyCommand = *stg.instruction.Data

	sourceContents, err := stg.expandSourceContents(stg.instruction.Data.SourceContents, env, dockerfile.ExpandOptions{})
	if err != nil {
		return err
	}
	stg.backendInstruction.SourceContents = sourceContents

	if stg.instruction.Data.From != "" {
		if ds := stg.instruction.GetDependencyByStageRef(stg.instruction.Data.From); ds != nil {
			depStageImageName := c.GetImageContentTagName(stg.TargetPlatform(), ds.GetWerfImageName())
			stg.backendInstruction.From = depStageImageName
		}
	}

	return nil
}

func (stg *Copy) GetContentDependencies(ctx context.Context, c stage.Conveyor, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	sourcePaths, err := stg.contentSourcePaths(c, stg.instruction.Data.SourcePaths)
	if err != nil {
		return "", err
	}

	sourceContents, err := stg.expandSourceContents(stg.instruction.Data.SourceContents, stg.GetExpandedEnvForContent(), dockerfile.ExpandOptions{SkipUnsetEnv: true})
	if err != nil {
		return "", err
	}

	return stg.getDependencies(ctx, buildContextArchive, sourcePaths, sourceContents)
}

func (stg *Copy) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	return stg.getDependencies(ctx, buildContextArchive, stg.instruction.Data.SourcePaths, stg.backendInstruction.SourceContents)
}

func (stg *Copy) getDependencies(ctx context.Context, buildContextArchive container_backend.BuildContextArchiver, checksumSourcePaths []string, sourceContents []instructions.SourceContent) (string, error) {
	var args []string

	args = append(args, "From", stg.instruction.Data.From)
	args = append(args, append([]string{"Sources"}, stg.instruction.Data.SourcePaths...)...)
	args = append(args, "Dest", stg.instruction.Data.DestPath)
	args = append(args, "Chown", stg.instruction.Data.Chown)
	args = append(args, "Chmod", stg.instruction.Data.Chmod)
	args = append(args, "ExpandedFrom", stg.backendInstruction.From)

	// appended only when set to keep digests of already built COPY stages intact
	if stg.instruction.Data.Parents {
		args = append(args, "Parents", "true")
	}
	if len(stg.instruction.Data.ExcludePatterns) > 0 {
		args = append(args, append([]string{"ExcludePatterns"}, stg.instruction.Data.ExcludePatterns...)...)
	}
	for _, content := range sourceContents {
		args = append(args, "SourceContent", content.Path, content.Data)
	}

	if stg.UsesBuildContext() {
		if srcChecksum, err := buildContextArchive.CalculateGlobsChecksum(ctx, checksumSourcePaths, container_backend.CalculateGlobsChecksumOptions{IncludeMatchedPaths: stg.instruction.Data.Parents}); err != nil {
			return "", fmt.Errorf("unable to calculate build context globs checksum: %w", err)
		} else {
			args = append(args, "SourcesChecksum", srcChecksum)
		}
	}

	// TODO(ilya-lesikov): should checksum of files from other image be calculated if --from specified?

	// TODO(staged-dockerfile): support --link option: https://docs.docker.com/engine/reference/builder/#copy---link

	return util.Sha256Hash(args...), nil
}
