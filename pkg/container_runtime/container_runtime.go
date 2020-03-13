package container_runtime

import "fmt"

type ContainerRuntime interface {
	RefreshImageObject(image Image) error

	String() string
}

type LocalDockerServerRuntime struct{}

// PushBuiltImage is only available for LocalDockerServerRuntime
func (runtime *LocalDockerServerRuntime) ExportBuiltImage(image Image) error {
	dockerImage := image.(*DockerImage)

	if err := dockerImage.Image.Export(dockerImage.Image.Name()); err != nil {
		return fmt.Errorf("unable to export image %s: %s", dockerImage.Image.Name(), err)
	}
	if err := dockerImage.Image.SyncDockerState(); err != nil {
		return fmt.Errorf("unable to sync docker state of image %s: %s", dockerImage.Image.Name(), err)
	}

	return nil
}

// StoreBuiltImage is only available for LocalDockerServerRuntime
func (runtime *LocalDockerServerRuntime) TagBuiltImageByName(image Image) error {
	dockerImage := image.(*DockerImage)

	if err := dockerImage.Image.TagBuiltImage(dockerImage.Image.Name()); err != nil {
		return fmt.Errorf("unable to tag image %s: %s", dockerImage.Image.Name(), err)
	}
	if err := dockerImage.Image.SyncDockerState(); err != nil {
		return fmt.Errorf("unable to sync docker state of image %s: %s", dockerImage.Image.Name(), err)
	}

	return nil
}

func (runtime *LocalDockerServerRuntime) RefreshImageObject(image Image) error {
	dockerImage := image.(*DockerImage)
	return dockerImage.Image.SyncDockerState()
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
