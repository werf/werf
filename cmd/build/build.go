package main

import (
	"errors"
	"fmt"
	"os"
	"path"

	"k8s.io/kubernetes/pkg/util/file"

	"github.com/flant/dapp/pkg/build"
	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/dapp"
	"github.com/flant/dapp/pkg/docker"
	"github.com/flant/dapp/pkg/lock"
	"github.com/flant/dapp/pkg/ssh_agent"
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

	if err := lock.Init(); err != nil {
		return err
	}

	if err := ssh_agent.Init(rubyCliOptions.SSHKey); err != nil {
		return fmt.Errorf("cannot initialize ssh-agent: %s", err)
	}

	if err := docker.Init(hostDockerConfigDir()); err != nil {
		return err
	}

	c := build.NewConveyor(dappfile, projectDir, path.Base(projectDir), dapp.GetTmpDir()) // TODO project name
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

/**
TODO
if options_with_docker_credentials? && !options[:repo].nil?
	host_docker_tmp_config_dir
end
*/
func hostDockerConfigDir() string {
	dappDockerConfigEnv := os.Getenv("DAPP_DOCKER_CONFIG")

	if dappDockerConfigEnv != "" {
		return dappDockerConfigEnv
	} else {
		return path.Join(os.Getenv("HOME"), ".docker")
	}
}
