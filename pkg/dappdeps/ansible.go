package dappdeps

import (
	"fmt"

	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/client"
)

const ANSIBLE_VERSION = "2.4.4.0-10"

func AnsibleContainer(cli *command.DockerCli, apiClient *client.Client) (string, error) {
	container := &container{
		Name:      fmt.Sprintf("dappdeps_ansible_%s", ANSIBLE_VERSION),
		ImageName: fmt.Sprintf("dappdeps/ansible:%s", ANSIBLE_VERSION),
		Volume:    fmt.Sprintf("/.dapp/deps/ansible/%s", ANSIBLE_VERSION),
	}

	if err := container.CreateIfNotExist(cli, apiClient); err != nil {
		return "", err
	} else {
		return container.Name, nil
	}
}

func AnsibleBinPath(bin string) string {
	return fmt.Sprintf("/.dapp/deps/ansible/%s/embedded/bin/%s", ANSIBLE_VERSION, bin)
}

func AnsiblePlaybookSoloCmd() string {
	return fmt.Sprintf("%s -c local", AnsibleBinPath("ansible_playbook"))
}
