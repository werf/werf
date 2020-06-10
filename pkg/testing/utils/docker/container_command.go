package docker

import (
	"fmt"
	"os"
	"strings"

	"github.com/alessio/shellescape"

	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/stapel"
	"github.com/werf/werf/pkg/testing/utils"
)

func init() {
	if err := docker.Init("", true, true); err != nil {
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

func RunSucceedContainerCommandWithStapel(werfBinPath string, projectPath string, extraDockerOptions []string, cmds []string) {
	container, err := stapel.GetOrCreateContainer()
	Ω(err).ShouldNot(HaveOccurred())

	dockerOptions := []string{
		"--volumes-from",
		container,
		"--rm",
	}

	dockerOptions = append(dockerOptions, extraDockerOptions...)

	baseWerfArgs := []string{
		"run", "--docker-options", strings.Join(dockerOptions, " "), "--", stapel.BashBinPath(), "-ec",
	}
	containerCommand := strings.Join(cmds, " && ")
	werfArgs := append(baseWerfArgs, utils.ShelloutPack(containerCommand))

	_, err = utils.RunCommandWithOptions(
		projectPath,
		werfBinPath,
		werfArgs,
		utils.RunCommandOptions{},
	)

	errorDesc := fmt.Sprintf("%[2]s (dir: %[1]s)", projectPath, strings.Join(append(baseWerfArgs, containerCommand), " "))
	Ω(err).ShouldNot(HaveOccurred(), errorDesc)
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
