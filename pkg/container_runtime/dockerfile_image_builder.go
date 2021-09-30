package container_runtime

import (
	"context"
	"fmt"
	"os"

	"github.com/werf/werf/pkg/docker"
)

type DockerfileImageBuilder struct {
	ContainerRuntime       ContainerRuntime
	Dockerfile             []byte
	BuildDockerfileOptions BuildDockerfileOptions
	ContextArchivePath     string

	builtID string
}

func NewDockerfileImageBuilder(containerRuntime ContainerRuntime) *DockerfileImageBuilder {
	return &DockerfileImageBuilder{ContainerRuntime: containerRuntime}
}

// filePathToStdin != "" ??
func (b *DockerfileImageBuilder) Build(ctx context.Context) error {
	contextReader, err := os.Open(b.ContextArchivePath)
	if err != nil {
		return fmt.Errorf("unable to open context archive %q: %s", b.ContextArchivePath, err)
	}
	defer contextReader.Close()

	opts := b.BuildDockerfileOptions
	opts.ContextTar = contextReader

	if debug() {
		fmt.Printf("ContextArchivePath=%q\n", b.ContextArchivePath)
		fmt.Printf("BiuldDockerfileOptions: %#v\n", opts)
	}

	builtID, err := b.ContainerRuntime.BuildDockerfile(ctx, b.Dockerfile, opts)
	if err != nil {
		return fmt.Errorf("error building dockerfile with %s: %s", b.ContainerRuntime.String(), err)
	}

	b.builtID = builtID

	return nil
}

func (b *DockerfileImageBuilder) Cleanup(ctx context.Context) error {
	if err := docker.CliRmi(ctx, b.builtID, "--force"); err != nil {
		return fmt.Errorf("unable to remove built dockerfile image %q: %s", b.builtID, err)
	}
	return nil
}

func (b *DockerfileImageBuilder) GetBuiltId() string {
	return b.builtID
}

func (b *DockerfileImageBuilder) SetDockerfile(dockerfile []byte) {
	b.Dockerfile = dockerfile
}

func (b *DockerfileImageBuilder) SetDockerfileCtxRelPath(dockerfileCtxRelPath string) {
	b.BuildDockerfileOptions.DockerfileCtxRelPath = dockerfileCtxRelPath
}

func (b *DockerfileImageBuilder) SetTarget(target string) {
	b.BuildDockerfileOptions.Target = target
}

func (b *DockerfileImageBuilder) AppendBuildArgs(args ...string) {
	b.BuildDockerfileOptions.BuildArgs = append(b.BuildDockerfileOptions.BuildArgs, args...)
}

func (b *DockerfileImageBuilder) AppendAddHost(addHost ...string) {
	b.BuildDockerfileOptions.AddHost = append(b.BuildDockerfileOptions.AddHost, addHost...)
}

func (b *DockerfileImageBuilder) SetNetwork(network string) {
	b.BuildDockerfileOptions.Network = network
}

func (b *DockerfileImageBuilder) SetSSH(ssh string) {
	b.BuildDockerfileOptions.SSH = ssh
}

func (b *DockerfileImageBuilder) AppendLabels(labels ...string) {
	b.BuildDockerfileOptions.Labels = append(b.BuildDockerfileOptions.Labels, labels...)
}

func (b *DockerfileImageBuilder) SetContextArchivePath(contextArchivePath string) {
	b.ContextArchivePath = contextArchivePath
}
