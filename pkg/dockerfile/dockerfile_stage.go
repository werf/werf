package dockerfile

func NewDockerfileStage() *DockerfileStage {
	return &DockerfileStage{}
}

type DockerfileStage struct {
	Dockerfile         *Dockerfile
	DependenciesStages []*DockerfileStage
}
