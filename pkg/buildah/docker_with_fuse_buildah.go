package buildah

import "context"

type DockerWithFuseBuildah struct {
	// TODO: store
}

func NewDockerWithFuseBuildah() (*DockerWithFuseBuildah, error) {
	return nil, nil // TODO
}

func (buildah *DockerWithFuseBuildah) BuildFromDockerfile(ctx context.Context, dockerfile []byte, opts BuildFromDockerfileOpts) (string, error) {
	panic("not implemented")
}

func (buildah *DockerWithFuseBuildah) RunCommand(ctx context.Context, container string, command []string, opts RunCommandOpts) error {
	panic("not implemented")
}
