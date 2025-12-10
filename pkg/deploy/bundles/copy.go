package bundles

import (
	"context"

	"github.com/werf/3p-helm/pkg/werf/helmopts"
	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/ref"
)

type CopyOptions struct {
	BundlesRegistryClient                BundlesRegistryClient
	FromRegistryClient, ToRegistryClient docker_registry.Interface
	HelmCompatibleChart                  bool
	RenameChart                          string
	HelmOptions                          helmopts.HelmOptions
}

func Copy(ctx context.Context, fromAddr, toAddr *ref.Addr, opts CopyOptions) error {
	fromBundle := NewBundleAccessor(fromAddr, BundleAccessorOptions{
		BundlesRegistryClient: opts.BundlesRegistryClient,
		RegistryClient:        opts.FromRegistryClient,
	})
	toBundle := NewBundleAccessor(toAddr, BundleAccessorOptions{
		BundlesRegistryClient: opts.BundlesRegistryClient,
		RegistryClient:        opts.ToRegistryClient,
	})

	return fromBundle.CopyTo(ctx, toBundle, copyToOptions{HelmCompatibleChart: opts.HelmCompatibleChart, RenameChart: opts.RenameChart, HelmOptions: opts.HelmOptions})
}
