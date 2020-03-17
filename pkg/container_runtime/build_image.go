package container_runtime

type buildImage struct {
	*baseImage
}

func newBuildImage(id string, localDockerServerRuntime *LocalDockerServerRuntime) *buildImage {
	image := &buildImage{}
	image.baseImage = newBaseImage(id, localDockerServerRuntime)
	return image
}
