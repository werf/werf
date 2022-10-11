package dockerfile

import "github.com/moby/buildkit/frontend/dockerfile/instructions"

func NewDockerfileStage(dockerStage instructions.Stage) *DockerfileStage {
	return &DockerfileStage{dockerStage: dockerStage}
}

type DockerfileStage struct {
	Dockerfile         *Dockerfile
	DependenciesStages []*DockerfileStage

	dockerStage instructions.Stage
}

func (stage *DockerfileStage) GetInstructions() []InstructionInterface {
	// TODO(staged-dockerfile)
	// for _, cmd := range stage.dockerStage.Commands {
	// 	switch typedCmd := cmd.(type) {
	// 	case *instructions.ArgCommand:
	// 	}
	// }

	return nil
}
