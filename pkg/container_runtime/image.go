package container_runtime

type Image interface{}

type DockerImage struct {
	Image ImageInterface
}
