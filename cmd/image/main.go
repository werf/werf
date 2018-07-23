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
		case "pull":
			return imageCommand(args, func(cli *command.DockerCli, apiClient *client.Client, stageImage *image.Stage) error {
				return stageImage.Pull(cli, apiClient)
			})
		case "push":
			return imageCommand(args, func(cli *command.DockerCli, apiClient *client.Client, stageImage *image.Stage) error {
				return stageImage.Push(cli, apiClient)
			})
		case "inspect":
			return imageCommand(args, func(cli *command.DockerCli, apiClient *client.Client, stageImage *image.Stage) error {
				stageImage.GetInspect(apiClient)
				return nil
			})
		case "build":
			return imageCommand(args, func(cli *command.DockerCli, apiClient *client.Client, stageImage *image.Stage) error {
				introspection, err := introspectionOptionFromArgs(args)
				if err != nil {
					return err
				}

				buildOptions := &image.StageBuildOptions{
					IntrospectBeforeError: introspection["before"],
					IntrospectAfterError:  introspection["after"],
				}

				if err := stageImage.Build(buildOptions, cli, apiClient); err != nil {
					return err
				}

				stageImage.BuildImage.MustGetInspect(apiClient)

				return nil
			})
		case "introspect":
			return imageCommand(args, func(cli *command.DockerCli, apiClient *client.Client, stageImage *image.Stage) error {
				return stageImage.Introspect(cli, apiClient)
			})
		case "save_in_cache":
			return imageCommand(args, func(cli *command.DockerCli, apiClient *client.Client, stageImage *image.Stage) error {
				return stageImage.SaveInCache(cli, apiClient)
			})
		case "untag":
			return imageCommand(args, func(cli *command.DockerCli, apiClient *client.Client, stageImage *image.Stage) error {
				return stageImage.Untag(cli, apiClient)
			})
		default:
			return nil, fmt.Errorf("command field `%s` isn't supported", cmd)
		}

		return nil, nil
	})
}

func imageCommand(args map[string]interface{}, command func(cli *command.DockerCli, apiClient *client.Client, stageImage *image.Stage) error) (map[string]interface{}, error) {
	cli, err := dockerClient()
	if err != nil {
		return nil, err
	}

	apiClient, err := dockerApiClient(cli)
	if err != nil {
		return nil, err
	}

	resultMap, err := image.CommandWithImage(args, func(stageImage *image.Stage) error {
		return command(cli, apiClient, stageImage)
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

func introspectionOptionFromArgs(args map[string]interface{}) (map[string]bool, error) {
	options, err := ruby2go.OptionsFromArgs(args)
	if err != nil {
		return nil, err
	}

	switch options["introspection"].(type) {
	case map[string]interface{}:
		res, err := mapInterfaceToMapBool(options["introspection"].(map[string]interface{}))
		if err != nil {
			return nil, fmt.Errorf("introspection option field value `%v` isn't supported: `%s`", options["introspection"], err)
		}
		return res, nil
	default:
		return nil, fmt.Errorf("introspection option field value `%v` isn't supported", options["introspection"])
	}
}

func mapInterfaceToMapBool(req map[string]interface{}) (map[string]bool, error) {
	res := map[string]bool{}
	for key, val := range req {
		if b, ok := val.(bool); !ok {
			return nil, fmt.Errorf("field value `%v` isn't supported", val)
		} else {
			res[key] = b
		}
	}
	return res, nil
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
