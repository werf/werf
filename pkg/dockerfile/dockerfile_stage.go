package dockerfile

import (
	"fmt"
	"strconv"
	"strings"

	dockerfile_instruction "github.com/werf/werf/pkg/dockerfile/instruction"
)

func NewDockerfileStage(index int, baseName, stageName string, instructions []InstructionInterface, platform string) *DockerfileStage {
	return &DockerfileStage{BaseName: baseName, StageName: stageName, Instructions: instructions, Platform: platform}
}

type DockerfileStage struct {
	Dockerfile   *Dockerfile
	Dependencies []*DockerfileStage

	BaseName     string
	Index        int
	StageName    string
	Platform     string
	Instructions []InstructionInterface
}

func (stage DockerfileStage) LogName() string {
	if stage.HasStageName() {
		return stage.StageName
	} else {
		return fmt.Sprintf("<%d>", stage.Index)
	}
}

func (stage DockerfileStage) HasStageName() bool {
	return stage.StageName != ""
}

func SetupDockerfileStagesDependencies(stages []*DockerfileStage) error {
	stageByName := make(map[string]*DockerfileStage)
	for _, stage := range stages {
		if stage.HasStageName() {
			stageByName[strings.ToLower(stage.StageName)] = stage
		}
	}

	for _, stage := range stages {
		// Base image dependency
		if dependency, hasKey := stageByName[strings.ToLower(stage.BaseName)]; hasKey {
			stage.Dependencies = append(stage.Dependencies, dependency)
		}

		for _, instr := range stage.Instructions {
			switch typedInstr := instr.(type) {
			case *dockerfile_instruction.Copy:
				if dep := findStageByNameOrIndex(typedInstr.From, stages, stageByName); dep != nil {
					stage.Dependencies = append(stage.Dependencies, dep)
				} else {
					return fmt.Errorf("unable to resolve stage %q instruction %s --from=%q: no such stage", stage.LogName(), instr.Name(), typedInstr.From)
				}

			case *dockerfile_instruction.Run:
				for _, mount := range typedInstr.Mounts {
					if mount.From != "" {
						if dep := findStageByNameOrIndex(mount.From, stages, stageByName); dep != nil {
							stage.Dependencies = append(stage.Dependencies, dep)
						} else {
							return fmt.Errorf("unable to resolve stage %q instruction %s --mount=from=%s: no such stage", stage.LogName(), instr.Name(), mount.From)
						}
					}
				}
			}
		}
	}

	return nil
}

func findStageByNameOrIndex(ref string, stages []*DockerfileStage, stageByName map[string]*DockerfileStage) *DockerfileStage {
	if stg, found := stageByName[strings.ToLower(ref)]; found {
		return stg
	} else if ind, err := strconv.Atoi(ref); err == nil && ind >= 0 && ind < len(stages) {
		return stages[ind]
	}
	return nil
}
