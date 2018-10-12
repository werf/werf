package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/flant/dapp/pkg/cleanup"
	"github.com/flant/dapp/pkg/docker"
	"github.com/flant/dapp/pkg/lock"
	"github.com/flant/dapp/pkg/ruby2go"
)

func main() {
	ruby2go.RunCli("cleanup", func(args map[string]interface{}) (interface{}, error) {
		err := lock.Init()
		if err != nil {
			return nil, err
		}

		hostDockerConfigDir, err := ruby2go.StringOptionFromArgs("host_docker_config_dir", args)
		if err != nil {
			return nil, err
		}

		os.Setenv("DOCKER_CONFIG", hostDockerConfigDir)
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)

		if err := docker.Init(hostDockerConfigDir); err != nil {
			return nil, err
		}

		cmd, err := ruby2go.CommandFieldFromArgs(args)
		if err != nil {
			return nil, err
		}

		switch cmd {
		case "reset":
			rawOptions, err := ruby2go.StringFieldFromMapInterface("command_options", args)
			if err != nil {
				return nil, err
			}

			options := &cleanup.ResetOptions{}
			err = json.Unmarshal([]byte(rawOptions), options)
			if err != nil {
				return nil, err
			}

			return nil, cleanup.Reset(*options)
		case "flush":
			rawOptions, err := ruby2go.StringFieldFromMapInterface("command_options", args)
			if err != nil {
				return nil, err
			}

			options := &cleanup.FlushOptions{}
			err = json.Unmarshal([]byte(rawOptions), options)
			if err != nil {
				return nil, err
			}

			return nil, cleanup.Flush(*options)
		case "sync":
			rawOptions, err := ruby2go.StringFieldFromMapInterface("command_options", args)
			if err != nil {
				return nil, err
			}

			options := &cleanup.SyncOptions{}
			err = json.Unmarshal([]byte(rawOptions), options)
			if err != nil {
				return nil, err
			}

			return nil, cleanup.Sync(*options)
		default:
			return nil, fmt.Errorf("command `%s` isn't supported", cmd)
		}

		return nil, nil
	})
}
