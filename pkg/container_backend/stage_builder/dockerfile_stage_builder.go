package stage_builder

import (
	"context"
	"fmt"
	"os"

	"github.com/werf/werf/pkg/container_backend"
)

type DockerfileStageBuilderInterface interface {
	Build(ctx context.Context) error
	Cleanup(ctx context.Context) error
	SetDockerfile(dockerfile []byte)
	SetDockerfileCtxRelPath(dockerfileCtxRelPath string)
	SetTarget(target string)
	AppendBuildArgs(args ...string)
	AppendAddHost(addHost ...string)
	SetNetwork(network string)
	SetSSH(ssh string)
	AppendLabels(labels ...string)
	SetContextArchivePath(contextArchivePath string)
}

type DockerfileStageBuilder struct {
	ContainerBackend       container_backend.ContainerBackend
	Dockerfile             []byte
	BuildDockerfileOptions container_backend.BuildDockerfileOpts
	ContextArchivePath     string

	Image container_backend.ImageInterface
}

func NewDockerfileStageBuilder(containerBackend container_backend.ContainerBackend, image container_backend.ImageInterface) *DockerfileStageBuilder {
	return &DockerfileStageBuilder{ContainerBackend: containerBackend, Image: image}
}

func (b *DockerfileStageBuilder) Build(ctx context.Context) error {
	// filePathToStdin != "" ??

	if container_backend.Debug() {
		fmt.Printf("[DOCKER BUILD] context archive path: %s\n", b.ContextArchivePath)
	}

	contextReader, err := os.Open(b.ContextArchivePath)
	if err != nil {
		return fmt.Errorf("unable to open context archive %q: %s", b.ContextArchivePath, err)
	}
	defer contextReader.Close()

	opts := b.BuildDockerfileOptions
	opts.ContextTar = contextReader

	if container_backend.Debug() {
		fmt.Printf("ContextArchivePath=%q\n", b.ContextArchivePath)
		fmt.Printf("BiuldDockerfileOptions: %#v\n", opts)
	}

	builtID, err := b.ContainerBackend.BuildDockerfile(ctx, b.Dockerfile, opts)
	if err != nil {
		return fmt.Errorf("error building dockerfile with %s: %s", b.ContainerBackend.String(), err)
	}

	b.Image.SetBuiltID(builtID)

	return nil
}

func (b *DockerfileStageBuilder) Cleanup(ctx context.Context) error {
	if err := b.ContainerBackend.Rmi(ctx, b.Image.BuiltID(), container_backend.RmiOpts{}); err != nil {
		return fmt.Errorf("unable to remove built dockerfile image %q: %s", b.Image.BuiltID(), err)
	}
	return nil
}

func (b *DockerfileStageBuilder) SetDockerfile(dockerfile []byte) {
	b.Dockerfile = dockerfile
}

func (b *DockerfileStageBuilder) SetDockerfileCtxRelPath(dockerfileCtxRelPath string) {
	b.BuildDockerfileOptions.DockerfileCtxRelPath = dockerfileCtxRelPath
}

func (b *DockerfileStageBuilder) SetTarget(target string) {
	b.BuildDockerfileOptions.Target = target
}

func (b *DockerfileStageBuilder) AppendBuildArgs(args ...string) {
	b.BuildDockerfileOptions.BuildArgs = append(b.BuildDockerfileOptions.BuildArgs, args...)
}

func (b *DockerfileStageBuilder) AppendAddHost(addHost ...string) {
	b.BuildDockerfileOptions.AddHost = append(b.BuildDockerfileOptions.AddHost, addHost...)
}

func (b *DockerfileStageBuilder) SetNetwork(network string) {
	b.BuildDockerfileOptions.Network = network
}

func (b *DockerfileStageBuilder) SetSSH(ssh string) {
	b.BuildDockerfileOptions.SSH = ssh
}

func (b *DockerfileStageBuilder) AppendLabels(labels ...string) {
	b.BuildDockerfileOptions.Labels = append(b.BuildDockerfileOptions.Labels, labels...)
}

func (b *DockerfileStageBuilder) SetContextArchivePath(contextArchivePath string) {
	b.ContextArchivePath = contextArchivePath
}
