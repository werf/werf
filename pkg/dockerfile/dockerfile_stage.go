package dockerfile

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
)

func NewDockerfileStage(index int, baseName, stageName, werfImageName string, instructions []DockerfileStageInstructionInterface, platform string, expanderFactory ExpanderFactory) *DockerfileStage {
	return &DockerfileStage{
		ExpanderFactory: expanderFactory,
		BaseName:        baseName,
		StageName:       stageName,
		WerfImageName:   werfImageName,
		Instructions:    instructions,
		Platform:        platform,
	}
}

type DockerfileStage struct {
	Dockerfile      *Dockerfile
	Dependencies    []*DockerfileStage
	BaseStage       *DockerfileStage
	ExpanderFactory ExpanderFactory

	BaseName      string
	Index         int
	StageName     string
	WerfImageName string
	Platform      string
	Instructions  []DockerfileStageInstructionInterface
}

func (stage *DockerfileStage) AppendDependencyStage(dep *DockerfileStage) {
	for _, d := range stage.Dependencies {
		if d.Index == dep.Index {
			return
		}
	}
	stage.Dependencies = append(stage.Dependencies, dep)
}

func (stage *DockerfileStage) GetWerfImageName() string {
	if stage.HasStageName() {
		return fmt.Sprintf("%s/stage/%s", stage.WerfImageName, stage.StageName)
	} else {
		return fmt.Sprintf("%s/stage%d", stage.WerfImageName, stage.Index)
	}
}

func (stage *DockerfileStage) LogName() string {
	if stage.HasStageName() {
		return stage.StageName
	} else {
		return fmt.Sprintf("<%d>", stage.Index)
	}
}

func (stage *DockerfileStage) HasStageName() bool {
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
		if baseStage, hasKey := stageByName[strings.ToLower(stage.BaseName)]; hasKey {
			stage.BaseStage = baseStage
			stage.Dependencies = append(stage.Dependencies, baseStage)
		}

		for _, instr := range stage.Instructions {
			switch typedInstr := instr.GetInstructionData().(type) {
			case *instructions.CopyCommand:
				if typedInstr.From != "" {
					if dep := findStageByRef(typedInstr.From, stages, stageByName); dep != nil {
						stage.AppendDependencyStage(dep)
						instr.SetDependencyByStageRef(typedInstr.From, dep)
					} else {
						return fmt.Errorf("unable to resolve stage %q instruction %s --from=%q: no such stage", stage.LogName(), instr.GetInstructionData().Name(), typedInstr.From)
					}
				}

			case *instructions.RunCommand:
				mounts := instructions.GetMounts(typedInstr)
				for _, mount := range mounts {
					if mount.From != "" {
						if dep := findStageByRef(mount.From, stages, stageByName); dep != nil {
							stage.AppendDependencyStage(dep)
							instr.SetDependencyByStageRef(mount.From, dep)
						} else {
							return fmt.Errorf("unable to resolve stage %q instruction %s --mount=from=%s: no such stage", stage.LogName(), instr.GetInstructionData().Name(), mount.From)
						}
					}
				}

			}
		}
	}

	return nil
}

// findStageByRef finds stage by stage reference which is stage index or stage name
func findStageByRef(ref string, stages []*DockerfileStage, stageByName map[string]*DockerfileStage) *DockerfileStage {
	if stg, found := stageByName[strings.ToLower(ref)]; found {
		return stg
	} else if ind, err := strconv.Atoi(ref); err == nil && ind >= 0 && ind < len(stages) {
		return stages[ind]
	}
	return nil
}
