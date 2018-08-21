package main

import (
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"

	"github.com/flant/dapp/pkg/docker"
	"github.com/flant/dapp/pkg/image"
	"github.com/flant/dapp/pkg/lock"
	"github.com/flant/dapp/pkg/ruby2go"
	"github.com/flant/dapp/pkg/util"
)

func main() {
	ruby2go.RunCli("image", func(args map[string]interface{}) (interface{}, error) {
		err := lock.Init()
		if err != nil {
			return nil, err
		}

		if err := docker.Init(); err != nil {
			return nil, err
		}

		cmd, err := ruby2go.CommandFieldFromArgs(args)
		if err != nil {
			return nil, err
		}

		switch cmd {
		case "pull":
			return image.ImageCommand(args, func(stageImage *image.Stage) error {
				if err := stageImage.Pull(); err != nil {
					return err
				}

				_, err = stageImage.Base.MustGetInspect()

				return err
			})
		case "push":
			return image.ImageCommand(args, func(stageImage *image.Stage) error {
				return stageImage.Push()
			})
		case "inspect":
			return image.ImageCommand(args, func(stageImage *image.Stage) error {
				_, err := stageImage.GetInspect()
				return err
			})
		case "build":
			return image.ImageCommand(args, func(stageImage *image.Stage) error {
				introspection, err := introspectionOptionFromArgs(args)
				if err != nil {
					return err
				}

				buildOptions := &image.StageBuildOptions{
					IntrospectBeforeError: introspection["before"],
					IntrospectAfterError:  introspection["after"],
				}

				if err := stageImage.Build(buildOptions); err != nil {
					return err
				}

				_, err = stageImage.BuildImage.MustGetInspect()

				return err
			})
		case "introspect":
			return image.ImageCommand(args, func(stageImage *image.Stage) error {
				return stageImage.Introspect()
			})
		case "save_in_cache":
			return image.ImageCommand(args, func(stageImage *image.Stage) error {
				if err := stageImage.SaveInCache(); err != nil {
					return err
				}

				_, err = stageImage.Base.MustGetInspect()

				return err
			})
		case "untag":
			return image.ImageCommand(args, func(stageImage *image.Stage) error {
				return stageImage.Untag()
			})
		case "export", "import", "tag":
			return image.ImageCommand(args, func(stageImage *image.Stage) error {
				name, err := ruby2go.StringOptionFromArgs("name", args)
				if err != nil {
					return err
				}

				switch cmd {
				case "export":
					return stageImage.Export(name)
				case "import":
					return stageImage.Import(name)
				default: // tag
					return stageImage.Tag(name)
				}
			})
		case "save", "load", "container_run":
			switch cmd {
			case "save", "load":
				filePath, err := ruby2go.StringOptionFromArgs("file_path", args)
				if err != nil {
					return nil, err
				}

				if cmd == "save" {
					images, err := stringArrayOptionFromArgs("images", args)
					if err != nil {
						return nil, err
					}

					var args []string
					args = append(args, []string{"-o", filePath}...)
					args = append(args, images...)

					return nil, docker.CliSave(args...)
				} else {
					return nil, docker.CliLoad("-i", filePath)
				}
			case "container_run":
				runArgs, err := stringArrayOptionFromArgs("args", args)
				if err != nil {
					return nil, err
				}

				return nil, docker.CliRun(runArgs...)
			default:
				return nil, fmt.Errorf("command `%s` isn't supported", cmd)
			}
		case "containers", "images":
			filtersOption, err := filtersOptionFromArgs(args)
			if err != nil {
				return nil, err
			}

			filterSet := filters.NewArgs()
			for _, set := range filtersOption {
				for k, v := range set {
					filterSet.Add(k, v)
				}
			}

			switch cmd {
			case "images":
				return docker.Images(types.ImageListOptions{Filters: filterSet})
			default: // "containers"
				options := types.ContainerListOptions{}
				options.All = true
				options.Quiet = true
				options.Filters = filterSet
				return docker.Containers(options)
			}
		case "rm", "rmi":
			ids, err := stringArrayOptionFromArgs("ids", args)
			if err != nil {
				return nil, err
			}

			force, err := forceOptionFromArgs(args)
			if err != nil {
				return nil, err
			}

			var args []string
			if force {
				args = append(args, fmt.Sprintf("-f"))
			}
			args = append(args, ids...)

			switch cmd {
			case "rm":
				return nil, docker.CliRm(args...)
			default: // "rmi"
				return nil, docker.CliRmi(args...)
			}
		default:
			return nil, fmt.Errorf("command `%s` isn't supported", cmd)
		}

		return nil, nil
	})
}

