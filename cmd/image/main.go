package main

import (
	"fmt"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/term"
	"golang.org/x/net/context"

	"github.com/flant/dapp/pkg/image"
	"github.com/flant/dapp/pkg/ruby2go"
)

func main() {
	ruby2go.RunCli("image", func(args map[string]interface{}) (map[string]interface{}, error) {
		cmd, err := commandFromArgs(args)
		if err != nil {
			return nil, err
		}

		switch cmd {
		case "inspect":
			return imageCommand(args, func(dockerClient *command.DockerCli, dockerApiClient *client.Client, stageImage *image.Stage) error {
				return stageImage.ResetInspect(dockerApiClient)
			})
		case "build":
			return imageCommand(args, func(dockerClient *command.DockerCli, dockerApiClient *client.Client, stageImage *image.Stage) error {
				if err := stageImage.Build(dockerClient, dockerApiClient); err != nil {
					return err
				}
				if err := stageImage.ResetBuiltInspect(dockerApiClient); err != nil {
					return err
				}
				return nil
			})
		default:
			return nil, fmt.Errorf("command field `%s` isn't supported", cmd)
		}

		return nil, nil
	})
}

func imageCommand(args map[string]interface{}, command func(dockerClient *command.DockerCli, dockerApiClient *client.Client, stageImage *image.Stage) error) (map[string]interface{}, error) {
	dockerClient, err := dockerClient()
	if err != nil {
		return nil, err
	}

	dockerApiClient, err := dockerApiClient(dockerClient)
	if err != nil {
		return nil, err
	}

	resultMap, err := ruby2go.CommandWithImage(args, func(stageImage *image.Stage) error {
		return command(dockerClient, dockerApiClient, stageImage)
	})
	if err != nil {
		return nil, err
	}

	return resultMap, nil
}

func commandFromArgs(args map[string]interface{}) (string, error) {
	switch args["command"].(type) {
	case string:
		return args["command"].(string), nil
	default:
		return "", fmt.Errorf("command field value `%v` isn't supported", args["command"])
	}
}

func dockerClient() (*command.DockerCli, error) {
	stdin, stdout, stderr := term.StdStreams()
	dockerClient := command.NewDockerCli(stdin, stdout, stderr, true)
	opts := flags.NewClientOptions()
	if err := dockerClient.Initialize(opts); err != nil {
		return nil, err
	}

	return dockerClient, nil
}

func dockerApiClient(dockerClient *command.DockerCli) (*client.Client, error) {
	ctx := context.Background()
	serverVersion, err := dockerClient.Client().ServerVersion(ctx)
	dockerApiClient, err := client.NewClientWithOpts(client.WithVersion(serverVersion.APIVersion))
	if err != nil {
		return nil, err
	}
	return dockerApiClient, nil
}
