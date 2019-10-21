package docker

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/container"
	"github.com/docker/cli/cli/flags"

	. "github.com/onsi/gomega"

	"github.com/flant/werf/integration/utils"
)

var cli *command.DockerCli

func init() {
	if err := initCli(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "init docker cli failed: %s\n", err)
		os.Exit(1)
	}
}

func initCli() error {
	cliOpts := []command.DockerCliOption{
		command.WithStandardStreams(),
		command.WithContentTrust(false),
	}

	newCli, err := command.NewDockerCli(cliOpts...)
	if err != nil {
		return err
	}

	newCli.Out().SetIsTerminal(terminal.IsTerminal(int(os.Stdout.Fd())))
	newCli.In().SetIsTerminal(terminal.IsTerminal(int(os.Stdin.Fd())))

	opts := flags.NewClientOptions()
	if err := newCli.Initialize(opts); err != nil {
		return err
	}

	cli = newCli

	return nil
}

func ContainerStopAndRemove(containerName string) {
	Ω(CliStop(containerName)).Should(Succeed(), fmt.Sprintf("docker stop %s", containerName))
	Ω(CliRm(containerName)).Should(Succeed(), fmt.Sprintf("docker rm %s", containerName))
}

func LocalDockerRegistryRun() (string, string) {
	containerName := fmt.Sprintf("werf_test_docker_registry-%s", utils.GetRandomString(10))
	imageName := "registry"

	hostPort := strconv.Itoa(utils.GetFreeTCPHostPort())
	dockerCliRunArgs := []string{
		"-d",
		"-p", fmt.Sprintf("%s:5000", hostPort),
		"-e", "REGISTRY_STORAGE_DELETE_ENABLED=true",
		"--name", containerName,
		imageName,
	}
	err := CliRun(dockerCliRunArgs...)
	Ω(err).ShouldNot(HaveOccurred(), "docker run "+strings.Join(dockerCliRunArgs, " "))

	registry := fmt.Sprintf("localhost:%s", hostPort)
	registryWithScheme := fmt.Sprintf("http://%s", registry)

	utils.WaitTillHostReadyToRespond(registryWithScheme, utils.DefaultWaitTillHostReadyToRespondMaxAttempts)

	return registry, containerName
}

func CliRun(args ...string) error {
	cmd := container.NewRunCommand(cli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err != nil {
		return err
	}

	return nil
}

func CliRm(args ...string) error {
	cmd := container.NewRmCommand(cli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err != nil {
		return err
	}

	return nil
}

func CliStop(args ...string) error {
	cmd := container.NewStopCommand(cli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err != nil {
		return err
	}

	return nil
}
