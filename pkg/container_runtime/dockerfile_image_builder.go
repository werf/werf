package container_runtime

import (
	"context"
	"fmt"
	"os"
)

type DockerfileImageBuilder struct {
	ContainerRuntime       ContainerRuntime
	Dockerfile             []byte
	BuildDockerfileOptions BuildDockerfileOpts
	ContextArchivePath     string

	Image ImageInterface
}

func NewDockerfileImageBuilder(containerRuntime ContainerRuntime, image ImageInterface) *DockerfileImageBuilder {
	return &DockerfileImageBuilder{ContainerRuntime: containerRuntime, Image: image}
}

// filePathToStdin != "" ??
func (b *DockerfileImageBuilder) Build(ctx context.Context) error {
	if debug() {
		fmt.Printf("[DOCKER BUILD] context archive path: %s\n", b.ContextArchivePath)
	}

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

	b.Image.SetBuiltID(builtID)

	return nil
}

func (b *DockerfileImageBuilder) Cleanup(ctx context.Context) error {
	if err := b.ContainerRuntime.Rmi(ctx, b.Image.GetBuiltID(), RmiOpts{}); err != nil {
		return fmt.Errorf("unable to remove built dockerfile image %q: %s", b.Image.GetBuiltID(), err)
	}
	return nil
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
