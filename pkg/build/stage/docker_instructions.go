package stage

import (
	"sort"

	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/util"
)

func GenerateDockerInstructionsStage(imageConfig *config.StapelImage, baseStageOptions *NewBaseStageOptions) *DockerInstructionsStage {
	if imageConfig.Docker != nil {
		return newDockerInstructionsStage(imageConfig.Docker, baseStageOptions)
	}

	return nil
}

func newDockerInstructionsStage(instructions *config.Docker, baseStageOptions *NewBaseStageOptions) *DockerInstructionsStage {
	s := &DockerInstructionsStage{}
	s.instructions = instructions
	s.BaseStage = newBaseStage(DockerInstructions, baseStageOptions)
	return s
}

type DockerInstructionsStage struct {
	*BaseStage

	instructions *config.Docker
}

func (s *DockerInstructionsStage) GetDependencies(_ Conveyor, _, _ container_runtime.ImageInterface) (string, error) {
	var args []string

	args = append(args, s.instructions.Volume...)
	args = append(args, s.instructions.Expose...)
	args = append(args, mapToSortedArgs(s.instructions.Env)...)
	args = append(args, mapToSortedArgs(s.instructions.Label)...)
	args = append(args, s.instructions.Cmd)
	args = append(args, s.instructions.Entrypoint)
	args = append(args, s.instructions.Workdir)
	args = append(args, s.instructions.User)
	args = append(args, s.instructions.HealthCheck)

	return util.Sha256Hash(args...), nil
}

func mapToSortedArgs(h map[string]string) (result []string) {
	keys := make([]string, 0, len(h))
	for key := range h {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		result = append(result, key, h[key])
	}

	return
}

func (s *DockerInstructionsStage) PrepareImage(c Conveyor, prevBuiltImage, image container_runtime.ImageInterface) error {
	if err := s.BaseStage.PrepareImage(c, prevBuiltImage, image); err != nil {
		return err
	}

	imageCommitChangeOptions := image.Container().CommitChangeOptions()
	imageCommitChangeOptions.AddVolume(s.instructions.Volume...)
	imageCommitChangeOptions.AddExpose(s.instructions.Expose...)
	imageCommitChangeOptions.AddEnv(s.instructions.Env)
	imageCommitChangeOptions.AddLabel(s.instructions.Label)
	imageCommitChangeOptions.AddCmd(s.instructions.Cmd)
	imageCommitChangeOptions.AddEntrypoint(s.instructions.Entrypoint)
	imageCommitChangeOptions.AddUser(s.instructions.User)
	imageCommitChangeOptions.AddWorkdir(s.instructions.Workdir)
	imageCommitChangeOptions.AddHealthCheck(s.instructions.HealthCheck)

	return nil
}
