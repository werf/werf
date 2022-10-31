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

	shlex := shell.NewLex(p.EscapeToken)

	metaArgs, err := processMetaArgs(dockerMetaArgs, opts.BuildArgs, shlex)
	if err != nil {
		return nil, fmt.Errorf("unable to process meta args: %w", err)
	}

	var stages []*dockerfile.DockerfileStage
	for i, dockerStage := range dockerStages {
		name, err := shlex.ProcessWordWithMap(dockerStage.BaseName, metaArgs)
		if err != nil {
			return nil, fmt.Errorf("unable to expand docker stage base image name %q: %w", dockerStage.BaseName, err)
		}
		if name == "" {
			return nil, fmt.Errorf("expanded docker stage base image name %q to empty string: expected image name", dockerStage.BaseName)
		}
		dockerStage.BaseName = name

		// TODO(staged-dockerfile): support meta-args expansion for dockerStage.Platform

		if stage, err := NewDockerfileStageFromBuildkitStage(i, dockerStage, shlex, metaArgs, opts.BuildArgs); err != nil {
			return nil, fmt.Errorf("error converting buildkit stage to dockerfile stage: %w", err)
		} else {
			stages = append(stages, stage)
		}
	}

	dockerfile.SetupDockerfileStagesDependencies(stages)

	d := dockerfile.NewDockerfile(stages, opts)
	for _, stage := range d.Stages {
		stage.Dockerfile = d
	}
	return d, nil
}

func NewDockerfileStageFromBuildkitStage(index int, stage instructions.Stage, shlex *shell.Lex, metaArgs, buildArgs map[string]string) (*dockerfile.DockerfileStage, error) {
	var stageInstructions []dockerfile.DockerfileStageInstructionInterface

	env := map[string]string{}
	opts := dockerfile.DockerfileStageInstructionOptions{Expander: shlex}

	for _, cmd := range stage.Commands {
		var i dockerfile.DockerfileStageInstructionInterface

		switch instrData := cmd.(type) {
		case *instructions.AddCommand:
			if instr, err := createAndExpandInstruction(instrData, env, opts); err != nil {
				return nil, err
			} else {
				i = instr
			}
		case *instructions.ArgCommand:
			if instr, err := createAndExpandInstruction(instrData, env, opts); err != nil {
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
			if instr, err := createAndExpandInstruction(instrData, env, opts); err != nil {
				return nil, err
			} else {
				i = instr
			}
		case *instructions.CopyCommand:
			if instr, err := createAndExpandInstruction(instrData, env, opts); err != nil {
				return nil, err
			} else {
				i = instr
			}
		case *instructions.EntrypointCommand:
			if instr, err := createAndExpandInstruction(instrData, env, opts); err != nil {
				return nil, err
			} else {
				i = instr
			}
		case *instructions.EnvCommand:
			if instr, err := createAndExpandInstruction(instrData, env, opts); err != nil {
				return nil, err
			} else {
				i = instr

				for _, envKV := range instr.Data.Env {
					env[envKV.Key] = envKV.Value
				}
			}
		case *instructions.ExposeCommand:
			if instr, err := createAndExpandInstruction(instrData, env, opts); err != nil {
				return nil, err
			} else {
				i = instr
			}
		case *instructions.HealthCheckCommand:
			if instr, err := createAndExpandInstruction(instrData, env, opts); err != nil {
				return nil, err
			} else {
				i = instr
			}
		case *instructions.LabelCommand:
			if instr, err := createAndExpandInstruction(instrData, env, opts); err != nil {
				return nil, err
			} else {
				i = instr
			}
		case *instructions.MaintainerCommand:
			if instr, err := createAndExpandInstruction(instrData, env, opts); err != nil {
				return nil, err
			} else {
				i = instr
			}
		case *instructions.OnbuildCommand:
			if instr, err := createAndExpandInstruction(instrData, env, opts); err != nil {
				return nil, err
			} else {
				i = instr
			}
		case *instructions.RunCommand:
			if instr, err := createAndExpandInstruction(instrData, env, opts); err != nil {
				return nil, err
			} else {
				i = instr
			}
		case *instructions.ShellCommand:
			if instr, err := createAndExpandInstruction(instrData, env, opts); err != nil {
				return nil, err
			} else {
				i = instr
			}
		case *instructions.StopSignalCommand:
			if instr, err := createAndExpandInstruction(instrData, env, opts); err != nil {
				return nil, err
			} else {
				i = instr
			}
		case *instructions.UserCommand:
			if instr, err := createAndExpandInstruction(instrData, env, opts); err != nil {
				return nil, err
			} else {
				i = instr
			}
		case *instructions.VolumeCommand:
			if instr, err := createAndExpandInstruction(instrData, env, opts); err != nil {
				return nil, err
			} else {
				i = instr
			}
		case *instructions.WorkdirCommand:
			if instr, err := createAndExpandInstruction(instrData, env, opts); err != nil {
				return nil, err
			} else {
				i = instr
			}
		}

		stageInstructions = append(stageInstructions, i)
	}

	return dockerfile.NewDockerfileStage(index, stage.BaseName, stage.Name, stageInstructions, stage.Platform), nil
}

func createAndExpandInstruction[T dockerfile.InstructionDataInterface](data T, env map[string]string, opts dockerfile.DockerfileStageInstructionOptions) (*dockerfile.DockerfileStageInstruction[T], error) {
	i := dockerfile.NewDockerfileStageInstruction(data, opts)
	if err := i.Expand(env); err != nil {
		return nil, fmt.Errorf("unable to expand instruction %q: %w", i.GetInstructionData().Name(), err)
	}
	return i, nil
}

func processMetaArgs(metaArgs []instructions.ArgCommand, buildArgs map[string]string, shlex *shell.Lex) (map[string]string, error) {
	var optMetaArgs []instructions.KeyValuePairOptional

	// TODO(staged-dockerfile): need to support builtin BUILD* and TARGET* args

	// platformOpt := buildPlatformOpt(&opt)
	// optMetaArgs := getPlatformArgs(platformOpt)
	// for i, arg := range optMetaArgs {
	// 	optMetaArgs[i] = setKVValue(arg, opt.BuildArgs)
	// }

	for _, cmd := range metaArgs {
		for _, metaArg := range cmd.Args {
			if metaArg.Value != nil {
				*metaArg.Value, _ = shlex.ProcessWordWithMap(*metaArg.Value, metaArgsToMap(optMetaArgs))
			}
			optMetaArgs = append(optMetaArgs, setKVValue(metaArg, buildArgs))
		}
	}

	return nil, nil
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
