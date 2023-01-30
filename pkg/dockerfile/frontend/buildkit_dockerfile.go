package frontend

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"

	"github.com/werf/werf/pkg/dockerfile"
)

func ParseDockerfileWithBuildkit(dockerfileID string, dockerfileBytes []byte, werfImageName string, opts dockerfile.DockerfileOptions) (*dockerfile.Dockerfile, error) {
	p, err := parser.Parse(bytes.NewReader(dockerfileBytes))
	if err != nil {
		return nil, fmt.Errorf("parsing dockerfile data: %w", err)
	}

	dockerStages, dockerMetaArgsCommands, err := instructions.Parse(p.AST)
	if err != nil {
		return nil, fmt.Errorf("parsing instructions tree: %w", err)
	}

	expanderFactory := NewShlexExpanderFactory(p.EscapeToken)

	metaArgs, err := resolveMetaArgs(dockerMetaArgsCommands, opts.BuildArgs, opts.DependenciesArgsKeys, expanderFactory)
	if err != nil {
		return nil, fmt.Errorf("unable to process meta args: %w", err)
	}

	var stages []*dockerfile.DockerfileStage
	for i, dockerStage := range dockerStages {
		name, err := expanderFactory.GetExpander(dockerfile.ExpandOptions{SkipUnsetEnv: true}).ProcessWordWithMap(dockerStage.BaseName, metaArgs)
		if err != nil {
			return nil, fmt.Errorf("unable to expand docker stage base image name %q: %w", dockerStage.BaseName, err)
		}
		if name == "" {
			return nil, fmt.Errorf("expanded docker stage base image name %q to empty string: expected image name", dockerStage.BaseName)
		}
		dockerStage.BaseName = name

		// TODO(staged-dockerfile): support meta-args expansion for dockerStage.Platform

		// Check stage not already created for another dockerfile?
		// Filepath = dockerfileID — should be generated at this stage, deduplicate stages building at a higher level.
		//   Same underhood stage could be printed several times for each werf-image-target-name.
		// <werf-image>/stage<N> || <werf-image>/stage/<name>

		if stage, err := NewDockerfileStageFromBuildkitStage(i, werfImageName, dockerStage, expanderFactory, metaArgs, opts.BuildArgs, opts.DependenciesArgsKeys); err != nil {
			return nil, fmt.Errorf("error converting buildkit stage to dockerfile stage: %w", err)
		} else {
			stages = append(stages, stage)
		}
	}

	dockerfile.SetupDockerfileStagesDependencies(stages)

	d := dockerfile.NewDockerfile(dockerfileID, stages, opts)
	for _, stage := range d.Stages {
		stage.Dockerfile = d
	}
	return d, nil
}

func NewDockerfileStageFromBuildkitStage(index int, werfImageName string, stage instructions.Stage, expanderFactory *ShlexExpanderFactory, metaArgs, buildArgs map[string]string, dependenciesArgsKeys []string) (*dockerfile.DockerfileStage, error) {
	var stageInstructions []dockerfile.DockerfileStageInstructionInterface

	env := make(map[string]string)

	for _, cmd := range stage.Commands {
		var i dockerfile.DockerfileStageInstructionInterface

		switch instrData := cmd.(type) {
		case *instructions.AddCommand:
			if instr, err := createAndExpandInstruction(instrData, expanderFactory, env); err != nil {
				return nil, err
			} else {
				i = instr
			}
		case *instructions.ArgCommand:
			instrData.Args = removeDependenciesArgs(instrData.Args, dependenciesArgsKeys)

			if instr, err := createAndExpandInstruction(instrData, expanderFactory, env); err != nil {
				return nil, err
			} else {
				i = instr

				for _, arg := range instr.Data.Args {
					if inputValue, hasKey := buildArgs[arg.Key]; hasKey {
						arg.Value = new(string)
						*arg.Value = inputValue
					}

					if arg.Value == nil {
						if mvalue, hasKey := metaArgs[arg.Key]; hasKey {
							arg.Value = new(string)
							*arg.Value = mvalue
						}
					}

					if arg.Value != nil {
						env[arg.Key] = *arg.Value
					}
				}
			}
		case *instructions.CmdCommand:
			if instr, err := createAndExpandInstruction(instrData, expanderFactory, env); err != nil {
				return nil, err
			} else {
				i = instr
			}
		case *instructions.CopyCommand:
			if instr, err := createAndExpandInstruction(instrData, expanderFactory, env); err != nil {
				return nil, err
			} else {
				i = instr
			}
		case *instructions.EntrypointCommand:
			if instr, err := createAndExpandInstruction(instrData, expanderFactory, env); err != nil {
				return nil, err
			} else {
				i = instr
			}
		case *instructions.EnvCommand:
			if instr, err := createAndExpandInstruction(instrData, expanderFactory, env); err != nil {
				return nil, err
			} else {
				i = instr

				for _, envKV := range instr.Data.Env {
					env[envKV.Key] = envKV.Value
				}
			}
		case *instructions.ExposeCommand:
			if instr, err := createAndExpandInstruction(instrData, expanderFactory, env); err != nil {
				return nil, err
			} else {
				i = instr
			}
		case *instructions.HealthCheckCommand:
			if instr, err := createAndExpandInstruction(instrData, expanderFactory, env); err != nil {
				return nil, err
			} else {
				i = instr
			}
		case *instructions.LabelCommand:
			if instr, err := createAndExpandInstruction(instrData, expanderFactory, env); err != nil {
				return nil, err
			} else {
				i = instr
			}
		case *instructions.MaintainerCommand:
			if instr, err := createAndExpandInstruction(instrData, expanderFactory, env); err != nil {
				return nil, err
			} else {
				i = instr
			}
		case *instructions.OnbuildCommand:
			if instr, err := createAndExpandInstruction(instrData, expanderFactory, env); err != nil {
				return nil, err
			} else {
				i = instr
			}
		case *instructions.RunCommand:
			if instr, err := createAndExpandInstruction(instrData, expanderFactory, env); err != nil {
				return nil, err
			} else {
				i = instr
			}
		case *instructions.ShellCommand:
			if instr, err := createAndExpandInstruction(instrData, expanderFactory, env); err != nil {
				return nil, err
			} else {
				i = instr
			}
		case *instructions.StopSignalCommand:
			if instr, err := createAndExpandInstruction(instrData, expanderFactory, env); err != nil {
				return nil, err
			} else {
				i = instr
			}
		case *instructions.UserCommand:
			if instr, err := createAndExpandInstruction(instrData, expanderFactory, env); err != nil {
				return nil, err
			} else {
				i = instr
			}
		case *instructions.VolumeCommand:
			if instr, err := createAndExpandInstruction(instrData, expanderFactory, env); err != nil {
				return nil, err
			} else {
				i = instr
			}
		case *instructions.WorkdirCommand:
			if instr, err := createAndExpandInstruction(instrData, expanderFactory, env); err != nil {
				return nil, err
			} else {
				i = instr
			}
		}

		stageInstructions = append(stageInstructions, i)
	}

	return dockerfile.NewDockerfileStage(index, stage.BaseName, stage.Name, werfImageName, stageInstructions, stage.Platform, expanderFactory), nil
}

