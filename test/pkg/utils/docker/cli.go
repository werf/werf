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

func ContainerStopAndRemove(ctx context.Context, containerName string) {
	Expect(CliStop(ctx, containerName)).Should(Succeed(), fmt.Sprintf("docker stop %s", containerName))
	Expect(CliRm(ctx, containerName)).Should(Succeed(), fmt.Sprintf("docker rm %s", containerName))
}

func ImageRemoveIfExists(ctx context.Context, imageName string) {
	if IsImageExist(ctx, imageName) {
		Expect(CliRmi(ctx, imageName)).Should(Succeed(), "docker rmi")
	}
}

func IsImageExist(ctx context.Context, imageName string) bool {
	_, err := imageInspect(ctx, imageName)
	if err != nil {
		if client.IsErrNotFound(err) {
			return false
		}

		Expect(err).ShouldNot(HaveOccurred(), err)
	}

	return true
}

func ImageParent(ctx context.Context, imageName string) string {
	return ImageInspect(ctx, imageName).Parent
}

func ImageID(ctx context.Context, imageName string) string {
	return ImageInspect(ctx, imageName).ID
}

func ImageInspect(ctx context.Context, imageName string) *types.ImageInspect {
	inspect, err := imageInspect(ctx, imageName)
	Expect(err).ShouldNot(HaveOccurred())
	return inspect
}

func ContainerInspect(ctx context.Context, ref string) types.ContainerJSON {
	inspect, err := apiClient.ContainerInspect(ctx, ref)
	Expect(err).ShouldNot(HaveOccurred())
	return inspect
}

func ContainerHostPort(ctx context.Context, ref, containerPortNumberAndProtocol string) string {
	inspect := ContainerInspect(ctx, ref)
	Expect(inspect.NetworkSettings).ShouldNot(BeNil())
	portMap := inspect.NetworkSettings.Ports
	Expect(portMap).ShouldNot(BeEmpty())
	portBindings := portMap[nat.Port(containerPortNumberAndProtocol)]
	Expect(portBindings).ShouldNot(HaveLen(0))
	return portBindings[0].HostPort
}

func CliRun(ctx context.Context, args ...string) error {
	cmd := container.NewRunCommand(cli)
	return cmdExecute(ctx, cmd, args)
}

func CliRm(ctx context.Context, args ...string) error {
	cmd := container.NewRmCommand(cli)
	return cmdExecute(ctx, cmd, args)
}

func CliStop(ctx context.Context, args ...string) error {
	cmd := container.NewStopCommand(cli)
	return cmdExecute(ctx, cmd, args)
}

func CliPull(ctx context.Context, args ...string) error {
	cmd := image.NewPullCommand(cli)
	return cmdExecute(ctx, cmd, args)
}

func CliPush(ctx context.Context, args ...string) error {
	cmd := image.NewPushCommand(cli)
	return cmdExecute(ctx, cmd, args)
}

func CliTag(ctx context.Context, args ...string) error {
	cmd := image.NewTagCommand(cli)
	return cmdExecute(ctx, cmd, args)
}

func CliRmi(ctx context.Context, args ...string) error {
	cmd := image.NewRemoveCommand(cli)
	return cmdExecute(ctx, cmd, args)
}

func cmdExecute(ctx context.Context, cmd *cobra.Command, args []string) error {
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs(args)
	cmd.SetContext(ctx)
	return cmd.Execute()
}

func Pull(ctx context.Context, imageName string) error {
tryPull:
	err := CliPull(ctx, imageName)
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

func imageInspect(ctx context.Context, ref string) (*types.ImageInspect, error) {
	inspect, _, err := apiClient.ImageInspectWithRaw(ctx, ref)
	if err != nil {
		return nil, err
	}

	return &inspect, nil
}
