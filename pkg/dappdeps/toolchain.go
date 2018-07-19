package dappdeps

import (
	"fmt"

	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/client"
)

const TOOLCHAIN_VERSION = "0.1.1"

// TODO lock, logging
func ToolchainContainer(dockerClient *command.DockerCli, dockerApiClient *client.Client) string {
	container := &Container{
		Ref:       fmt.Sprintf("dappdeps_toolchain_%s", TOOLCHAIN_VERSION),
		ImageName: fmt.Sprintf("dappdeps/toolchain:%s", TOOLCHAIN_VERSION),
		Volume:    fmt.Sprintf("/.dapp/deps/toolchain/%s", TOOLCHAIN_VERSION),
	}

	if !container.isExist(dockerApiClient) {
		container.Create(dockerClient)
	}
	return container.Ref
}
