package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"reflect"

	"github.com/flant/dapp/pkg/build/builder"
	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/config/ruby_marshal_config"
	"github.com/flant/dapp/pkg/docker"
	"github.com/flant/dapp/pkg/image"
	"github.com/flant/dapp/pkg/lock"
	"github.com/flant/dapp/pkg/ruby2go"
)

func main() {
	ruby2go.RunCli("builder", func(args map[string]interface{}) (interface{}, error) {
		err := lock.Init()
		if err != nil {
			return nil, err
		}

		cmd, err := ruby2go.CommandFieldFromArgs(args)
		if err != nil {
			return nil, err
		}

		builderName, err := ruby2go.StringFieldFromMapInterface("builder", args)
		if err != nil {
			return nil, err
		}

		extra, err := extraFromArgs(args)
		if err != nil {
			return nil, err
		}

		switch cmd {
		case "IsBeforeInstallEmpty", "IsInstallEmpty", "IsBeforeSetupEmpty", "IsSetupEmpty", "IsBuildArtifactEmpty":
			switch builderName {
			case "shell":
				shellConfig, err := shellConfigFromArgs(args)
				if err != nil {
					return nil, err
				}
				res := runBuilderMethod(builder.NewShellBuilder(shellConfig), cmd)
				return res[0].Bool(), nil
			case "ansible":
				ansibleConfig, err := ansibleConfigFromArgs(args)
				if err != nil {
					return nil, err
				}
				res := runBuilderMethod(builder.NewAnsibleBuilder(ansibleConfig, extra), cmd)
				return res[0].Bool(), nil
			case "none":
				res := runBuilderMethod(builder.NewNoneBuilder(), cmd)
				return res[0].Bool(), nil
			default:
				return nil, fmt.Errorf("command `%s` isn't supported for builder `%s`", cmd, builderName)
			}
		case "BeforeInstall", "Install", "BeforeSetup", "Setup", "BuildArtifact":
			return image.ImageCommand(args, func(stageImage *image.Stage) error {
				switch builderName {
				case "shell":
					shellConfig, err := shellConfigFromArgs(args)
					if err != nil {
						return err
					}
					res := runBuilderMethod(builder.NewShellBuilder(shellConfig), cmd, stageImage.BuilderContainer())
					err, ok := res[0].Interface().(error)
					if ok {
						return err
					} else {
						return nil
					}
				case "ansible":
					hostDockerConfigDir, err := ruby2go.StringOptionFromArgs("host_docker_config_dir", args)
					if err != nil {
						return err
					}

					if err := docker.Init(hostDockerConfigDir); err != nil {
						return err
					}

					ansibleConfig, err := ansibleConfigFromArgs(args)
					if err != nil {
						return err
					}
					res := runBuilderMethod(builder.NewAnsibleBuilder(ansibleConfig, extra), cmd, stageImage.BuilderContainer())
					err, ok := res[0].Interface().(error)
					if ok {
						return err
					} else {
						return nil
					}
				case "none":
					res := runBuilderMethod(builder.NewNoneBuilder(), cmd, stageImage.BuilderContainer())
					err, ok := res[0].Interface().(error)
					if ok {
						return err
					} else {
						return nil
					}
				default:
					return fmt.Errorf("command `%s` isn't supported for builder `%s`", cmd, builderName)
				}
			})
		case "BeforeInstallChecksum", "InstallChecksum", "BeforeSetupChecksum", "SetupChecksum", "BuildArtifactChecksum":
			switch builderName {
			case "shell":
				shellConfig, err := shellConfigFromArgs(args)
				if err != nil {
					return nil, err
				}
				res := runBuilderMethod(builder.NewShellBuilder(shellConfig), cmd)
				return res[0].String(), nil
			case "ansible":
				ansibleConfig, err := ansibleConfigFromArgs(args)
				if err != nil {
					return nil, err
				}
				res := runBuilderMethod(builder.NewAnsibleBuilder(ansibleConfig, extra), cmd)
				return res[0].String(), nil
			case "none":
				res := runBuilderMethod(builder.NewNoneBuilder(), cmd)
				return res[0].String(), nil
			default:
				return nil, fmt.Errorf("command `%s` isn't supported for builder `%s`", cmd, builderName)
			}
		default:
			return nil, fmt.Errorf("command `%s` isn't supported", cmd)
		}

		return nil, nil
	})
}

func extraFromArgs(args map[string]interface{}) (*builder.Extra, error) {
	value, exist := args["extra"]
	if exist {
		data, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}

		var extra *builder.Extra
		if err := json.Unmarshal(data, &extra); err != nil {
			return nil, err
		}

		return extra, nil
	} else {
		return nil, fmt.Errorf("extra field required")
	}
}

func shellConfigFromArgs(args map[string]interface{}) (config.Shell, error) {
	c, err := ruby2go.StringFieldFromMapInterface("config", args)
	if err != nil {
		return nil, err
	}

	artifact, err := ruby2go.BoolFieldFromMapInterface("artifact", args)
	if err != nil {
		return nil, err
	}

	if artifact {
		var dimgArtifactConfig *ruby_marshal_config.DimgArtifact
		if err := yaml.Unmarshal([]byte(c), &dimgArtifactConfig); err == nil {
			return rubyShellArtifactToShellArtifact(&dimgArtifactConfig.Shell), nil
		} else {
			return nil, err
		}
	} else {
		var dimgConfig *ruby_marshal_config.Dimg
		if err := yaml.Unmarshal([]byte(c), &dimgConfig); err == nil {
			return rubyShellDimgToShellDimg(&dimgConfig.Shell), nil
		} else {
			return nil, err
		}
	}
}

