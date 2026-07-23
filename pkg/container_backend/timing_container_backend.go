package container_backend

import (
	"context"
	"io"

	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/opstats"
)

var _ ContainerBackend = (*TimingContainerBackend)(nil)

// TimingContainerBackend records wall-clock durations of the lowest-level
// backend operations into the opstats collector bound to the context.
type TimingContainerBackend struct {
	ContainerBackend
}

func NewTimingContainerBackend(backend ContainerBackend) *TimingContainerBackend {
	return &TimingContainerBackend{ContainerBackend: backend}
}

func (t *TimingContainerBackend) Pull(ctx context.Context, ref string, opts PullOpts) error {
	defer opstats.Observe(ctx, opstats.OperationImagePull)()
	return t.ContainerBackend.Pull(ctx, ref, opts)
}

func (t *TimingContainerBackend) PullImageFromRegistry(ctx context.Context, img LegacyImageInterface) error {
	defer opstats.Observe(ctx, opstats.OperationImagePull)()
	return t.ContainerBackend.PullImageFromRegistry(ctx, img)
}

func (t *TimingContainerBackend) Push(ctx context.Context, ref string, opts PushOpts) error {
	defer opstats.Observe(ctx, opstats.OperationImagePush)()
	return t.ContainerBackend.Push(ctx, ref, opts)
}

func (t *TimingContainerBackend) BuildDockerfile(ctx context.Context, dockerfile []byte, opts BuildDockerfileOpts) (string, error) {
	defer opstats.Observe(ctx, opstats.OperationImageBuild)()
	return t.ContainerBackend.BuildDockerfile(ctx, dockerfile, opts)
}

func (t *TimingContainerBackend) BuildDockerfileStage(ctx context.Context, baseImage string, opts BuildDockerfileStageOptions, instructions ...InstructionInterface) (string, error) {
	defer opstats.Observe(ctx, opstats.OperationImageBuild)()
	return t.ContainerBackend.BuildDockerfileStage(ctx, baseImage, opts, instructions...)
}

func (t *TimingContainerBackend) BuildStapelStage(ctx context.Context, baseImage string, opts BuildStapelStageOptions) (string, error) {
	defer opstats.Observe(ctx, opstats.OperationImageBuild)()
	return t.ContainerBackend.BuildStapelStage(ctx, baseImage, opts)
}

func (t *TimingContainerBackend) GetImageInfo(ctx context.Context, ref string, opts GetImageInfoOpts) (*image.Info, error) {
	defer opstats.Observe(ctx, opstats.OperationImageInfo)()
	return t.ContainerBackend.GetImageInfo(ctx, ref, opts)
}

func (t *TimingContainerBackend) CalculateDependencyImportChecksum(ctx context.Context, dependencyImport DependencyImportSpec, opts CalculateDependencyImportChecksum) (string, error) {
	defer opstats.Observe(ctx, opstats.OperationImportChecksum)()
	return t.ContainerBackend.CalculateDependencyImportChecksum(ctx, dependencyImport, opts)
}

func (t *TimingContainerBackend) SaveImageToStream(ctx context.Context, imageName string) (io.ReadCloser, error) {
	defer opstats.Observe(ctx, opstats.OperationImageSaveLoad)()
	return t.ContainerBackend.SaveImageToStream(ctx, imageName)
}

func (t *TimingContainerBackend) LoadImageFromStream(ctx context.Context, input io.Reader) (string, error) {
	defer opstats.Observe(ctx, opstats.OperationImageSaveLoad)()
	return t.ContainerBackend.LoadImageFromStream(ctx, input)
}
