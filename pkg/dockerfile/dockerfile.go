package dockerfile

import (
	"fmt"
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

func (df *Dockerfile) GetTargetStage() (*DockerfileStage, error) {
	if df.Target == "" {
		return df.Stages[len(df.Stages)-1], nil
	}

	for _, s := range df.Stages {
		if s.StageName == df.Target {
			return s, nil
		}
	}

	return nil, fmt.Errorf("%s is not a valid target dockerfile stage", df.Target)
}

func (df *Dockerfile) FindStage(name string) *DockerfileStage {
	for _, s := range df.Stages {
		if s.StageName == name {
			return s
		}
	}
	return nil
}
