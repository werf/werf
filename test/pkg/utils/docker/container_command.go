package docker

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/alessio/shellescape"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/docker"
	"github.com/werf/werf/v2/pkg/stapel"
	"github.com/werf/werf/v2/test/pkg/utils"
)

func init() {
	var platform string
	for _, envName := range []string{"WERF_PLATFORM", "DOCKER_DEFAULT_PLATFORM"} {
		platform = os.Getenv(envName)
		if platform != "" {
			break
		}
	}

	opts := docker.InitOptions{
		Verbose:         true,
		Debug:           true,
		DefaultPlatform: platform,
	}
	if err := docker.Init(context.Background(), opts); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "init werf docker failed: %s\n", err)
		os.Exit(1)
	}
}

func CheckContainerDirectoryExists(ctx context.Context, werfBinPath, projectPath, containerDirPath string) {
	CheckContainerDirectory(ctx, werfBinPath, projectPath, containerDirPath, true)
}

func CheckContainerDirectoryDoesNotExist(ctx context.Context, werfBinPath, projectPath, containerDirPath string) {
	CheckContainerDirectory(ctx, werfBinPath, projectPath, containerDirPath, false)
}

func CheckContainerDirectory(ctx context.Context, werfBinPath, projectPath, containerDirPath string, exist bool) {
	cmd := CheckContainerFileCommand(containerDirPath, true, exist)
	RunSucceedContainerCommandWithStapel(ctx, werfBinPath, projectPath, []string{}, []string{cmd})
}

func RunSucceedContainerCommandWithStapel(ctx context.Context, werfBinPath, projectPath string, extraDockerOptions, cmds []string) {
	container, err := stapel.GetOrCreateContainer(utils.WithDependencies(ctx))
	Expect(err).ShouldNot(HaveOccurred())

	dockerOptions := []string{
		"--volumes-from",
		container,
		"--rm",
		"--entrypoint=",
	}

	dockerOptions = append(dockerOptions, extraDockerOptions...)

	baseWerfArgs := []string{
		"run", "--docker-options", strings.Join(dockerOptions, " "), "--", stapel.BashBinPath(), "-ec",
	}

	containerCommand := strings.Join(cmds, " && ")
	werfArgs := baseWerfArgs
	werfArgs = append(werfArgs, utils.ShelloutPack(containerCommand))

	_, err = utils.RunCommandWithOptions(ctx, projectPath, werfBinPath, werfArgs, utils.RunCommandOptions{})

	errorDesc := fmt.Sprintf("%[2]s (dir: %[1]s)", projectPath, strings.Join(append(baseWerfArgs, containerCommand), " "))
	Expect(err).ShouldNot(HaveOccurred(), errorDesc)
}

func CheckContainerFileCommand(containerDirPath string, directory, exist bool) string {
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
