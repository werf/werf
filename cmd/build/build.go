package main

import (
	"errors"
	"fmt"
	"path"

	"k8s.io/kubernetes/pkg/util/file"

	"github.com/flant/dapp/pkg/build"
	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/dapp"
)

type buildRubyCliOptions struct {
	Name     string   `json:"name"`
	BuildDir string   `json:"build_dir"`
	SSHKey   []string `json:"ssh_key"`
}

func runBuild(projectDir string, rubyCliOptions buildRubyCliOptions) error {
	dappfile, err := parseDappfile(projectDir)
	if err != nil {
		return err
	}

	c := build.NewConveyor(dappfile, projectDir, rubyCliOptions.Name, dapp.GetTmpDir())
	if err = c.Build(); err != nil {
		return err
	}

	fmt.Printf("runBuild called: %s %#v\n", projectDir, rubyCliOptions)
	return nil
}

func parseDappfile(projectPath string) ([]*config.Dimg, error) {
	for _, dappfileName := range []string{"dappfile.yml", "dappfile.yaml"} {
		dappfilePath := path.Join(projectPath, dappfileName)
		if exist, err := file.FileExists(dappfilePath); err != nil {
			return nil, err
		} else if exist {
			return config.ParseDimgs(dappfilePath)
		}
	}

	return nil, errors.New("dappfile.y[a]ml not found")
}
