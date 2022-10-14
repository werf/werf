package dockerfile

import (
	"context"
)

type DockerfileOptions struct {
	Target    string
	BuildArgs map[string]string
	AddHost   []string
	Network   string
	SSH       string
}

func NewDockerfile(stages []*DockerfileStage, opts DockerfileOptions) *Dockerfile {
	return &Dockerfile{
		DockerfileOptions: opts,
		Stages:            stages,
	}
}

type Dockerfile struct {
	DockerfileOptions

	Stages []*DockerfileStage
}

func (df *Dockerfile) GroupStagesByIndependentSets(ctx context.Context) ([][]*DockerfileStage, error) {
	// FIXME(staged-dockerfile): build real dependencies tree

	// var res [][]*DockerfileStage
	// var curLevel []*DockerfileStage

	// stagesQueue

	// res = append(res, curLevel)

	// for _, stg := range df.Stages {
	// 	stg.Dependencies
	// }

	// var res [][]*DockerfileStage
	// for _, stg := range df.Stages {
	// 	res = append(res, []*DockerfileStage{stg})
	// }
	// return res, nil
	return nil, nil
}
