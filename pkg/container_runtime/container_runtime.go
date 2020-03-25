package container_runtime

import (
	"fmt"

	"github.com/flant/logboek"

	"github.com/docker/docker/api/types"
	"github.com/flant/werf/pkg/docker"

	"github.com/docker/docker/client"
)

type ContainerRuntime interface {
	RefreshImageObject(img Image) error
	PullImageFromRegistry(img Image) error
	String() string
}

type LocalDockerServerRuntime struct{}

// GetImageInspect only avaiable for LocalDockerServerRuntime
func (runtime *LocalDockerServerRuntime) GetImageInspect(ref string) (*types.ImageInspect, error) {
	inspect, err := docker.ImageInspect(ref)
	if client.IsErrNotFound(err) {
		return nil, nil
	}
	return inspect, err
}

func (runtime *LocalDockerServerRuntime) RefreshImageObject(img Image) error {
	dockerImage := img.(*DockerImage)

	if inspect, err := runtime.GetImageInspect(dockerImage.Image.Name()); err != nil {
		return err
	} else {
		dockerImage.Image.SetInspect(inspect)
	}
	return nil
}

func (runtime *LocalDockerServerRuntime) PullImageFromRegistry(img Image) error {
	dockerImage := img.(*DockerImage)

	if err := dockerImage.Image.Pull(); err != nil {
		return fmt.Errorf("unable to export image %s: %s", dockerImage.Image.Name(), err)
	}

	if inspect, err := runtime.GetImageInspect(dockerImage.Image.Name()); err != nil {
		return fmt.Errorf("unable to get inspect of image %s: %s", dockerImage.Image.Name(), err)
	} else {
		dockerImage.Image.SetInspect(inspect)
	}

	return nil
}

// PushBuiltImage is only available for LocalDockerServerRuntime
func (runtime *LocalDockerServerRuntime) PushBuiltImage(img Image) error {
	dockerImage := img.(*DockerImage)

	if err := logboek.Info.LogProcess(fmt.Sprintf("Tagging built image by name %s", dockerImage.Image.Name()), logboek.LevelLogProcessOptions{}, func() error {
		if err := dockerImage.Image.TagBuiltImage(dockerImage.Image.Name()); err != nil {
			return fmt.Errorf("unable to tag built image by name %s: %s", dockerImage.Image.Name(), err)
		}
		return nil
	}); err != nil {
		return err
	}

	if err := logboek.Info.LogProcess(fmt.Sprintf("Pushing %s", dockerImage.Image.Name()), logboek.LevelLogProcessOptions{}, func() error {
		return docker.CliPushWithRetries(dockerImage.Image.Name())
	}); err != nil {
		return err
	}

	return nil
}

// TagBuiltImageByName is only available for LocalDockerServerRuntime
func (runtime *LocalDockerServerRuntime) TagBuiltImageByName(img Image) error {
	dockerImage := img.(*DockerImage)

	if err := dockerImage.Image.TagBuiltImage(dockerImage.Image.Name()); err != nil {
		return fmt.Errorf("unable to tag image %s: %s", dockerImage.Image.Name(), err)
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
