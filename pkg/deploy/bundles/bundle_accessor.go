package bundles

import (
	"context"
	"fmt"

	"helm.sh/helm/v3/pkg/chart"

	"github.com/werf/werf/pkg/docker_registry"
)

type copyToOptions struct {
	HelmCompatibleChart bool
	RenameChart         string
}

type BundleAccessor interface {
	ReadChart(ctx context.Context) (*chart.Chart, error)
	WriteChart(ctx context.Context, ch *chart.Chart) error

	CopyTo(ctx context.Context, to BundleAccessor, opts copyToOptions) error
	CopyFromArchive(ctx context.Context, fromArchive *BundleArchive, opts copyToOptions) error
	CopyFromRemote(ctx context.Context, fromRemote *RemoteBundle, opts copyToOptions) error
}

type BundleAccessorOptions struct {
	BundlesRegistryClient BundlesRegistryClient
	RegistryClient        docker_registry.Interface
}

func NewBundleAccessor(addr *Addr, opts BundleAccessorOptions) BundleAccessor {
	switch {
	case addr.RegistryAddress != nil:
		return NewRemoteBundle(addr.RegistryAddress, opts.BundlesRegistryClient, opts.RegistryClient)
	case addr.ArchiveAddress != nil:
		return NewBundleArchive(NewBundleArchiveFileReader(addr.ArchiveAddress.Path), NewBundleArchiveFileWriter(addr.ArchiveAddress.Path))
	default:
		panic(fmt.Sprintf("invalid address given %#v", addr))
	}
}
