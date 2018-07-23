package main

import (
	"fmt"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/term"
	"golang.org/x/net/context"

	"github.com/flant/dapp/pkg/dappdeps"
	"github.com/flant/dapp/pkg/lock"
	"github.com/flant/dapp/pkg/ruby2go"
)

func main() {
	ruby2go.RunCli("dappdeps", func(args map[string]interface{}) (interface{}, error) {
		err := lock.Init()
		if err != nil {
			return nil, err
		}

		cmd, err := ruby2go.CommandFromArgs(args)
		if err != nil {
			return nil, err
		}

		dappdepsName, err := ruby2go.StringFromMapInterface("dappdeps_name", args)
		if err != nil {
			return nil, err
		}

		switch cmd {
		case "container":
			cli, err := dockerClient()
			if err != nil {
				return nil, err
			}

			apiClient, err := dockerApiClient(cli)
			if err != nil {
				return nil, err
			}

			switch dappdepsName {
			case "base":
				return dappdeps.BaseContainer(cli, apiClient)
			case "gitartifact":
				return dappdeps.GitArtifactContainer(cli, apiClient)
			case "toolchain":
				return dappdeps.ToolchainContainer(cli, apiClient)
			case "ansible":
				return dappdeps.AnsibleContainer(cli, apiClient)
			default:
				return nil, fmt.Errorf("command `%s`: dappdepsName `%s` isn't supported", cmd, dappdepsName)
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
				return nil, fmt.Errorf("command `%s`: dappdepsName `%s` isn't supported", cmd, dappdepsName)
			}
		case "path":
			if dappdepsName == "base" {
				return dappdeps.BasePath(), nil
			} else {
				return nil, fmt.Errorf("command `%s`: dappdepsName `%s` isn't supported", cmd, dappdepsName)
			}
		case "sudo_command":
			ownerOption, err := ruby2go.SafeStringOptionFromArgs("owner", args)
			if err != nil {
				return nil, err
			}

			groupOption, err := ruby2go.SafeStringOptionFromArgs("group", args)
			if err != nil {
				return nil, err
			}

			if dappdepsName == "base" {
				return dappdeps.SudoCommand(ownerOption, groupOption), nil
			} else {
				return nil, fmt.Errorf("command `%s`: dappdepsName `%s` isn't supported", cmd, dappdepsName)
			}
		default:
			return nil, fmt.Errorf("command field value `%s` isn't supported", cmd)
		}

		return nil, nil
	})
}

func dockerClient() (*command.DockerCli, error) {
	stdin, stdout, stderr := term.StdStreams()
	cli := command.NewDockerCli(stdin, stdout, stderr, false)
	opts := flags.NewClientOptions()
	if err := cli.Initialize(opts); err != nil {
		return nil, err
	}

	return cli, nil
}

func dockerApiClient(cli *command.DockerCli) (*client.Client, error) {
	ctx := context.Background()
	serverVersion, err := cli.Client().ServerVersion(ctx)
	apiClient, err := client.NewClientWithOpts(client.WithVersion(serverVersion.APIVersion))
	if err != nil {
		return nil, err
	}
	return apiClient, nil
}
