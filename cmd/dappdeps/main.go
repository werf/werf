package main

import (
	"fmt"

	"github.com/flant/dapp/pkg/dappdeps"
	"github.com/flant/dapp/pkg/docker"
	"github.com/flant/dapp/pkg/lock"
	"github.com/flant/dapp/pkg/ruby2go"
)

func main() {
	ruby2go.RunCli("dappdeps", func(args map[string]interface{}) (interface{}, error) {
		err := lock.Init()
		if err != nil {
			return nil, err
		}

		cmd, err := ruby2go.CommandFieldFromArgs(args)
		if err != nil {
			return nil, err
		}

		dappdepsName, err := ruby2go.StringFieldFromMapInterface("dappdeps", args)
		if err != nil {
			return nil, err
		}

		switch cmd {
		case "container":
			if err := docker.Init(); err != nil {
				return nil, err
			}

			switch dappdepsName {
			case "base":
				return dappdeps.BaseContainer()
			case "gitartifact":
				return dappdeps.GitArtifactContainer()
			case "toolchain":
				return dappdeps.ToolchainContainer()
			case "ansible":
				return dappdeps.AnsibleContainer()
			default:
				return nil, fmt.Errorf("command `%s` isn't supported for dappdeps `%s`", cmd, dappdepsName)
			}
		case "bin":
			switch dappdepsName {
			case "base", "ansible":
				binOption, err := ruby2go.StringOptionFromArgs("bin", args)
				if err != nil {
					return nil, err
				}

				if dappdepsName == "base" {
					return dappdeps.BaseBinPath(binOption), nil
				} else {
					return dappdeps.AnsibleBinPath(binOption), nil
				}
			case "gitartifact":
				return dappdeps.GitBin(), nil
			default:
				return nil, fmt.Errorf("command `%s` isn't supported for dappdeps `%s`", cmd, dappdepsName)
			}
		case "path":
			if dappdepsName == "base" {
				return dappdeps.BasePath(), nil
			} else {
				return nil, fmt.Errorf("command `%s` isn't supported for dappdeps `%s`", cmd, dappdepsName)
			}
		case "sudo_command":
			ownerOption, err := ownerOrGroupOptionFromArgs("owner", args)
			if err != nil {
				return nil, err
			}

			groupOption, err := ownerOrGroupOptionFromArgs("group", args)
			if err != nil {
				return nil, err
			}

			if dappdepsName == "base" {
				return dappdeps.SudoCommand(ownerOption, groupOption), nil
			} else {
				return nil, fmt.Errorf("command `%s` isn't supported for dappdeps `%s`", cmd, dappdepsName)
			}
		default:
			return nil, fmt.Errorf("command `%s` isn't supported", cmd)
		}

		return nil, nil
	})
}

func ownerOrGroupOptionFromArgs(option string, args map[string]interface{}) (string, error) {
	options, err := ruby2go.OptionsFieldFromArgs(args)
	if err != nil {
		return "", err
	}

	value, ok := options[option]
	if !ok || value == nil {
		return "", nil
	}

	return fmt.Sprintf("%v", value), nil
}
