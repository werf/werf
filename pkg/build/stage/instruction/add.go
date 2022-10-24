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

type Add struct {
	*Base[*dockerfile_instruction.Add]
}

func NewAdd(name stage.StageName, i *dockerfile.DockerfileStageInstruction[*dockerfile_instruction.Add], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Add {
	return &Add{Base: NewBase(name, i, backend_instruction.NewAdd(*i.Data), dependencies, hasPrevStage, opts)}
}

func (stage *Add) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	var args []string

	args = append(args, stage.instruction.Data.Name())
	args = append(args, stage.instruction.Data.Raw)
	args = append(args, stage.instruction.Data.Src...)
	args = append(args, stage.instruction.Data.Dst)
	args = append(args, stage.instruction.Data.Chown)
	args = append(args, stage.instruction.Data.Chmod)

	// TODO(staged-dockerfile): support http src and --checksum option: https://docs.docker.com/engine/reference/builder/#verifying-a-remote-file-checksum-add---checksumchecksum-http-src-dest
	// TODO(staged-dockerfile): support git ref: https://docs.docker.com/engine/reference/builder/#adding-a-git-repository-add-git-ref-dir
	// TODO(staged-dockerfile): support --keep-git-dir for git: https://docs.docker.com/engine/reference/builder/#adding-a-git-repository-add-git-ref-dir
	// TODO(staged-dockerfile): support --link

	pathsChecksum, err := buildContextArchive.CalculatePathsChecksum(ctx, stage.instruction.Data.Src)
	if err != nil {
		return "", fmt.Errorf("unable to calculate build context paths checksum: %w", err)
	}
	args = append(args, fmt.Sprintf("src-checksum=%s", pathsChecksum))

	return util.Sha256Hash(args...), nil
}
