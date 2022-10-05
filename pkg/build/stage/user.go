package stage

import (
	"context"
	"os"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/build/builder"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/util"
)

func getBuilder(imageBaseConfig *config.StapelImageBase, baseStageOptions *BaseStageOptions) builder.Builder {
	var b builder.Builder
	extra := &builder.Extra{ContainerWerfPath: baseStageOptions.ContainerWerfDir, TmpPath: baseStageOptions.ImageTmpDir}
	if imageBaseConfig.Shell != nil {
		b = builder.NewShellBuilder(imageBaseConfig.Shell, extra)
	} else if imageBaseConfig.Ansible != nil {
		b = builder.NewAnsibleBuilder(imageBaseConfig.Ansible, extra)
	}

	return b
}

func newUserStage(builder builder.Builder, name StageName, baseStageOptions *BaseStageOptions) *UserStage {
	s := &UserStage{}
	s.builder = builder
	s.BaseStage = NewBaseStage(name, baseStageOptions)
	return s
}

type UserStage struct {
	*BaseStage

	builder builder.Builder
}

func (s *UserStage) getStageDependenciesChecksum(ctx context.Context, c Conveyor, name StageName) (string, error) {
	var args []string
	for _, gitMapping := range s.gitMappings {
		checksum, err := gitMapping.StageDependenciesChecksum(ctx, c, name)
		if err != nil {
			return "", err
		}

		if debugUserStageChecksum() {
			logboek.Context(ctx).Debug().LogFHighlight(
				"DEBUG: %s stage git mapping %s checksum %v\n",
				name, gitMapping.Name, checksum,
			)
		}

		args = append(args, checksum)
	}

	return util.Sha256Hash(args...), nil
}

func debugUserStageChecksum() bool {
	return os.Getenv("WERF_DEBUG_USER_STAGE_CHECKSUM") == "1"
}
