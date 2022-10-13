package dockerfile

func NewDockerfileStage(dockerfile *Dockerfile, instructions []InstructionInterface) *DockerfileStage {
	return &DockerfileStage{Dockerfile: dockerfile, Instructions: instructions}
}

type DockerfileStage struct {
	Dockerfile   *Dockerfile
	Instructions []InstructionInterface
}
