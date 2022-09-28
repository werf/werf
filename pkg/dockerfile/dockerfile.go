package dockerfile

import "context"

func NewDockerfile(dockerfile []byte) *Dockerfile {
	return &Dockerfile{dockerfile: dockerfile}
}

type Dockerfile struct {
	dockerfile []byte
}

func (dockerfile *Dockerfile) GroupStagesByIndependentSets(ctx context.Context) ([][]*DockerfileStage, error) {
	return nil, nil
}
