package container_runtime

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/docker"
)

type ContainerRuntime interface {
	RefreshImageObject(ctx context.Context, img Image) error
	PullImageFromRegistry(ctx context.Context, img Image) error
	RenameImage(ctx context.Context, img Image, newImageName string, removeOldName bool) error
	RemoveImage(ctx context.Context, img Image) error
	String() string
}

type LocalDockerServerRuntime struct{}

// GetImageInspect only available for LocalDockerServerRuntime
func (runtime *LocalDockerServerRuntime) GetImageInspect(ctx context.Context, ref string) (*types.ImageInspect, error) {
	inspect, err := docker.ImageInspect(ctx, ref)
	if client.IsErrNotFound(err) {
		return nil, nil
	}
	return inspect, err
}

// PullImage only available for LocalDockerServerRuntime
func (runtime *LocalDockerServerRuntime) PullImage(ctx context.Context, ref string) error {
	if err := docker.CliPull(ctx, ref); err != nil {
		return fmt.Errorf("unable to pull image %s: %s", ref, err)
	}

	return nil
}

func (runtime *LocalDockerServerRuntime) RefreshImageObject(ctx context.Context, img Image) error {
	dockerImage := img.(*DockerImage)

	if inspect, err := runtime.GetImageInspect(ctx, dockerImage.Image.Name()); err != nil {
		return err
	} else {
		dockerImage.Image.SetInspect(inspect)
	}
	return nil
}

func (runtime *LocalDockerServerRuntime) RenameImage(ctx context.Context, img Image, newImageName string, removeOldName bool) error {
	dockerImage := img.(*DockerImage)

	if err := logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Tagging image %s by name %s", dockerImage.Image.Name(), newImageName)).DoError(func() error {
		if err := docker.CliTag(ctx, dockerImage.Image.Name(), newImageName); err != nil {
			return fmt.Errorf("unable to tag image %s by name %s: %s", dockerImage.Image.Name(), newImageName, err)
		}
		return nil
	}); err != nil {
		return err
	}

	if removeOldName {
		if err := logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Removing old image tag %s", dockerImage.Image.Name())).DoError(func() error {
			if err := docker.CliRmi(ctx, dockerImage.Image.Name()); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
	}

	dockerImage.Image.SetName(newImageName)

	return nil
}

func (runtime *LocalDockerServerRuntime) RemoveImage(ctx context.Context, img Image) error {
	dockerImage := img.(*DockerImage)

	if err := logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Removing image tag %s", dockerImage.Image.Name())).DoError(func() error {
		if err := docker.CliRmi(ctx, dockerImage.Image.Name()); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (runtime *LocalDockerServerRuntime) PullImageFromRegistry(ctx context.Context, img Image) error {
	dockerImage := img.(*DockerImage)

	if err := dockerImage.Image.Pull(ctx); err != nil {
		return fmt.Errorf("unable to pull image %s: %s", dockerImage.Image.Name(), err)
	}

	if inspect, err := runtime.GetImageInspect(ctx, dockerImage.Image.Name()); err != nil {
		return fmt.Errorf("unable to get inspect of image %s: %s", dockerImage.Image.Name(), err)
	} else {
		dockerImage.Image.SetInspect(inspect)
	}

	return nil
}

func (runtime *LocalDockerServerRuntime) PushImage(ctx context.Context, img Image) error {
	dockerImage := img.(*DockerImage)

	if err := logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Pushing %s", dockerImage.Image.Name())).DoError(func() error {
		return docker.CliPushWithRetries(ctx, dockerImage.Image.Name())
	}); err != nil {
		return err
	}

	return nil
}

// PushBuiltImage is only available for LocalDockerServerRuntime
func (runtime *LocalDockerServerRuntime) PushBuiltImage(ctx context.Context, img Image) error {
	dockerImage := img.(*DockerImage)

	if err := logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Tagging built image by name %s", dockerImage.Image.Name())).DoError(func() error {
		if err := dockerImage.Image.TagBuiltImage(ctx); err != nil {
			return fmt.Errorf("unable to tag built image by name %s: %s", dockerImage.Image.Name(), err)
		}
		return nil
	}); err != nil {
		return err
	}

	if err := logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Pushing %s", dockerImage.Image.Name())).DoError(func() error {
		return dockerImage.Image.Push(ctx)
	}); err != nil {
		return err
	}

	return nil
}

// TagBuiltImageByName is only available for LocalDockerServerRuntime
func (runtime *LocalDockerServerRuntime) TagImageByName(ctx context.Context, img Image) error {
	dockerImage := img.(*DockerImage)

	if dockerImage.Image.GetBuiltId() != "" {
		if err := dockerImage.Image.TagBuiltImage(ctx); err != nil {
			return fmt.Errorf("unable to tag image %s: %s", dockerImage.Image.Name(), err)
		}
	} else {
		if err := runtime.RefreshImageObject(ctx, img); err != nil {
			return err
		}
	}

	return nil
}

func (runtime *LocalDockerServerRuntime) String() string {
	return "local-docker-server"
}

type LocalHostRuntime struct {
	ContainerRuntime // TODO: kaniko-like builds
}

func (runtime *LocalHostRuntime) String() string {
	return "localhost"
}
