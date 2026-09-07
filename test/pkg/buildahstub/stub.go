package buildahstub

import (
	"context"
	"io"
	"sync"

	"github.com/werf/werf/v2/pkg/buildah"
	"github.com/werf/werf/v2/pkg/buildah/thirdparty"
	"github.com/werf/werf/v2/pkg/container_backend/info"
	"github.com/werf/werf/v2/pkg/image"
)

type BuildahStub struct {
	callsMutex        sync.Mutex
	FromCommandFunc   func(ctx context.Context, container, image string, opts buildah.FromCommandOpts) (string, error)
	PullFunc          func(ctx context.Context, ref string, opts buildah.PullOpts) (string, error)
	InspectFunc       func(ctx context.Context, ref string) (*thirdparty.BuilderInfo, error)
	FromCommandImages []string
	PullRefs          []string
	InspectRefs       []string
}

var _ buildah.Buildah = (*BuildahStub)(nil)

func (b *BuildahStub) Info(context.Context) (info.Info, error) {
	return info.Info{}, nil
}

func (b *BuildahStub) GetDefaultPlatform() string {
	return ""
}

func (b *BuildahStub) GetRuntimePlatform() string {
	return ""
}

func (b *BuildahStub) Tag(context.Context, string, string, buildah.TagOpts) error {
	return nil
}

func (b *BuildahStub) Push(context.Context, string, buildah.PushOpts) error {
	return nil
}

func (b *BuildahStub) BuildFromDockerfile(context.Context, string, buildah.BuildFromDockerfileOpts) (string, error) {
	return "", nil
}

func (b *BuildahStub) RunCommand(context.Context, string, []string, buildah.RunCommandOpts) error {
	return nil
}

func (b *BuildahStub) FromCommand(ctx context.Context, container, image string, opts buildah.FromCommandOpts) (string, error) {
	b.callsMutex.Lock()
	b.FromCommandImages = append(b.FromCommandImages, image)
	b.callsMutex.Unlock()

	if b.FromCommandFunc != nil {
		return b.FromCommandFunc(ctx, container, image, opts)
	}
	return "", nil
}

func (b *BuildahStub) Pull(ctx context.Context, ref string, opts buildah.PullOpts) (string, error) {
	b.callsMutex.Lock()
	b.PullRefs = append(b.PullRefs, ref)
	b.callsMutex.Unlock()

	if b.PullFunc != nil {
		return b.PullFunc(ctx, ref, opts)
	}
	return "", nil
}

func (b *BuildahStub) Inspect(ctx context.Context, ref string) (*thirdparty.BuilderInfo, error) {
	b.callsMutex.Lock()
	b.InspectRefs = append(b.InspectRefs, ref)
	b.callsMutex.Unlock()

	if b.InspectFunc != nil {
		return b.InspectFunc(ctx, ref)
	}
	return nil, nil
}

func (b *BuildahStub) Rm(context.Context, string, buildah.RmOpts) error {
	return nil
}

func (b *BuildahStub) Rmi(context.Context, string, buildah.RmiOpts) error {
	return nil
}

func (b *BuildahStub) Mount(context.Context, string, buildah.MountOpts) (string, error) {
	return "", nil
}

func (b *BuildahStub) Umount(context.Context, string, buildah.UmountOpts) error {
	return nil
}

func (b *BuildahStub) Commit(context.Context, string, buildah.CommitOpts) (string, error) {
	return "", nil
}

func (b *BuildahStub) Config(context.Context, string, buildah.ConfigOpts) error {
	return nil
}

func (b *BuildahStub) Copy(context.Context, string, string, []string, string, buildah.CopyOpts) error {
	return nil
}

func (b *BuildahStub) Add(context.Context, string, []string, string, buildah.AddOpts) error {
	return nil
}

func (b *BuildahStub) Images(context.Context, buildah.ImagesOptions) (image.ImagesList, error) {
	return nil, nil
}

func (b *BuildahStub) Containers(context.Context, buildah.ContainersOptions) (image.ContainerList, error) {
	return nil, nil
}

func (b *BuildahStub) PruneImages(context.Context, buildah.PruneImagesOptions) (buildah.PruneImagesReport, error) {
	return buildah.PruneImagesReport{}, nil
}

func (b *BuildahStub) SaveImageToStream(context.Context, string) (io.ReadCloser, error) {
	return nil, nil
}

func (b *BuildahStub) LoadImageFromStream(context.Context, io.Reader) (string, error) {
	return "", nil
}
