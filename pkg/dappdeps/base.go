package dappdeps

import (
	"fmt"

	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/client"
)

const BASE_VERSION = "0.2.3"

// TODO lock, logging
func BaseContainer(dockerClient *command.DockerCli, dockerApiClient *client.Client) string {
	container := &Container{
		Ref:       fmt.Sprintf("dappdeps_base_%s", BASE_VERSION),
		ImageName: fmt.Sprintf("dappdeps/base:%s", BASE_VERSION),
		Volume:    fmt.Sprintf("/.dapp/deps/base/%s", BASE_VERSION),
	}

	if !container.isExist(dockerApiClient) {
		container.Create(dockerClient)
	}
	return container.Ref
}

func BaseBinPath(bin string) string {
	return fmt.Sprintf("/.dapp/deps/base/%s/embedded/bin/%s", BASE_VERSION, bin)
}