func introspectionOptionFromArgs(args map[string]interface{}) (map[string]bool, error) {
	options, err := ruby2go.OptionsFieldFromArgs(args)
	if err != nil {
		return nil, err
	}

	switch options["introspection"].(type) {
	case map[string]interface{}:
		res, err := mapInterfaceToMapBool(options["introspection"].(map[string]interface{}))
		if err != nil {
			return nil, fmt.Errorf("introspection option field value `%#v` can't be casting into map[string]bool: `%s`", options["introspection"], err)
		}
		return res, nil
	default:
		return nil, fmt.Errorf("introspection option field value `%#v` can't be casting into map[string]bool", options["introspection"])
	}
}

func mapInterfaceToMapBool(req map[string]interface{}) (map[string]bool, error) {
	res := map[string]bool{}
	for key, val := range req {
		if b, ok := val.(bool); !ok {
			return nil, fmt.Errorf("key `%s` value `%#v` can't be casting into bool", key, val)
		} else {
			res[key] = b
		}
	}
	return res, nil
}

func stringArrayOptionFromArgs(optionName string, args map[string]interface{}) ([]string, error) {
	options, err := ruby2go.OptionsFieldFromArgs(args)
	if err != nil {
		return nil, err
	}

	switch options[optionName].(type) {
	case []interface{}:
		res, err := util.InterfaceArrayToStringArray(options[optionName].([]interface{}))
		if err != nil {
			return nil, fmt.Errorf("%s option field value `%#v` can't be casting into []string: `%s`", optionName, options[optionName], err)
		}
		return res, nil
	default:
		return nil, fmt.Errorf("option `%s` field value `%#v` can't be casting into []string", optionName, options[optionName])
	}
}

func filtersOptionFromArgs(args map[string]interface{}) ([]map[string]string, error) {
	options, err := ruby2go.OptionsFieldFromArgs(args)
	if err != nil {
		return nil, err
	}

	filtersOption := options["filters"]

	switch filtersOption.(type) {
	case []interface{}:
		var res []map[string]string
		for _, elm := range filtersOption.([]interface{}) {
			mapStringInterface, err := util.InterfaceToMapStringInterface(elm)
			if err != nil {
				return nil, err
			}

			mapStringString, err := mapStringInterfaceToMapStringString(mapStringInterface)
			if err != nil {
				return nil, fmt.Errorf("option `filters` field value `%#v` can't be casting into map[string]string: %s", filtersOption, err)
			}

			res = append(res, mapStringString)
		}
		return res, nil
	default:
		return nil, fmt.Errorf("option `filters` field value `%#v` can't be casting into map[string]string", filtersOption)
	}
}

func mapStringInterfaceToMapStringString(req map[string]interface{}) (map[string]string, error) {
	res := map[string]string{}
	for key, val := range req {
		if b, ok := val.(string); !ok {
			return nil, fmt.Errorf("key `%s` value `%#v` can't be casting into string", key, val)
		} else {
			res[key] = b
		}
	}
	return res, nil
}

func forceOptionFromArgs(args map[string]interface{}) (bool, error) {
	options, err := ruby2go.OptionsFieldFromArgs(args)
	if err != nil {
		return false, err
	}

	return ruby2go.BoolFieldFromMapInterface("force", options)
}
