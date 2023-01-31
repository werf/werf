package bundles

import (
	"context"

	"github.com/werf/werf/pkg/docker_registry"
)

type CopyOptions struct {
	BundlesRegistryClient                BundlesRegistryClient
	FromRegistryClient, ToRegistryClient docker_registry.Interface
	HelmCompatibleChart                  bool
	RenameChart                          string
}

func Copy(ctx context.Context, fromAddr, toAddr *Addr, opts CopyOptions) error {
	fromBundle := NewBundleAccessor(fromAddr, BundleAccessorOptions{
		BundlesRegistryClient: opts.BundlesRegistryClient,
		RegistryClient:        opts.FromRegistryClient,
	})
	toBundle := NewBundleAccessor(toAddr, BundleAccessorOptions{
		BundlesRegistryClient: opts.BundlesRegistryClient,
		RegistryClient:        opts.ToRegistryClient,
	})

	return fromBundle.CopyTo(ctx, toBundle, copyToOptions{HelmCompatibleChart: opts.HelmCompatibleChart, RenameChart: opts.RenameChart})
}
