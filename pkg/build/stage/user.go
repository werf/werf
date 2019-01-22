package stage

import (
	"github.com/flant/werf/pkg/build/builder"
	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/util"
)

func getBuilder(imageBaseConfig *config.ImageBase, baseStageOptions *NewBaseStageOptions) builder.Builder {
	var b builder.Builder
	if imageBaseConfig.Shell != nil {
		b = builder.NewShellBuilder(imageBaseConfig.Shell)
	} else if imageBaseConfig.Ansible != nil {
		extra := &builder.Extra{ContainerWerfPath: baseStageOptions.ContainerWerfDir, TmpPath: baseStageOptions.ImageTmpDir}
		b = builder.NewAnsibleBuilder(imageBaseConfig.Ansible, extra)
	}

	return b
}

func newUserStage(builder builder.Builder, name StageName, baseStageOptions *NewBaseStageOptions) *UserStage {
	s := &UserStage{}
	s.builder = builder
	s.BaseStage = newBaseStage(name, baseStageOptions)
	return s
}

type UserStage struct {
	*BaseStage

	builder builder.Builder
}

func (s *UserStage) getStageDependenciesChecksum(name StageName) (string, error) {
	var args []string
	for _, gitPath := range s.gitPaths {
		checksum, err := gitPath.StageDependenciesChecksum(name)
		if err != nil {
			return "", err
		}

		args = append(args, checksum)
	}

	return util.Sha256Hash(args...), nil
}
