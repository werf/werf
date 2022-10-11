package dockerfile

import (
	"bytes"
	"context"
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

func ParseDockerfile(dockerfile []byte, opts DockerfileOptions) (*Dockerfile, error) {
	p, err := parser.Parse(bytes.NewReader(dockerfile))
	if err != nil {
		return nil, fmt.Errorf("parsing dockerfile data: %w", err)
	}

	dockerStages, dockerMetaArgs, err := instructions.Parse(p.AST)
	if err != nil {
		return nil, fmt.Errorf("parsing instructions tree: %w", err)
	}

	// FIXME(staged-dockerfile): is this needed?
	ResolveDockerStagesFromValue(dockerStages)

	dockerTargetIndex, err := GetDockerTargetStageIndex(dockerStages, opts.Target)
	if err != nil {
		return nil, fmt.Errorf("determine target stage: %w", err)
	}

	return newDockerfile(dockerStages, dockerMetaArgs, dockerTargetIndex, opts), nil
}

type DockerfileOptions struct {
	Target    string
	BuildArgs map[string]string
	AddHost   []string
	Network   string
	SSH       string
}

func newDockerfile(dockerStages []instructions.Stage, dockerMetaArgs []instructions.ArgCommand, dockerTargetStageIndex int, opts DockerfileOptions) *Dockerfile {
	return &Dockerfile{
		DockerfileOptions: opts,

		dockerStages:           dockerStages,
		dockerMetaArgs:         dockerMetaArgs,
		dockerTargetStageIndex: dockerTargetStageIndex,
		nameToIndex:            GetDockerStagesNameToIndexMap(dockerStages),
	}
}

type Dockerfile struct {
	DockerfileOptions

	dockerStages           []instructions.Stage
	dockerMetaArgs         []instructions.ArgCommand
	dockerTargetStageIndex int
	nameToIndex            map[string]string
}

func (dockerfile *Dockerfile) GroupStagesByIndependentSets(ctx context.Context) ([][]*DockerfileStage, error) {
	return nil, nil
}