func createAndExpandInstruction[T dockerfile.InstructionDataInterface](data T, expanderFactory dockerfile.ExpanderFactory, env map[string]string) (*dockerfile.DockerfileStageInstruction[T], error) {
	opts := dockerfile.DockerfileStageInstructionOptions{
		ExpanderFactory: expanderFactory,
		Env:             make(map[string]string),
	}
	for k, v := range env {
		opts.Env[k] = v
	}

	i := dockerfile.NewDockerfileStageInstruction(data, opts)

	// NOTE: 1st stage expansion (skip unset envs)
	if err := i.Expand(opts.Env, dockerfile.ExpandOptions{SkipUnsetEnv: true}); err != nil {
		return nil, fmt.Errorf("unable to expand instruction %q: %w", i.GetInstructionData().Name(), err)
	}

	return i, nil
}

func isDependencyArg(argKey string, dependenciesArgsKeys []string) bool {
	for _, dep := range dependenciesArgsKeys {
		if dep == argKey {
			return true
		}
	}
	return false
}

func removeDependenciesArgs(args []instructions.KeyValuePairOptional, dependenciesArgsKeys []string) (res []instructions.KeyValuePairOptional) {
	// NOTE: dependencies will be expanded on the second stage expansion
	for _, arg := range args {
		if !isDependencyArg(arg.Key, dependenciesArgsKeys) {
			res = append(res, arg)
		}
	}
	return
}

func resolveMetaArgs(metaArgsCommands []instructions.ArgCommand, buildArgs map[string]string, dependenciesArgsKeys []string, expanderFactory *ShlexExpanderFactory) (map[string]string, error) {
	var optMetaArgs []instructions.KeyValuePairOptional

	// TODO(staged-dockerfile): need to support builtin BUILD* and TARGET* args

	// platformOpt := buildPlatformOpt(&opt)
	// optMetaArgs := getPlatformArgs(platformOpt)
	// for i, arg := range optMetaArgs {
	// 	optMetaArgs[i] = setKVValue(arg, opt.BuildArgs)
	// }

	for _, cmd := range metaArgsCommands {
		for _, metaArg := range cmd.Args {
			if isDependencyArg(metaArg.Key, dependenciesArgsKeys) {
				continue
			}

			if metaArg.Value != nil {
				*metaArg.Value, _ = expanderFactory.GetExpander(dockerfile.ExpandOptions{SkipUnsetEnv: true}).ProcessWordWithMap(*metaArg.Value, metaArgsToMap(optMetaArgs))
			}
			optMetaArgs = append(optMetaArgs, setKVValue(metaArg, buildArgs))
		}
	}

	return metaArgsToMap(optMetaArgs), nil
}

func metaArgsToMap(metaArgs []instructions.KeyValuePairOptional) map[string]string {
	m := map[string]string{}
	for _, arg := range metaArgs {
		m[arg.Key] = arg.ValueString()
	}
	return m
}

func setKVValue(kvpo instructions.KeyValuePairOptional, values map[string]string) instructions.KeyValuePairOptional {
	if v, ok := values[kvpo.Key]; ok {
		kvpo.Value = &v
	}
	return kvpo
}

// TODO(staged-dockerfile)
//
// func getPlatformArgs(po *platformOpt) []instructions.KeyValuePairOptional {
// 	bp := po.buildPlatforms[0]
// 	tp := po.targetPlatform
// 	m := map[string]string{
// 		"BUILDPLATFORM":  platforms.Format(bp),
// 		"BUILDOS":        bp.OS,
// 		"BUILDARCH":      bp.Architecture,
// 		"BUILDVARIANT":   bp.Variant,
// 		"TARGETPLATFORM": platforms.Format(tp),
// 		"TARGETOS":       tp.OS,
// 		"TARGETARCH":     tp.Architecture,
// 		"TARGETVARIANT":  tp.Variant,
// 	}
// 	opts := make([]instructions.KeyValuePairOptional, 0, len(m))
// 	for k, v := range m {
// 		s := v
// 		opts = append(opts, instructions.KeyValuePairOptional{Key: k, Value: &s})
// 	}
// 	return opts
// }

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
