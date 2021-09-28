package container_runtime

import (
	"context"
)

type ContainerRuntime interface {
	RefreshImageObject(ctx context.Context, img Image) error

	PullImageFromRegistry(ctx context.Context, img Image) error

	RenameImage(ctx context.Context, img Image, newImageName string, removeOldName bool) error
	RemoveImage(ctx context.Context, img Image) error

	String() string
}

type Image interface {
}
