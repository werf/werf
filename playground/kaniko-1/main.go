package main

import (
	"fmt"
	"os"

	"github.com/GoogleContainerTools/kaniko/pkg/config"
	"github.com/GoogleContainerTools/kaniko/pkg/executor"
)

func main() {
	opts := &config.KanikoOptions{
		DockerfilePath: os.Args[1],
		SnapshotMode:   "full",
	}

	image, err := executor.DoBuild(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}

	_ = image
}
