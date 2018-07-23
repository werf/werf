package dappdeps

import (
	"fmt"

	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/client"
)

const TOOLCHAIN_VERSION = "0.1.1"

func ToolchainContainer(cli *command.DockerCli, apiClient *client.Client) (string, error) {
	container := &container{
		Name:      fmt.Sprintf("dappdeps_toolchain_%s", TOOLCHAIN_VERSION),
		ImageName: fmt.Sprintf("dappdeps/toolchain:%s", TOOLCHAIN_VERSION),
		Volume:    fmt.Sprintf("/.dapp/deps/toolchain/%s", TOOLCHAIN_VERSION),
	}

	if err := container.CreateIfNotExist(cli, apiClient); err != nil {
		return "", err
	} else {
		return container.Name, nil
	}
}
