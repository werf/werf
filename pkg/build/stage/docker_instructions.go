package stage

import (
	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/util"
)

func GenerateDockerInstructionsStage(dimgConfig *config.Dimg) Interface {
	if dimgConfig.Docker != nil {
		return newDockerInstructionsStage(dimgConfig.Docker)
	}

	return nil
}

func newDockerInstructionsStage(instructions *config.Docker) *DockerInstructionsStage {
	s := &DockerInstructionsStage{}
	s.instructions = instructions
	s.BaseStage = newBaseStage()
	return s
}

type DockerInstructionsStage struct {
	*BaseStage

	instructions *config.Docker
}

func (s *DockerInstructionsStage) Name() string {
	return "dockerInstructions"
}

func (s *DockerInstructionsStage) GetDependencies() string {
	var args []string

	args = append(args, s.instructions.Volume...)
	args = append(args, s.instructions.Expose...)

	for k, v := range s.instructions.Env {
		args = append(args, k, v)
	}

	for k, v := range s.instructions.Label {
		args = append(args, k, v)
	}

	args = append(args, s.instructions.Cmd...)
	args = append(args, s.instructions.Onbuild...)
	args = append(args, s.instructions.Entrypoint...)
	args = append(args, s.instructions.Workdir)
	args = append(args, s.instructions.User)

	return util.Sha256Hash(args...)
}
