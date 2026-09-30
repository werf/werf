package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/commands"
	"github.com/docker/cli/cli/flags"
	"github.com/moby/moby/api/types/image"
	mobyclient "github.com/moby/moby/client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	cli       *command.DockerCli
	apiClient mobyclient.APIClient
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
	cliOpts := []command.CLIOption{
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
	if _, err := cli.Client().ServerVersion(ctx, mobyclient.ServerVersionOptions{}); err != nil {
		return err
	}

	apiClient = cli.Client()

	return nil
}

func ImageRemoveIfExists(ctx context.Context, imageName string) {
	if IsImageExist(ctx, imageName) {
		Expect(CliRmi(ctx, imageName)).Should(Succeed(), "docker rmi")
	}
}

func IsImageExist(ctx context.Context, imageName string) bool {
	_, err := imageInspect(ctx, imageName)
	if err != nil {
		if cerrdefs.IsNotFound(err) {
			return false
		}

		Expect(err).ShouldNot(HaveOccurred(), err)
	}

	return true
}

func ImageParent(ctx context.Context, imageName string) string {
	var raw bytes.Buffer
	_, err := apiClient.ImageInspect(ctx, imageName, mobyclient.ImageInspectWithRawResponse(&raw))
	Expect(err).ShouldNot(HaveOccurred())
	var inspect struct {
		Parent string
	}
	Expect(json.Unmarshal(raw.Bytes(), &inspect)).To(Succeed())
	return inspect.Parent
}

func ImageID(ctx context.Context, imageName string) string {
	return ImageInspect(ctx, imageName).ID
}

func ImageInspect(ctx context.Context, imageName string) *image.InspectResponse {
	inspect, err := imageInspect(ctx, imageName)
	Expect(err).ShouldNot(HaveOccurred())
	return inspect
}

func lookupCliCommand(c command.Cli, name string) (*cobra.Command, error) {
	root := &cobra.Command{Use: "docker"}
	commands.AddCommands(root, c)
	for _, cmd := range root.Commands() {
		if cmd.Name() == name {
			root.RemoveCommand(cmd)
			return cmd, nil
		}
	}
	return nil, fmt.Errorf("docker CLI command %q not found", name)
}

func CliRm(ctx context.Context, args ...string) error {
	cmd, err := lookupCliCommand(cli, "rm")
	if err != nil {
		return err
	}
	return cmdExecute(ctx, cmd, args)
}

func CliPull(ctx context.Context, args ...string) error {
	cmd, err := lookupCliCommand(cli, "pull")
	if err != nil {
		return err
	}
	return cmdExecute(ctx, cmd, args)
}

func CliPush(ctx context.Context, args ...string) error {
	cmd, err := lookupCliCommand(cli, "push")
	if err != nil {
		return err
	}
	return cmdExecute(ctx, cmd, args)
}

func CliTag(ctx context.Context, args ...string) error {
	cmd, err := lookupCliCommand(cli, "tag")
	if err != nil {
		return err
	}
	return cmdExecute(ctx, cmd, args)
}

func CliRmi(ctx context.Context, args ...string) error {
	cmd, err := lookupCliCommand(cli, "rmi")
	if err != nil {
		return err
	}
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

func imageInspect(ctx context.Context, ref string) (*image.InspectResponse, error) {
	result, err := apiClient.ImageInspect(ctx, ref)
	if err != nil {
		return nil, err
	}

	return &result.InspectResponse, nil
}
