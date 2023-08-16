package docker

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/container"
	"github.com/docker/cli/cli/command/image"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	cli       *command.DockerCli
	apiClient *client.Client
)

func init() {
	if err := initCli(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "init docker cli failed: %s\n", err)
		os.Exit(1)
	}

	if err := initApiClient(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "init docker api client failed: %s\n", err)
		os.Exit(1)
	}
}

func initCli() error {
	cliOpts := []command.DockerCliOption{
		command.WithContentTrust(false),
		command.WithOutputStream(GinkgoWriter),
		command.WithErrorStream(GinkgoWriter),
	}

	logrus.SetOutput(GinkgoWriter)

	newCli, err := command.NewDockerCli(cliOpts...)
	if err != nil {
		return err
	}

	opts := flags.NewClientOptions()
	if err := newCli.Initialize(opts); err != nil {
		return err
	}

	cli = newCli

	return nil
}

func initApiClient() error {
	ctx := context.Background()
	serverVersion, err := cli.Client().ServerVersion(ctx)
	if err != nil {
		return err
	}

	apiClient, err = client.NewClientWithOpts(client.WithVersion(serverVersion.APIVersion))
	if err != nil {
		return err
	}

	return nil
}

func ContainerStopAndRemove(containerName string) {
	Ω(CliStop(containerName)).Should(Succeed(), fmt.Sprintf("docker stop %s", containerName))
	Ω(CliRm(containerName)).Should(Succeed(), fmt.Sprintf("docker rm %s", containerName))
}

func ImageRemoveIfExists(imageName string) {
	if IsImageExist(imageName) {
		Ω(CliRmi(imageName)).Should(Succeed(), "docker rmi")
	}
}

func IsImageExist(imageName string) bool {
	_, err := imageInspect(imageName)
	if err == nil {
		return true
	} else {
		if !strings.HasPrefix(err.Error(), "Error: No such image") && !strings.HasPrefix(err.Error(), "No such image:") {
			Ω(err).ShouldNot(HaveOccurred())
		}

		return false
	}
}

func ImageParent(imageName string) string {
	return ImageInspect(imageName).Parent
}

func ImageID(imageName string) string {
	return ImageInspect(imageName).ID
}

func ImageInspect(imageName string) *types.ImageInspect {
	inspect, err := imageInspect(imageName)
	Ω(err).ShouldNot(HaveOccurred())
	return inspect
}

func ContainerInspect(ref string) types.ContainerJSON {
	ctx := context.Background()
	inspect, err := apiClient.ContainerInspect(ctx, ref)
	Ω(err).ShouldNot(HaveOccurred())
	return inspect
}

func ContainerHostPort(ref, containerPortNumberAndProtocol string) string {
	inspect := ContainerInspect(ref)
	Ω(inspect.NetworkSettings).ShouldNot(BeNil())
	portMap := inspect.NetworkSettings.Ports
	Ω(portMap).ShouldNot(BeEmpty())
	portBindings := portMap[nat.Port(containerPortNumberAndProtocol)]
	Ω(portBindings).ShouldNot(HaveLen(0))
	return portBindings[0].HostPort
}

func CliRun(args ...string) error {
	cmd := container.NewRunCommand(cli)
	return cmdExecute(cmd, args)
}

func CliRm(args ...string) error {
	cmd := container.NewRmCommand(cli)
	return cmdExecute(cmd, args)
}

func CliStop(args ...string) error {
	cmd := container.NewStopCommand(cli)
	return cmdExecute(cmd, args)
}

func CliPull(args ...string) error {
	cmd := image.NewPullCommand(cli)
	return cmdExecute(cmd, args)
}

func CliPush(args ...string) error {
	cmd := image.NewPushCommand(cli)
	return cmdExecute(cmd, args)
}

func CliTag(args ...string) error {
	cmd := image.NewTagCommand(cli)
	return cmdExecute(cmd, args)
}

func CliRmi(args ...string) error {
	cmd := image.NewRemoveCommand(cli)
	return cmdExecute(cmd, args)
}

func cmdExecute(cmd *cobra.Command, args []string) error {
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs(args)
	return cmd.Execute()
}

func Pull(imageName string) error {
tryPull:
	err := CliPull(imageName)
	if err != nil {
		specificErrors := []string{
			"Client.Timeout exceeded while awaiting headers",
			"TLS handshake timeout",
			"i/o timeout",
		}

		for _, specificError := range specificErrors {
			if strings.Contains(err.Error(), specificError) {
				fmt.Fprintf(GinkgoWriter, "Retrying pull in 5 seconds ...")
				time.Sleep(5 * time.Second)
				goto tryPull
			}
		}
	}

	return err
}

func imageInspect(ref string) (*types.ImageInspect, error) {
	ctx := context.Background()
	inspect, _, err := apiClient.ImageInspectWithRaw(ctx, ref)
	if err != nil {
		return nil, err
	}

	return &inspect, nil
}
