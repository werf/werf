package stage_builder

import (
	"context"
	"fmt"
	"io"

	"github.com/werf/werf/pkg/container_backend"
)

type DockerfileStageBuilderInterface interface {
	AppendPreCommands(commands ...any) DockerfileStageBuilderInterface
	AppendMainCommands(commands ...any) DockerfileStageBuilderInterface
	AppendPostCommands(commands ...any) DockerfileStageBuilderInterface
	SetContext(contextTar io.ReadCloser) DockerfileStageBuilderInterface
	Build(ctx context.Context, opts container_backend.BuildOptions) error
}

type DockerfileStageBuilder struct {
	preCommands  []any
	mainCommands []any
	postCommands []any
	contextTar   io.ReadCloser

	baseImage        string
	resultImage      container_backend.ImageInterface
	containerBackend container_backend.ContainerBackend
}

func NewDockerfileStageBuilder(containerBackend container_backend.ContainerBackend, baseImage string, resultImage container_backend.ImageInterface) *DockerfileStageBuilder {
	return &DockerfileStageBuilder{
		containerBackend: containerBackend,
		baseImage:        baseImage,
		resultImage:      resultImage,
	}
}

func (b *DockerfileStageBuilder) AppendPreCommands(commands ...any) DockerfileStageBuilderInterface {
	b.preCommands = append(b.preCommands, commands...)
	return b
}

func (b *DockerfileStageBuilder) AppendMainCommands(commands ...any) DockerfileStageBuilderInterface {
	b.mainCommands = append(b.mainCommands, commands...)
	return b
}

func (b *DockerfileStageBuilder) AppendPostCommands(commands ...any) DockerfileStageBuilderInterface {
	b.postCommands = append(b.postCommands, commands...)
	return b
}

func (b *DockerfileStageBuilder) SetContext(contextTar io.ReadCloser) DockerfileStageBuilderInterface {
	b.contextTar = contextTar
	return b
}

func (b *DockerfileStageBuilder) Build(ctx context.Context, opts container_backend.BuildOptions) error {
	commands := append(append(b.preCommands, b.mainCommands...), b.postCommands...)
	backendOpts := container_backend.BuildDockerfileStageOptions{ContextTar: b.contextTar}

	if builtID, err := b.containerBackend.BuildDockerfileStage(ctx, b.baseImage, backendOpts, commands...); err != nil {
		return fmt.Errorf("error building dockerfile stage: %w", err)
	} else {
		b.resultImage.SetBuiltID(builtID)
	}

	return nil
}
