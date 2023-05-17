package stage_builder

import (
	"context"
	"fmt"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/container_backend"
)

type DockerfileBuilderInterface interface {
	Build(ctx context.Context, opts container_backend.BuildOptions) error
	Cleanup(ctx context.Context) error
	SetDockerfile(dockerfile []byte)
	SetDockerfileCtxRelPath(dockerfileCtxRelPath string)
	SetTarget(target string)
	AppendBuildArgs(args ...string)
	AppendAddHost(addHost ...string)
	SetNetwork(network string)
	SetSSH(ssh string)
	AppendLabels(labels ...string)
	SetBuildContextArchive(buildContextArchive container_backend.BuildContextArchiver)
}

type DockerfileBuilder struct {
	ContainerBackend       container_backend.ContainerBackend
	Dockerfile             []byte
	BuildDockerfileOptions container_backend.BuildDockerfileOpts
	BuildContextArchive    container_backend.BuildContextArchiver

	Image container_backend.ImageInterface
}

func NewDockerfileBuilder(containerBackend container_backend.ContainerBackend, image container_backend.ImageInterface) *DockerfileBuilder {
	return &DockerfileBuilder{ContainerBackend: containerBackend, Image: image}
}

func (b *DockerfileBuilder) Build(ctx context.Context, opts container_backend.BuildOptions) error {
	// filePathToStdin != "" ??

	if container_backend.Debug() {
		fmt.Printf("[DOCKER BUILD] context archive path: %s\n", b.BuildContextArchive.Path())
	}

	finalOpts := b.BuildDockerfileOptions
	finalOpts.BuildContextArchive = b.BuildContextArchive
	finalOpts.TargetPlatform = opts.TargetPlatform

	if container_backend.Debug() {
		fmt.Printf("BuildContextArchive=%q\n", b.BuildContextArchive)
		fmt.Printf("BiuldDockerfileOptions: %#v\n", opts)
	}

	builtID, err := b.ContainerBackend.BuildDockerfile(ctx, b.Dockerfile, finalOpts)
	if err != nil {
		return fmt.Errorf("error building dockerfile with %s: %w", b.ContainerBackend.String(), err)
	}
	b.Image.SetBuiltID(builtID)

	return nil
}

func (b *DockerfileBuilder) Cleanup(ctx context.Context) error {
	if !b.ContainerBackend.ShouldCleanupDockerfileImage() {
		return nil
	}

	if b.Image.BuiltID() != "" {
		logboek.Context(ctx).Info().LogF("Cleanup built dockerfile image %q\n", b.Image.BuiltID())
		if err := b.ContainerBackend.Rmi(ctx, b.Image.BuiltID(), container_backend.RmiOpts{}); err != nil {
			return fmt.Errorf("unable to remove built dockerfile image %q: %w", b.Image.BuiltID(), err)
		}
	}
	return nil
}

func (b *DockerfileBuilder) SetDockerfile(dockerfile []byte) {
	b.Dockerfile = dockerfile
}

func (b *DockerfileBuilder) SetDockerfileCtxRelPath(dockerfileCtxRelPath string) {
	b.BuildDockerfileOptions.DockerfileCtxRelPath = dockerfileCtxRelPath
}

func (b *DockerfileBuilder) SetTarget(target string) {
	b.BuildDockerfileOptions.Target = target
}

func (b *DockerfileBuilder) AppendBuildArgs(args ...string) {
	b.BuildDockerfileOptions.BuildArgs = append(b.BuildDockerfileOptions.BuildArgs, args...)
}

func (b *DockerfileBuilder) AppendAddHost(addHost ...string) {
	b.BuildDockerfileOptions.AddHost = append(b.BuildDockerfileOptions.AddHost, addHost...)
}

func (b *DockerfileBuilder) SetNetwork(network string) {
	b.BuildDockerfileOptions.Network = network
}

func (b *DockerfileBuilder) SetSSH(ssh string) {
	b.BuildDockerfileOptions.SSH = ssh
}

func (b *DockerfileBuilder) AppendLabels(labels ...string) {
	b.BuildDockerfileOptions.Labels = append(b.BuildDockerfileOptions.Labels, labels...)
}

func (b *DockerfileBuilder) SetBuildContextArchive(buildContextArchive container_backend.BuildContextArchiver) {
	b.BuildContextArchive = buildContextArchive
}
