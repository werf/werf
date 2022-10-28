package dockerfile

import (
	"fmt"
)

type InstructionDataInterface interface {
	Name() string
}

type DockerfileStageInstructionInterface interface {
	SetDependencyByStageRef(ref string, dep *DockerfileStage)
	GetDependencyByStageRef(ref string) *DockerfileStage
	GetDependenciesByStageRef() map[string]*DockerfileStage
	GetInstructionData() InstructionDataInterface
	// TODO(staged-dockerfile): something like Expand(args, envs map[string]string)
}

type DockerfileStageInstruction[T InstructionDataInterface] struct {
	Data                   T
	DependenciesByStageRef map[string]*DockerfileStage
}

func NewDockerfileStageInstruction[T InstructionDataInterface](data T) *DockerfileStageInstruction[T] {
	return &DockerfileStageInstruction[T]{
		Data:                   data,
		DependenciesByStageRef: make(map[string]*DockerfileStage),
	}
}

func (i *DockerfileStageInstruction[T]) SetDependencyByStageRef(ref string, dep *DockerfileStage) {
	if d, hasDep := i.DependenciesByStageRef[ref]; hasDep {
		if d.Index != dep.Index {
			panic(fmt.Sprintf("already set instruction dependency %q to stage %d named %q, cannot replace with stage %d named %q, please report a bug", ref, d.Index, d.StageName, dep.Index, dep.StageName))
		}
		return
	}
	i.DependenciesByStageRef[ref] = dep
}

func (i *DockerfileStageInstruction[T]) GetDependencyByStageRef(ref string) *DockerfileStage {
	return i.DependenciesByStageRef[ref]
}

func (i *DockerfileStageInstruction[T]) GetDependenciesByStageRef() map[string]*DockerfileStage {
	return i.DependenciesByStageRef
}

func (i *DockerfileStageInstruction[T]) GetInstructionData() InstructionDataInterface {
	return i.Data
}