func rubyShellArtifactToShellArtifact(rubyShellArtifactConfig *ruby_marshal_config.ShellArtifact) *config.ShellArtifact {
	shellArtifactConfig := &config.ShellArtifact{}
	shellArtifactConfig.ShellDimg = rubyShellDimgToShellDimg(&rubyShellArtifactConfig.ShellDimg)
	shellArtifactConfig.BuildArtifact = rubyShellArtifactConfig.BuildArtifact.Run
	shellArtifactConfig.BuildArtifactCacheVersion = rubyShellArtifactConfig.BuildArtifact.Version
	return shellArtifactConfig
}

func rubyShellDimgToShellDimg(rubyShellDimgConfig *ruby_marshal_config.ShellDimg) *config.ShellDimg {
	return &config.ShellDimg{rubyShellDimgToShellBase(rubyShellDimgConfig)}
}

func rubyShellDimgToShellBase(rubyShellDimgConfig *ruby_marshal_config.ShellDimg) *config.ShellBase {
	return &config.ShellBase{
		BeforeInstall:             rubyShellDimgConfig.BeforeInstall.Run,
		Install:                   rubyShellDimgConfig.Install.Run,
		BeforeSetup:               rubyShellDimgConfig.BeforeSetup.Run,
		Setup:                     rubyShellDimgConfig.Setup.Run,
		CacheVersion:              rubyShellDimgConfig.Version,
		BeforeInstallCacheVersion: rubyShellDimgConfig.BeforeInstall.Version,
		InstallCacheVersion:       rubyShellDimgConfig.Install.Version,
		BeforeSetupCacheVersion:   rubyShellDimgConfig.BeforeSetup.Version,
		SetupCacheVersion:         rubyShellDimgConfig.Setup.Version,
	}
}

func ansibleConfigFromArgs(args map[string]interface{}) (*config.Ansible, error) {
	c, err := ruby2go.StringFieldFromMapInterface("config", args)
	if err != nil {
		return nil, err
	}

	artifact, err := ruby2go.BoolFieldFromMapInterface("artifact", args)
	if err != nil {
		return nil, err
	}

	if artifact {
		var dimgArtifactConfig *ruby_marshal_config.DimgArtifact
		if err := yaml.Unmarshal([]byte(c), &dimgArtifactConfig); err == nil {
			return rubyAnsibleToAnsible(dimgArtifactConfig.Ansible), nil
		} else {
			return nil, err
		}
	} else {
		var dimgConfig *ruby_marshal_config.Dimg
		if err := yaml.Unmarshal([]byte(c), &dimgConfig); err == nil {
			return rubyAnsibleToAnsible(dimgConfig.Ansible), nil
		} else {
			return nil, err
		}
	}
}

func rubyAnsibleToAnsible(rubyAnsibleConfig ruby_marshal_config.Ansible) *config.Ansible {
	return &config.Ansible{
		BeforeInstall:             rubyAnsibleTasksToAnsibleTasks(rubyAnsibleConfig.BeforeInstall),
		Install:                   rubyAnsibleTasksToAnsibleTasks(rubyAnsibleConfig.Install),
		BeforeSetup:               rubyAnsibleTasksToAnsibleTasks(rubyAnsibleConfig.BeforeSetup),
		Setup:                     rubyAnsibleTasksToAnsibleTasks(rubyAnsibleConfig.Setup),
		BuildArtifact:             rubyAnsibleTasksToAnsibleTasks(rubyAnsibleConfig.BuildArtifact),
		CacheVersion:              rubyAnsibleConfig.Version,
		BeforeInstallCacheVersion: rubyAnsibleConfig.BeforeInstallVersion,
		InstallCacheVersion:       rubyAnsibleConfig.InstallVersion,
		BeforeSetupCacheVersion:   rubyAnsibleConfig.BeforeSetupVersion,
		SetupCacheVersion:         rubyAnsibleConfig.SetupVersion,
		BuildArtifactCacheVersion: rubyAnsibleConfig.BuildArtifactVersion,
		DumpConfigSection:         rubyAnsibleConfig.DumpConfigDoc,
	}
}

func rubyAnsibleTasksToAnsibleTasks(rubyAnsibleTasks []ruby_marshal_config.AnsibleTask) []*config.AnsibleTask {
	var ansibleTasks []*config.AnsibleTask
	for _, rubyAnsibleTask := range rubyAnsibleTasks {
		ansibleTasks = append(ansibleTasks, &config.AnsibleTask{
			Config:            rubyAnsibleTask.Config,
			DumpConfigSection: rubyAnsibleTask.DumpConfigSection,
		})
	}
	return ansibleTasks
}

func runBuilderMethod(b builder.Builder, command string, args ...interface{}) []reflect.Value {
	inputs := make([]reflect.Value, len(args))
	for ind := range args {
		inputs[ind] = reflect.ValueOf(args[ind])
	}
	return reflect.ValueOf(b).MethodByName(command).Call(inputs)
}
