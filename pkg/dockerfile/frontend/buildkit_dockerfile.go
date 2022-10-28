package frontend

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/moby/buildkit/frontend/dockerfile/shell"

	"github.com/werf/werf/pkg/dockerfile"
)

func ParseDockerfileWithBuildkit(dockerfileBytes []byte, opts dockerfile.DockerfileOptions) (*dockerfile.Dockerfile, error) {
	p, err := parser.Parse(bytes.NewReader(dockerfileBytes))
	if err != nil {
		return nil, fmt.Errorf("parsing dockerfile data: %w", err)
	}

	dockerStages, dockerMetaArgs, err := instructions.Parse(p.AST)
	if err != nil {
		return nil, fmt.Errorf("parsing instructions tree: %w", err)
	}

	dockerTargetIndex, err := GetDockerTargetStageIndex(dockerStages, opts.Target)
	if err != nil {
		return nil, fmt.Errorf("determine target stage: %w", err)
	}

	shlex := shell.NewLex(p.EscapeToken)

	var stages []*dockerfile.DockerfileStage
	for i, dockerStage := range dockerStages {
		if stage, err := NewDockerfileStageFromBuildkitStage(i, dockerStage, shlex); err != nil {
			return nil, fmt.Errorf("error converting buildkit stage to dockerfile stage: %w", err)
		} else {
			stages = append(stages, stage)
		}
	}

	// TODO(staged-dockerfile): convert meta-args and initialize into Dockerfile obj
	_ = dockerMetaArgs
	_ = dockerTargetIndex

	dockerfile.SetupDockerfileStagesDependencies(stages)

	d := dockerfile.NewDockerfile(stages, opts)
	for _, stage := range d.Stages {
		stage.Dockerfile = d
	}
	return d, nil
}

func NewDockerfileStageFromBuildkitStage(index int, stage instructions.Stage, shlex *shell.Lex) (*dockerfile.DockerfileStage, error) {
	var stageInstructions []dockerfile.DockerfileStageInstructionInterface

	for _, cmd := range stage.Commands {
		if expandable, ok := cmd.(instructions.SupportsSingleWordExpansion); ok {
			if err := expandable.Expand(func(word string) (string, error) {
				// FIXME(ilya-lesikov): add envs/buildargs here
				return shlex.ProcessWord(word, []string{})
			}); err != nil {
				return nil, fmt.Errorf("error expanding command %q: %w", cmd.Name(), err)
			}
		}

		var i dockerfile.DockerfileStageInstructionInterface
		switch typedCmd := cmd.(type) {
		case *instructions.AddCommand:
			i = dockerfile.NewDockerfileStageInstruction(typedCmd)
		case *instructions.ArgCommand:
			i = dockerfile.NewDockerfileStageInstruction(typedCmd)
		case *instructions.CmdCommand:
			i = dockerfile.NewDockerfileStageInstruction(typedCmd)
		case *instructions.CopyCommand:
			i = dockerfile.NewDockerfileStageInstruction(typedCmd)
		case *instructions.EntrypointCommand:
			i = dockerfile.NewDockerfileStageInstruction(typedCmd)
		case *instructions.EnvCommand:
			i = dockerfile.NewDockerfileStageInstruction(typedCmd)
		case *instructions.ExposeCommand:
			i = dockerfile.NewDockerfileStageInstruction(typedCmd)
		case *instructions.HealthCheckCommand:
			i = dockerfile.NewDockerfileStageInstruction(typedCmd)
		case *instructions.LabelCommand:
			i = dockerfile.NewDockerfileStageInstruction(typedCmd)
		case *instructions.MaintainerCommand:
			i = dockerfile.NewDockerfileStageInstruction(typedCmd)
		case *instructions.OnbuildCommand:
			i = dockerfile.NewDockerfileStageInstruction(typedCmd)
		case *instructions.RunCommand:
			i = dockerfile.NewDockerfileStageInstruction(typedCmd)
		case *instructions.ShellCommand:
			i = dockerfile.NewDockerfileStageInstruction(typedCmd)
		case *instructions.StopSignalCommand:
			i = dockerfile.NewDockerfileStageInstruction(typedCmd)
		case *instructions.UserCommand:
			i = dockerfile.NewDockerfileStageInstruction(typedCmd)
		case *instructions.VolumeCommand:
			i = dockerfile.NewDockerfileStageInstruction(typedCmd)
		case *instructions.WorkdirCommand:
			i = dockerfile.NewDockerfileStageInstruction(typedCmd)
		}
		stageInstructions = append(stageInstructions, i)
	}

	return dockerfile.NewDockerfileStage(index, stage.BaseName, stage.Name, stageInstructions, stage.Platform), nil
}

func GetDockerStagesNameToIndexMap(stages []instructions.Stage) map[string]int {
	nameToIndex := make(map[string]int)
	for i, s := range stages {
		name := strings.ToLower(s.Name)
		if name != strconv.Itoa(i) {
			nameToIndex[name] = i
		}
	}
	return nameToIndex
}

func ResolveDockerStagesFromValue(stages []instructions.Stage) {
	nameToIndex := GetDockerStagesNameToIndexMap(stages)

	for _, s := range stages {
		for _, cmd := range s.Commands {
			switch typedCmd := cmd.(type) {
			case *instructions.CopyCommand:
				if typedCmd.From != "" {
					from := strings.ToLower(typedCmd.From)
					if val, ok := nameToIndex[from]; ok {
						typedCmd.From = strconv.Itoa(val)
					}
				}

			case *instructions.RunCommand:
				for _, mount := range instructions.GetMounts(typedCmd) {
					if mount.From != "" {
						from := strings.ToLower(mount.From)
						if val, ok := nameToIndex[from]; ok {
							mount.From = strconv.Itoa(val)
						}
					}
				}
			}
		}
	}
}

func GetDockerTargetStageIndex(dockerStages []instructions.Stage, dockerTargetStage string) (int, error) {
	if dockerTargetStage == "" {
		return len(dockerStages) - 1, nil
	}

	for i, s := range dockerStages {
		if s.Name == dockerTargetStage {
			return i, nil
		}
	}

	return -1, fmt.Errorf("%s is not a valid target build stage", dockerTargetStage)
}
