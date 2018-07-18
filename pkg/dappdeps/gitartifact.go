package dappdeps

import (
	"fmt"

	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/client"
)

const GITARTIFACT_VERSION = "0.2.1"

// TODO lock, logging
func GitArtifactContainer(dockerClient *command.DockerCli, dockerApiClient *client.Client) string {
	container := &Container{
		Ref:       fmt.Sprintf("dappdeps_gitartifact_%s", GITARTIFACT_VERSION),
		ImageName: fmt.Sprintf("dappdeps/gitartifact:%s", GITARTIFACT_VERSION),
		Volume:    fmt.Sprintf("/.dapp/deps/gitartifact/%s", GITARTIFACT_VERSION),
	}

	if !container.isExist(dockerApiClient) {
		container.Create(dockerClient)
	}
	return container.Ref
}

func GitBin() string {
	return fmt.Sprintf("/.dapp/deps/gitartifact/%s/bin/git", GITARTIFACT_VERSION)
}
