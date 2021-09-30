package container_runtime

import (
	"context"
	"fmt"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/image"
)

type BuildahRuntime struct {
	buildah buildah.Buildah
}

type BuildahImage struct {
	Image LegacyImageInterface
}

func NewBuildahRuntime() (*BuildahRuntime, error) {
	var err error
	buildahRuntime := &BuildahRuntime{}

	buildahRuntime.buildah, err = buildah.NewBuildah(buildah.ModeAuto)
	if err != nil {
		return nil, err
	}

	return buildahRuntime, nil
}

func (runtime *BuildahRuntime) GetImageInfo(ctx context.Context, ref string) (*image.Info, error) {
	var b buildah.Buildah // FIXME: store buildah client in the BuildahRuntime object

	inspect, err := b.Inspect(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("error getting buildah inspect of %s: %s", ref, err)
	}

	repository, tag := image.ParseRepositoryAndTag(ref)

	return &image.Info{
		Name:              ref,
		Repository:        repository,
		Tag:               tag,
		Labels:            inspect.Docker.Config.Labels,
		CreatedAtUnixNano: inspect.Docker.Created.UnixNano(),
		// RepoDigest:        repoDigest, // FIXME
		ID:       inspect.Docker.ID,
		ParentID: inspect.Docker.Config.Image,
		Size:     inspect.Docker.Size,
	}, nil
}

func (runtime *BuildahRuntime) Rmi(ctx context.Context, ref string) error {
	panic("not implemented yet")
}

func (runtime *BuildahRuntime) Pull(ctx context.Context, ref string) error {
	var b buildah.Buildah // FIXME: store buildah client in the BuildahRuntime object
	return b.Pull(ctx, ref, buildah.PullOpts{})
}

func (runtime *BuildahRuntime) Tag(ctx context.Context, ref, newRef string) error {
	return runtime.buildah.Tag(ctx, ref, newRef)
}

func (runtime *BuildahRuntime) Push(ctx context.Context, ref string) error {
	return runtime.buildah.Push(ctx, ref, buildah.PushOpts{})
}

func (runtime *BuildahRuntime) BuildDockerfile(ctx context.Context, dockerfile []byte, opts BuildDockerfileOptions) (string, error) {
	return runtime.buildah.BuildFromDockerfile(ctx, dockerfile, buildah.BuildFromDockerfileOpts{
		CommonOpts: buildah.CommonOpts{
			LogWriter: logboek.Context(ctx).OutStream(),
		},
		ContextTar: opts.ContextTar,
	})
}

func (runtime *BuildahRuntime) RefreshImageObject(ctx context.Context, img LegacyImageInterface) error {
	if info, err := runtime.GetImageInfo(ctx, img.Name()); err != nil {
		return err
	} else {
		img.SetInfo(info)
	}
	return nil
}

func (runtime *BuildahRuntime) PullImageFromRegistry(ctx context.Context, img LegacyImageInterface) error {
	if err := img.Pull(ctx); err != nil {
		return fmt.Errorf("unable to pull image %s: %s", img.Name(), err)
	}

	if info, err := runtime.GetImageInfo(ctx, img.Name()); err != nil {
		return fmt.Errorf("unable to get inspect of image %s: %s", img.Name(), err)
	} else {
		img.SetInfo(info)
	}

	return nil
}

func (runtime *BuildahRuntime) RenameImage(ctx context.Context, img LegacyImageInterface, newImageName string, removeOldName bool) error {
	if err := logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Tagging image %s by name %s", img.Name(), newImageName)).DoError(func() error {
		if err := runtime.Tag(ctx, img.Name(), newImageName); err != nil {
			return fmt.Errorf("unable to tag image %s by name %s: %s", img.Name(), newImageName, err)
		}
		return nil
	}); err != nil {
		return err
	}

	if removeOldName {
		if err := logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Removing old image tag %s", img.Name())).DoError(func() error {
			if err := runtime.Rmi(ctx, img.Name()); err != nil {
				return fmt.Errorf("unable to remove image %q: %s", img.Name(), err)
			}
			return nil
		}); err != nil {
			return err
		}
	}

	img.SetName(newImageName)

	if info, err := runtime.GetImageInfo(ctx, img.Name()); err != nil {
		return err
	} else {
		img.SetInfo(info)
	}

	desc := img.GetStageDescription()

	repository, tag := image.ParseRepositoryAndTag(newImageName)
	desc.Info.Repository = repository
	desc.Info.Tag = tag

	return nil
}

func (runtime *BuildahRuntime) RemoveImage(ctx context.Context, img LegacyImageInterface) error {
	if err := logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Removing image tag %s", img.Name())).DoError(func() error {
		if err := runtime.Rmi(ctx, img.Name()); err != nil {
			return fmt.Errorf("unable to remove image %q: %s", img.Name(), err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (runtime *BuildahRuntime) String() string {
	return "buildah-runtime"
}
