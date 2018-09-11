package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/flant/dapp/pkg/docker_registry"
	"github.com/flant/dapp/pkg/lock"
	"github.com/flant/dapp/pkg/ruby2go"
)

func main() {
	ruby2go.RunCli("docker_registry", func(args map[string]interface{}) (interface{}, error) {
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

		cmd, err := ruby2go.CommandFieldFromArgs(args)
		if err != nil {
			return nil, err
		}

		reference, err := ruby2go.StringOptionFromArgs("reference", args)
		if err != nil {
			return nil, err
		}

		switch cmd {
		case "dimg_tags":
			return docker_registry.DimgTags(reference)
		case "dimgstage_tags":
			return docker_registry.DimgstageTags(reference)
		case "image_id":
			return docker_registry.ImageId(reference)
		case "image_parent_id":
			return docker_registry.ImageParentId(reference)
		case "image_config":
			return docker_registry.ImageConfigFile(reference)
		case "image_delete":
			return nil, docker_registry.ImageDelete(reference)
		case "image_digest":
			return docker_registry.ImageDigest(reference)
		default:
			return nil, fmt.Errorf("command `%s` isn't supported", cmd)
		}

		return nil, nil
	})
}
