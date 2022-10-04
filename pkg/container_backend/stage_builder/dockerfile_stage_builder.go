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
	Build(ctx context.Context, contextTar io.ReadCloser) error
}

type DockerfileStageBuilder struct {
	preCommands  []any
	mainCommands []any
	postCommands []any

	fromImage        container_backend.ImageInterface
	resultImage      container_backend.ImageInterface
	containerBackend container_backend.ContainerBackend
}

func NewDockerfileStageBuilder(containerBackend container_backend.ContainerBackend, fromImage, resultImage container_backend.ImageInterface) *DockerfileStageBuilder {
	return &DockerfileStageBuilder{
		containerBackend: containerBackend,
		fromImage:        fromImage,
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

func (b *DockerfileStageBuilder) Build(ctx context.Context, contextTar io.ReadCloser) error {
	commands := append(append(b.preCommands, b.mainCommands...), b.postCommands)
	if len(commands) == 0 {
		b.resultImage.SetName(b.fromImage.Name())
		b.resultImage.SetBuiltID(b.fromImage.BuiltID())
		return nil
	}

	if builtID, err := b.containerBackend.BuildDockerfileStage(ctx, b.resultImage, contextTar, container_backend.BuildDockerfileStageOptions{}, commands...); err != nil {
		return fmt.Errorf("error building dockerfile stage: %w", err)
	} else {
		b.resultImage.SetBuiltID(builtID)
	}

	return nil
}
