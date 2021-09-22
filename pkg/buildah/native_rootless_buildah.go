package buildah

import "context"

type NativeRootlessBuildah struct {
	// TODO: store
}

func NewNativeRootlessBuildah() (*NativeRootlessBuildah, error) {
	return nil, nil // TODO
}

func (buildah *NativeRootlessBuildah) BuildFromDockerfile(ctx context.Context, dockerfile []byte, opts BuildFromDockerfileOpts) (string, error) {
	panic("not implemented")
}

func (buildah *NativeRootlessBuildah) RunCommand(ctx context.Context, container string, command []string, opts RunCommandOpts) error {
	panic("not implemented")
}
