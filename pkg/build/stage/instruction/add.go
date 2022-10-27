package instruction

import (
	"context"
	"fmt"
	"strings"

	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_backend"
	backend_instruction "github.com/werf/werf/pkg/container_backend/instruction"
	"github.com/werf/werf/pkg/dockerfile"
	dockerfile_instruction "github.com/werf/werf/pkg/dockerfile/instruction"
	"github.com/werf/werf/pkg/util"
)

type Add struct {
	*Base[*dockerfile_instruction.Add, *backend_instruction.Add]
}

func NewAdd(name stage.StageName, i *dockerfile.DockerfileStageInstruction[*dockerfile_instruction.Add], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Add {
	return &Add{Base: NewBase(name, i, backend_instruction.NewAdd(*i.Data), dependencies, hasPrevStage, opts)}
}

func (stg *Add) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	args, err := stg.getDependencies(ctx, c, cb, prevImage, prevBuiltImage, buildContextArchive, stg)
	if err != nil {
		return "", err
	}

	args = append(args, "Instruction", stg.instruction.Data.Name())
	args = append(args, "Raw", stg.instruction.Data.Raw)
	args = append(args, append([]string{"Src"}, stg.instruction.Data.Src...)...)
	args = append(args, "Dst", stg.instruction.Data.Dst)
	args = append(args, "Chown", stg.instruction.Data.Chown)
	args = append(args, "Chmod", stg.instruction.Data.Chmod)

	var fileGlobSrc []string
	for _, src := range stg.instruction.Data.Src {
		if !strings.HasPrefix(src, "http://") && !strings.HasPrefix(src, "https://") {
			fileGlobSrc = append(fileGlobSrc, src)
		}
	}

	if len(fileGlobSrc) > 0 {
		if srcChecksum, err := buildContextArchive.CalculateGlobsChecksum(ctx, fileGlobSrc, true); err != nil {
			return "", fmt.Errorf("unable to calculate build context globs checksum: %w", err)
		} else {
			args = append(args, "SrcChecksum", srcChecksum)
		}
	}

	// TODO(staged-dockerfile): support http src and --checksum option: https://docs.docker.com/engine/reference/builder/#verifying-a-remote-file-checksum-add---checksumchecksum-http-src-dest
	// TODO(staged-dockerfile): support git ref: https://docs.docker.com/engine/reference/builder/#adding-a-git-repository-add-git-ref-dir
	// TODO(staged-dockerfile): support --keep-git-dir for git: https://docs.docker.com/engine/reference/builder/#adding-a-git-repository-add-git-ref-dir
	// TODO(staged-dockerfile): support --link

	return util.Sha256Hash(args...), nil
}
