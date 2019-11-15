package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/alessio/shellescape"

	. "github.com/onsi/gomega"

	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/stapel"
)

func init() {
	if err := docker.Init(""); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "init werf docker failed: %s\n", err)
		os.Exit(1)
	}
}

func CheckContainerDirectoryExists(werfBinPath, projectPath, containerDirPath string) {
	CheckContainerDirectory(werfBinPath, projectPath, containerDirPath, true)
}

func CheckContainerDirectoryDoesNotExist(werfBinPath, projectPath, containerDirPath string) {
	CheckContainerDirectory(werfBinPath, projectPath, containerDirPath, false)
}

func CheckContainerDirectory(werfBinPath, projectPath, containerDirPath string, exist bool) {
	cmd := CheckContainerFileCommand(containerDirPath, true, exist)
	RunSucceedContainerCommandWithStapel(werfBinPath, projectPath, []string{}, []string{cmd})
}

func RunSucceedContainerCommandWithStapel(werfBinPath string, projectPath string, extraDockerOptions []string, cmd []string) {
	container, err := stapel.GetOrCreateContainer()
	Ω(err).ShouldNot(HaveOccurred())

	dockerOptions := []string{
		"--volumes-from",
		container,
		"--rm",
	}

	dockerOptions = append(dockerOptions, extraDockerOptions...)

	werfArgs := []string{
		"run", "--docker-options", strings.Join(dockerOptions, " "), "--", stapel.BashBinPath(), "-ec",
	}
	werfArgs = append(werfArgs, strings.Join(cmd, " && "))

	RunSucceedCommand(
		projectPath,
		werfBinPath,
		werfArgs...,
	)
}

func CheckContainerFileCommand(containerDirPath string, directory bool, exist bool) string {
	var cmd string
	var flag string

	if directory {
		flag = "-d"
	} else {
		flag = "-f"
	}

	if exist {
		cmd = fmt.Sprintf("test %s %s", flag, shellescape.Quote(containerDirPath))
	} else {
		cmd = fmt.Sprintf("! test %s %s", flag, shellescape.Quote(containerDirPath))
	}

	return cmd
}
