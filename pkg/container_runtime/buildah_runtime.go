package container_runtime

import "context"

type BuildahRuntime struct {
}

type BuildahImage struct {
	Image LegacyImageInterface
}

func (runtime *BuildahRuntime) RefreshImageObject(ctx context.Context, img Image) error {
	panic("not implemented")
}

func (runtime *BuildahRuntime) PullImageFromRegistry(ctx context.Context, img Image) error {
	panic("not implemented")
}

func (runtime *BuildahRuntime) RenameImage(ctx context.Context, img Image, newImageName string, removeOldName bool) error {
	panic("not implemented")
}

func (runtime *BuildahRuntime) RemoveImage(ctx context.Context, img Image) error {
	panic("not implemented")
}

func (runtime *BuildahRuntime) String() string {
	return "buildah"
}
