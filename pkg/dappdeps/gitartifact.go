package dappdeps

import (
	"fmt"
)

const GITARTIFACT_VERSION = "0.2.1"

func GitArtifactContainer() (string, error) {
	container := &container{
		Name:      fmt.Sprintf("dappdeps_gitartifact_%s", GITARTIFACT_VERSION),
		ImageName: fmt.Sprintf("dappdeps/gitartifact:%s", GITARTIFACT_VERSION),
		Volume:    fmt.Sprintf("/.dapp/deps/gitartifact/%s", GITARTIFACT_VERSION),
	}

	if err := container.CreateIfNotExist(); err != nil {
		return "", err
	} else {
		return container.Name, nil
	}
}

func GitBin() string {
	return fmt.Sprintf("/.dapp/deps/gitartifact/%s/bin/git", GITARTIFACT_VERSION)
}
