package stage

import (
	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/image"
	"github.com/flant/dapp/pkg/util"
)

func GenerateDockerInstructionsStage(dimgConfig *config.Dimg, baseStageOptions *NewBaseStageOptions) *DockerInstructionsStage {
	if dimgConfig.Docker != nil {
		return newDockerInstructionsStage(dimgConfig.Docker, baseStageOptions)
	}

	return nil
}

func newDockerInstructionsStage(instructions *config.Docker, baseStageOptions *NewBaseStageOptions) *DockerInstructionsStage {
	s := &DockerInstructionsStage{}
	s.instructions = instructions
	s.BaseStage = newBaseStage(baseStageOptions)
	return s
}

type DockerInstructionsStage struct {
	*BaseStage

	instructions *config.Docker
}

func (s *DockerInstructionsStage) Name() StageName {
	return DockerInstructions
}

func (s *DockerInstructionsStage) GetDependencies(_ Conveyor, _ image.Image) (string, error) {
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

	return util.Sha256Hash(args...), nil
}

func (s *DockerInstructionsStage) PrepareImage(_, image image.Image) error {
	imageCommitChangeOptions := image.Container().CommitChangeOptions()

	imageCommitChangeOptions.AddVolume(s.instructions.Volume...)
	imageCommitChangeOptions.AddExpose(s.instructions.Expose...)
	imageCommitChangeOptions.AddEnv(s.instructions.Env)
	imageCommitChangeOptions.AddCmd(s.instructions.Cmd...)
	imageCommitChangeOptions.AddOnbuild(s.instructions.Onbuild...)
	imageCommitChangeOptions.AddEntrypoint(s.instructions.Entrypoint...)
	imageCommitChangeOptions.AddUser(s.instructions.User)
	imageCommitChangeOptions.AddWorkdir(s.instructions.Workdir)

	return nil
}
