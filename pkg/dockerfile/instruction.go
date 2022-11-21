package dockerfile

import (
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
)

type InstructionDataInterface interface {
	Name() string
}

type ExpandOptions struct {
	SkipUnsetEnv bool
}

type DockerfileStageInstructionInterface interface {
	SetDependencyByStageRef(ref string, dep *DockerfileStage)
	GetDependencyByStageRef(ref string) *DockerfileStage
	GetDependenciesByStageRef() map[string]*DockerfileStage
	GetInstructionData() InstructionDataInterface
	Expand(env map[string]string, opts ExpandOptions) error
}

type ExpanderFactory interface {
	GetExpander(opts ExpandOptions) Expander
}

type Expander interface {
	ProcessWordWithMap(word string, env map[string]string) (string, error)
	ProcessWordsWithMap(word string, env map[string]string) ([]string, error)
}

type DockerfileStageInstructionOptions struct {
	ExpanderFactory ExpanderFactory
}

type DockerfileStageInstruction[T InstructionDataInterface] struct {
	Data                   T
	DependenciesByStageRef map[string]*DockerfileStage
	ExpanderFactory        ExpanderFactory
}

func NewDockerfileStageInstruction[T InstructionDataInterface](data T, opts DockerfileStageInstructionOptions) *DockerfileStageInstruction[T] {
	return &DockerfileStageInstruction[T]{
		Data:                   data,
		DependenciesByStageRef: make(map[string]*DockerfileStage),
		ExpanderFactory:        opts.ExpanderFactory,
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

func (i *DockerfileStageInstruction[T]) Expand(env map[string]string, opts ExpandOptions) error {
	if i.ExpanderFactory == nil {
		return nil
	}
	expander := i.ExpanderFactory.GetExpander(opts)

	switch instr := any(i.Data).(type) {
	case instructions.SupportsSingleWordExpansion:
		return instr.Expand(func(word string) (string, error) {
			return expander.ProcessWordWithMap(word, env)
		})

	case *instructions.ExposeCommand:
		// NOTE: ExposeCommand does not implement Expander interface, but actually needs expansion

		ports := []string{}
		for _, p := range instr.Ports {
			ps, err := expander.ProcessWordsWithMap(p, env)
			if err != nil {
				return fmt.Errorf("unable to expand expose instruction port %q: %w", p, err)
			}
			ports = append(ports, ps...)
		}
		instr.Ports = ports

	case *instructions.RunCommand:
		var newCmdLine []string
		for _, line := range instr.CmdLine {
			exline, err := expander.ProcessWordWithMap(line, env)
			if err != nil {
				return fmt.Errorf("unable to expand cmd line %q: %w", line, err)
			}
			newCmdLine = append(newCmdLine, exline)
		}
		instr.ShellDependantCmdLine = instructions.ShellDependantCmdLine{
			CmdLine:      newCmdLine,
			PrependShell: instr.PrependShell,
		}
	}

	return nil
}
