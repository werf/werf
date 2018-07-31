package main

import (
	"fmt"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/term"
	"golang.org/x/net/context"

	"github.com/flant/dapp/pkg/image"
	"github.com/flant/dapp/pkg/lock"
	"github.com/flant/dapp/pkg/ruby2go"
)

func main() {
	ruby2go.RunCli("image", func(args map[string]interface{}) (interface{}, error) {
		err := lock.Init()
		if err != nil {
			return nil, err
		}

		cmd, err := ruby2go.CommandFromArgs(args)
		if err != nil {
			return nil, err
		}

		switch cmd {
		case "pull":
			return imageCommand(args, func(cli *command.DockerCli, apiClient *client.Client, stageImage *image.Stage) error {
				if err := stageImage.Pull(cli, apiClient); err != nil {
					return err
				}

				_, err = stageImage.Base.MustGetInspect(apiClient)

				return err
			})
		case "push":
			return imageCommand(args, func(cli *command.DockerCli, apiClient *client.Client, stageImage *image.Stage) error {
				return stageImage.Push(cli, apiClient)
			})
		case "inspect":
			return imageCommand(args, func(cli *command.DockerCli, apiClient *client.Client, stageImage *image.Stage) error {
				_, err := stageImage.GetInspect(apiClient)
				return err
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

				_, err = stageImage.BuildImage.MustGetInspect(apiClient)

				return err
			})
		case "introspect":
			return imageCommand(args, func(cli *command.DockerCli, apiClient *client.Client, stageImage *image.Stage) error {
				return stageImage.Introspect(cli, apiClient)
			})
		case "save_in_cache":
			return imageCommand(args, func(cli *command.DockerCli, apiClient *client.Client, stageImage *image.Stage) error {
				if err := stageImage.SaveInCache(cli, apiClient); err != nil {
					return err
				}

				_, err = stageImage.Base.MustGetInspect(apiClient)

				return err
			})
		case "export", "import", "tag":
			return imageCommand(args, func(cli *command.DockerCli, apiClient *client.Client, stageImage *image.Stage) error {
				name, err := ruby2go.StringOptionFromArgs("name", args)
				if err != nil {
					return err
				}

				switch cmd {
				case "export":
					return stageImage.Export(name, cli, apiClient)
				case "import":
					return stageImage.Import(name, cli, apiClient)
				default: // tag
					return stageImage.Tag(name, cli, apiClient)
				}
			})
		case "save", "load":
			cli, err := dockerClient()
			if err != nil {
				return nil, err
			}

			filePath, err := ruby2go.StringOptionFromArgs("file_path", args)
			if err != nil {
				return nil, err
			}

			if cmd == "save" {
				images, err := arrStringOptionFromArgs("images", args)
				if err != nil {
					return nil, err
				}

				return nil, image.Save(images, filePath, cli)
			} else {
				return nil, image.Load(filePath, cli)
			}
		case "container_run":
			cli, err := dockerClient()
			if err != nil {
				return nil, err
			}

			runArgs, err := arrStringOptionFromArgs("args", args)
			if err != nil {
				return nil, err
			}

			return nil, image.ContainerRun(runArgs, cli)
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

func introspectionOptionFromArgs(args map[string]interface{}) (map[string]bool, error) {
	options, err := ruby2go.OptionsFromArgs(args)
	if err != nil {
		return nil, err
	}

	switch options["introspection"].(type) {
	case map[string]interface{}:
		res, err := mapInterfaceToMapBool(options["introspection"].(map[string]interface{}))
		if err != nil {
			return nil, fmt.Errorf("introspection option field value `%#v` isn't supported: `%s`", options["introspection"], err)
		}
		return res, nil
	default:
		return nil, fmt.Errorf("introspection option field value `%#v` isn't supported", options["introspection"])
	}
}

func mapInterfaceToMapBool(req map[string]interface{}) (map[string]bool, error) {
	res := map[string]bool{}
	for key, val := range req {
		if b, ok := val.(bool); !ok {
			return nil, fmt.Errorf("field value `%#v` isn't supported", val)
		} else {
			res[key] = b
		}
	}
	return res, nil
}

func arrStringOptionFromArgs(optionName string, args map[string]interface{}) ([]string, error) {
	options, err := ruby2go.OptionsFromArgs(args)
	if err != nil {
		return nil, err
	}

	switch options[optionName].(type) {
	case []interface{}:
		res, err := arrInterfaceToArrString(options[optionName].([]interface{}))
		if err != nil {
			return nil, fmt.Errorf("%s option field value `%#v` isn't supported: `%s`", optionName, options[optionName], err)
		}
		return res, nil
	default:
		return nil, fmt.Errorf("option `%s` field value `%#v` isn't supported", optionName, options[optionName])
	}
}

func arrInterfaceToArrString(req []interface{}) ([]string, error) {
	var res []string
	for _, val := range req {
		if str, ok := val.(string); !ok {
			return nil, fmt.Errorf("field value `%#v` isn't supported", val)
		} else {
			res = append(res, str)
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
