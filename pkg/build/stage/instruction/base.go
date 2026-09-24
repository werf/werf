package instruction

import (
	"context"
	"fmt"
	"maps"

	"github.com/werf/werf/v3/pkg/build/stage"
	"github.com/werf/werf/v3/pkg/config"
	"github.com/werf/werf/v3/pkg/container_backend"
	"github.com/werf/werf/v3/pkg/dockerfile"
)

type Base[T dockerfile.InstructionDataInterface, BT container_backend.InstructionInterface] struct {
	*stage.BaseStage

	instruction        *dockerfile.DockerfileStageInstruction[T]
	backendInstruction BT
	dependencies       []*config.Dependency
	hasPrevStage       bool

	baseEnv     map[string]string
	expandedEnv map[string]string
}

func NewBase[T dockerfile.InstructionDataInterface, BT container_backend.InstructionInterface](instruction *dockerfile.DockerfileStageInstruction[T], backendInstruction BT, dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Base[T, BT] {
	return &Base[T, BT]{
		BaseStage:          stage.NewBaseStage(stage.StageName(instruction.Data.Name()), opts),
		instruction:        instruction,
		backendInstruction: backendInstruction,
		dependencies:       dependencies,
		hasPrevStage:       hasPrevStage,
	}
}

func (stg *Base[T, BT]) HasPrevStage() bool {
	return stg.hasPrevStage
}

func (stg *Base[T, BT]) IsStapelStage() bool {
	return false
}

func (stg *Base[T, BT]) UsesBuildContext() bool {
	return stg.backendInstruction.UsesBuildContext()
}

func (stg *Base[T, BT]) ExpandDependencies(ctx context.Context, c stage.Conveyor, baseEnv map[string]string) error {
	return stg.doExpandDependencies(ctx, c, baseEnv, stg)
}

func (stg *Base[T, BT]) doExpandDependencies(ctx context.Context, c stage.Conveyor, baseEnv map[string]string, expander InstructionExpander) error {
	if err := stg.expandBaseEnv(baseEnv); err != nil {
		return fmt.Errorf("2nd stage env expansion failed: %w", err)
	}
	if err := expander.ExpandInstruction(c, stg.GetExpandedEnv(c)); err != nil {
		return fmt.Errorf("2nd stage instruction expansion failed: %w", err)
	}
	return nil
}

func (stg *Base[T, BT]) PrepareImage(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevBuiltImage, stageImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) error {
	stageImage.Builder.DockerfileStageBuilder().AppendInstruction(stg.backendInstruction)
	return nil
}

func (stg *Base[T, BT]) expandBaseEnv(baseEnv map[string]string) error {
	env := make(map[string]string)

	// NOTE: default fallback builtin PATH
	env["PATH"] = "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
	for k, v := range baseEnv {
		env[k] = v
	}

	// NOTE: this is 2nd stage expansion (without skipping unset envs)
	instructionEnv, err := stg.instruction.ExpandEnv(env, dockerfile.ExpandOptions{})
	if err != nil {
		return fmt.Errorf("error expanding env: %w", err)
	}
	for k, v := range instructionEnv {
		env[k] = v
	}

	stg.baseEnv = baseEnv
	stg.expandedEnv = env

	return nil
}

func (stg *Base[T, BT]) GetExpandedEnv(c stage.Conveyor) map[string]string {
	return stg.expandedEnvWithDependenciesArgs(stage.ResolveDependenciesArgs(stg.TargetPlatform(), stg.dependencies, c))
}

// GetExpandedEnvForContent is the environment a content digest is derived from:
// the environment the instruction itself declares, with dependency build args
// resolved to stable placeholders. It never reads the environment of the built
// base image, which is only known once the base image exists — the content of a
// base image reaches the digest through the anchor digest of that image instead.
func (stg *Base[T, BT]) GetExpandedEnvForContent() map[string]string {
	env := make(map[string]string)
	maps.Copy(env, stg.instruction.Env)
	maps.Copy(env, stage.ResolveDependenciesArgsForContent(stg.dependencies))
	return env
}

func (stg *Base[T, BT]) contentSourcePaths(c stage.Conveyor, sourcePaths []string) ([]string, error) {
	if stg.instruction.ExpanderFactory == nil {
		return sourcePaths, nil
	}

	dependencyArgs := stage.ResolveDependenciesArgsForContent(stg.dependencies)
	env := make(map[string]string)
	maps.Copy(env, stg.instruction.Env)
	maps.Copy(env, dependencyArgs)
	expander := stg.instruction.ExpanderFactory.GetExpander(dockerfile.ExpandOptions{SkipUnsetEnv: true})

	resolvedSourcePaths := make([]string, 0, len(sourcePaths))
	for _, sourcePath := range sourcePaths {
		resolvedSourcePath, matched, unmatched, err := expander.ProcessWordWithMatches(sourcePath, env)
		if err != nil {
			return nil, fmt.Errorf("expand content source %q: %w", sourcePath, err)
		}
		if len(unmatched) > 0 {
			return []string{"."}, nil
		}

		for key := range matched {
			if _, ok := dependencyArgs[key]; !ok {
				continue
			}

			actualEnv := make(map[string]string)
			maps.Copy(actualEnv, stg.instruction.Env)
			maps.Copy(actualEnv, stage.ResolveDependenciesArgs(stg.TargetPlatform(), stg.dependencies, c))
			resolvedSourcePath, _, unmatched, err = expander.ProcessWordWithMatches(sourcePath, actualEnv)
			if err != nil {
				return nil, fmt.Errorf("expand content source %q: %w", sourcePath, err)
			}
			if len(unmatched) > 0 {
				return []string{"."}, nil
			}
			break
		}

		resolvedSourcePaths = append(resolvedSourcePaths, resolvedSourcePath)
	}

	return resolvedSourcePaths, nil
}

func (stg *Base[T, BT]) expandedEnvWithDependenciesArgs(dependenciesArgs map[string]string) map[string]string {
	env := make(map[string]string)
	maps.Copy(env, stg.expandedEnv)
	maps.Copy(env, dependenciesArgs)
	return env
}

func (stg *Base[T, BT]) ExpandInstruction(_ stage.Conveyor, env map[string]string) error {
	// NOTE: this is 2nd stage expansion (without skipping unset envs)
	return stg.instruction.Expand(env, dockerfile.ExpandOptions{})
}

type InstructionExpander interface {
	ExpandInstruction(c stage.Conveyor, env map[string]string) error
}
