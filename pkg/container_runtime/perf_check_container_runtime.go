package container_runtime

import (
	"context"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/image"
)

type PerfCheckContainerRuntime struct {
	ContainerRuntime ContainerRuntime
}

func NewPerfCheckContainerRuntime(containerRuntime ContainerRuntime) *PerfCheckContainerRuntime {
	return &PerfCheckContainerRuntime{ContainerRuntime: containerRuntime}
}

func (runtime *PerfCheckContainerRuntime) GetImageInfo(ctx context.Context, ref string) (resImg *image.Info, resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerRuntime.GetImageInfo %q", ref).
		Do(func() {
			resImg, resErr = runtime.ContainerRuntime.GetImageInfo(ctx, ref)
		})
	return
}

func (runtime *PerfCheckContainerRuntime) Rmi(ctx context.Context, ref string) (resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerRuntime.Rmi %q", ref).
		Do(func() {
			resErr = runtime.ContainerRuntime.Rmi(ctx, ref)
		})
	return
}

func (runtime *PerfCheckContainerRuntime) Pull(ctx context.Context, ref string) (resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerRuntime.Pull %q", ref).
		Do(func() {
			resErr = runtime.ContainerRuntime.Pull(ctx, ref)
		})
	return
}

func (runtime *PerfCheckContainerRuntime) Tag(ctx context.Context, ref, newRef string) (resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerRuntime.Tag %q as %q", ref, newRef).
		Do(func() {
			resErr = runtime.ContainerRuntime.Tag(ctx, ref, newRef)
		})
	return
}

func (runtime *PerfCheckContainerRuntime) Push(ctx context.Context, ref string) (resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerRuntime.Push %q", ref).
		Do(func() {
			resErr = runtime.ContainerRuntime.Push(ctx, ref)
		})
	return
}

func (runtime *PerfCheckContainerRuntime) BuildDockerfile(ctx context.Context, dockerfile []byte, opts BuildDockerfileOptions) (resID string, resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerRuntime.BuildDockerfile").
		Do(func() {
			resID, resErr = runtime.ContainerRuntime.BuildDockerfile(ctx, dockerfile, opts)
		})
	return
}

func (runtime *PerfCheckContainerRuntime) RefreshImageObject(ctx context.Context, img LegacyImageInterface) (resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerRuntime.RefreshImageObject %q", img.Name()).
		Do(func() {
			resErr = runtime.ContainerRuntime.RefreshImageObject(ctx, img)
		})
	return
}

func (runtime *PerfCheckContainerRuntime) PullImageFromRegistry(ctx context.Context, img LegacyImageInterface) (resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerRuntime.PullImageFromRegistry %q", img.Name()).
		Do(func() {
			resErr = runtime.ContainerRuntime.PullImageFromRegistry(ctx, img)
		})
	return
}

func (runtime *PerfCheckContainerRuntime) RenameImage(ctx context.Context, img LegacyImageInterface, newImageName string, removeOldName bool) (resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerRuntime.RenameImage %q to %q", img.Name(), newImageName).
		Do(func() {
			resErr = runtime.ContainerRuntime.RenameImage(ctx, img, newImageName, removeOldName)
		})
	return
}

func (runtime *PerfCheckContainerRuntime) RemoveImage(ctx context.Context, img LegacyImageInterface) (resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerRuntime.RemoveImage %q", img.Name()).
		Do(func() {
			resErr = runtime.ContainerRuntime.RemoveImage(ctx, img)
		})
	return
}

func (runtime *PerfCheckContainerRuntime) String() string {
	return runtime.ContainerRuntime.String()
}
