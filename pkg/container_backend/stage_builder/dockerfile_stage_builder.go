package stage_builder

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/container_backend"
)

type DockerfileStageBuilderInterface interface {
	DockerfileStageIntructionBuilderInterface
	SetBuildContextArchive(buildContextArchive container_backend.BuildContextArchiver) DockerfileStageBuilderInterface
	Build(ctx context.Context, opts container_backend.BuildOptions) error
}

type DockerfileStageIntructionBuilderInterface interface {
	AppendPreInstruction(i container_backend.InstructionInterface) DockerfileStageBuilderInterface
	AppendInstruction(i container_backend.InstructionInterface) DockerfileStageBuilderInterface
	AppendPostInstruction(i container_backend.InstructionInterface) DockerfileStageBuilderInterface
}

type DockerfileStageBuilder struct {
	preInstructions     []container_backend.InstructionInterface
	instructions        []container_backend.InstructionInterface
	postInstructions    []container_backend.InstructionInterface
	buildContextArchive container_backend.BuildContextArchiver

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

func (b *DockerfileStageBuilder) AppendPreInstruction(i container_backend.InstructionInterface) DockerfileStageBuilderInterface {
	b.preInstructions = append(b.preInstructions, i)
	return b
}

func (b *DockerfileStageBuilder) AppendInstruction(i container_backend.InstructionInterface) DockerfileStageBuilderInterface {
	b.instructions = append(b.instructions, i)
	return b
}

func (b *DockerfileStageBuilder) AppendPostInstruction(i container_backend.InstructionInterface) DockerfileStageBuilderInterface {
	b.postInstructions = append(b.postInstructions, i)
	return b
}

func (b *DockerfileStageBuilder) SetBuildContextArchive(buildContextArchive container_backend.BuildContextArchiver) DockerfileStageBuilderInterface {
	b.buildContextArchive = buildContextArchive
	return b
}

func (b *DockerfileStageBuilder) Build(ctx context.Context, opts container_backend.BuildOptions) error {
	instructions := append(append(b.preInstructions, b.instructions...), b.postInstructions...)
	backendOpts := container_backend.BuildDockerfileStageOptions{
		CommonOpts:          container_backend.CommonOpts{TargetPlatform: opts.TargetPlatform},
		BuildContextArchive: b.buildContextArchive,
	}

	if builtID, err := b.containerBackend.BuildDockerfileStage(ctx, b.baseImage, backendOpts, instructions...); err != nil {
		return fmt.Errorf("error building dockerfile stage: %w", err)
	} else {
		b.resultImage.SetBuiltID(builtID)
	}

	return nil
}
