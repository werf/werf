package container_runtime

import (
	"context"
	"fmt"
	"strings"

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

func NewBuildahRuntime(buildah buildah.Buildah) *BuildahRuntime {
	return &BuildahRuntime{buildah: buildah}
}

func (runtime *BuildahRuntime) HasContainerRootMountSupport() bool {
	return true
}

// GetImageInfo returns nil, nil if image not found.
func (runtime *BuildahRuntime) GetImageInfo(ctx context.Context, ref string, opts GetImageInfoOpts) (*image.Info, error) {
	inspect, err := runtime.buildah.Inspect(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("error getting buildah inspect of %q: %s", ref, err)
	}
	if inspect == nil {
		return nil, nil
	}

	repository, tag := image.ParseRepositoryAndTag(ref)

	return &image.Info{
		Name:              ref,
		Repository:        repository,
		Tag:               tag,
		Labels:            inspect.Docker.Config.Labels,
		CreatedAtUnixNano: inspect.Docker.Created.UnixNano(),
		// RepoDigest:        repoDigest, // FIXME
		OnBuild:  inspect.Docker.Config.OnBuild,
		ID:       inspect.Docker.ID,
		ParentID: inspect.Docker.Config.Image,
		Size:     inspect.Docker.Size,
	}, nil
}

func (runtime *BuildahRuntime) Rmi(ctx context.Context, ref string, opts RmiOpts) error {
	return runtime.buildah.Rmi(ctx, ref, buildah.RmiOpts{
		Force: true,
		CommonOpts: buildah.CommonOpts{
			LogWriter: logboek.Context(ctx).OutStream(),
		},
	})
}

func (runtime *BuildahRuntime) Pull(ctx context.Context, ref string, opts PullOpts) error {
	return runtime.buildah.Pull(ctx, ref, buildah.PullOpts{
		LogWriter: logboek.Context(ctx).OutStream(),
	})
}

func (runtime *BuildahRuntime) Tag(ctx context.Context, ref, newRef string, opts TagOpts) error {
	return runtime.buildah.Tag(ctx, ref, newRef, buildah.TagOpts{
		LogWriter: logboek.Context(ctx).OutStream(),
	})
}

func (runtime *BuildahRuntime) Push(ctx context.Context, ref string, opts PushOpts) error {
	return runtime.buildah.Push(ctx, ref, buildah.PushOpts{
		LogWriter: logboek.Context(ctx).OutStream(),
	})
}

func (runtime *BuildahRuntime) BuildDockerfile(ctx context.Context, dockerfile []byte, opts BuildDockerfileOpts) (string, error) {
	buildArgs := make(map[string]string)
	for _, argStr := range opts.BuildArgs {
		argParts := strings.SplitN(argStr, "=", 2)
		if len(argParts) < 2 {
			return "", fmt.Errorf("invalid build argument %q given, expected string in the key=value format", argStr)
		}
		buildArgs[argParts[0]] = argParts[1]
	}

	return runtime.buildah.BuildFromDockerfile(ctx, dockerfile, buildah.BuildFromDockerfileOpts{
		CommonOpts: buildah.CommonOpts{
			LogWriter: logboek.Context(ctx).OutStream(),
		},
		ContextTar: opts.ContextTar,
		BuildArgs:  buildArgs,
		Target:     opts.Target,
	})
}

func (runtime *BuildahRuntime) RefreshImageObject(ctx context.Context, img LegacyImageInterface) error {
	if info, err := runtime.GetImageInfo(ctx, img.Name(), GetImageInfoOpts{}); err != nil {
		return err
	} else {
		img.SetInfo(info)
	}
	return nil
}

func (runtime *BuildahRuntime) PullImageFromRegistry(ctx context.Context, img LegacyImageInterface) error {
	if err := runtime.Pull(ctx, img.Name(), PullOpts{}); err != nil {
		return fmt.Errorf("unable to pull image %s: %s", img.Name(), err)
	}

	if info, err := runtime.GetImageInfo(ctx, img.Name(), GetImageInfoOpts{}); err != nil {
		return fmt.Errorf("unable to get inspect of image %s: %s", img.Name(), err)
	} else {
		img.SetInfo(info)
	}

	return nil
}

func (runtime *BuildahRuntime) RenameImage(ctx context.Context, img LegacyImageInterface, newImageName string, removeOldName bool) error {
	if err := logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Tagging image %s by name %s", img.Name(), newImageName)).DoError(func() error {
		if err := runtime.Tag(ctx, img.Name(), newImageName, TagOpts{}); err != nil {
			return fmt.Errorf("unable to tag image %s by name %s: %s", img.Name(), newImageName, err)
		}
		return nil
	}); err != nil {
		return err
	}

	if removeOldName {
		if err := logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Removing old image tag %s", img.Name())).DoError(func() error {
			if err := runtime.Rmi(ctx, img.Name(), RmiOpts{}); err != nil {
				return fmt.Errorf("unable to remove image %q: %s", img.Name(), err)
			}
			return nil
		}); err != nil {
			return err
		}
	}

	img.SetName(newImageName)

	if info, err := runtime.GetImageInfo(ctx, img.Name(), GetImageInfoOpts{}); err != nil {
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
		if err := runtime.Rmi(ctx, img.Name(), RmiOpts{}); err != nil {
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
