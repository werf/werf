package container_backend

import (
	"context"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/image"
)

type PerfCheckContainerBackend struct {
	ContainerBackend ContainerBackend
}

func NewPerfCheckContainerBackend(containerBackend ContainerBackend) *PerfCheckContainerBackend {
	return &PerfCheckContainerBackend{ContainerBackend: containerBackend}
}

func (runtime *PerfCheckContainerBackend) HasStapelBuildSupport() bool {
	return runtime.ContainerBackend.HasStapelBuildSupport()
}

func (runtime *PerfCheckContainerBackend) GetImageInfo(ctx context.Context, ref string, opts GetImageInfoOpts) (resImg *image.Info, resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerBackend.GetImageInfo %q", ref).
		Do(func() {
			resImg, resErr = runtime.ContainerBackend.GetImageInfo(ctx, ref, opts)
		})
	return
}

func (runtime *PerfCheckContainerBackend) Rmi(ctx context.Context, ref string, opts RmiOpts) (resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerBackend.Rmi %q", ref).
		Do(func() {
			resErr = runtime.ContainerBackend.Rmi(ctx, ref, opts)
		})
	return
}

func (runtime *PerfCheckContainerBackend) Pull(ctx context.Context, ref string, opts PullOpts) (resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerBackend.Pull %q", ref).
		Do(func() {
			resErr = runtime.ContainerBackend.Pull(ctx, ref, opts)
		})
	return
}

func (runtime *PerfCheckContainerBackend) Tag(ctx context.Context, ref, newRef string, opts TagOpts) (resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerBackend.Tag %q as %q", ref, newRef).
		Do(func() {
			resErr = runtime.ContainerBackend.Tag(ctx, ref, newRef, opts)
		})
	return
}

func (runtime *PerfCheckContainerBackend) Push(ctx context.Context, ref string, opts PushOpts) (resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerBackend.Push %q", ref).
		Do(func() {
			resErr = runtime.ContainerBackend.Push(ctx, ref, opts)
		})
	return
}

func (runtime *PerfCheckContainerBackend) BuildDockerfile(ctx context.Context, dockerfile []byte, opts BuildDockerfileOpts) (resID string, resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerBackend.BuildDockerfile").
		Do(func() {
			resID, resErr = runtime.ContainerBackend.BuildDockerfile(ctx, dockerfile, opts)
		})
	return
}

func (runtime *PerfCheckContainerBackend) BuildStapelStage(ctx context.Context, opts BuildStapelStageOptions) (resID string, resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerBackend.BuildDockerfile").
		Do(func() {
			resID, resErr = runtime.ContainerBackend.BuildStapelStage(ctx, opts)
		})
	return
}

func (runtime *PerfCheckContainerBackend) RefreshImageObject(ctx context.Context, img LegacyImageInterface) (resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerBackend.RefreshImageObject %q", img.Name()).
		Do(func() {
			resErr = runtime.ContainerBackend.RefreshImageObject(ctx, img)
		})
	return
}

func (runtime *PerfCheckContainerBackend) PullImageFromRegistry(ctx context.Context, img LegacyImageInterface) (resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerBackend.PullImageFromRegistry %q", img.Name()).
		Do(func() {
			resErr = runtime.ContainerBackend.PullImageFromRegistry(ctx, img)
		})
	return
}

func (runtime *PerfCheckContainerBackend) RenameImage(ctx context.Context, img LegacyImageInterface, newImageName string, removeOldName bool) (resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerBackend.RenameImage %q to %q", img.Name(), newImageName).
		Do(func() {
			resErr = runtime.ContainerBackend.RenameImage(ctx, img, newImageName, removeOldName)
		})
	return
}

func (runtime *PerfCheckContainerBackend) RemoveImage(ctx context.Context, img LegacyImageInterface) (resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerBackend.RemoveImage %q", img.Name()).
		Do(func() {
			resErr = runtime.ContainerBackend.RemoveImage(ctx, img)
		})
	return
}

func (runtime *PerfCheckContainerBackend) String() string {
	return runtime.ContainerBackend.String()
}
