package bundles

import (
	"context"
	"fmt"

	nelmcommon "github.com/werf/nelm/pkg/common"
	chart "github.com/werf/nelm/pkg/helm/pkg/chart/v2"
	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/ref"
)

type copyToOptions struct {
	HelmCompatibleChart bool
	RenameChart         string
	HelmOptions         nelmcommon.HelmOptions
}

type BundleAccessor interface {
	ReadChart(ctx context.Context, opts nelmcommon.HelmOptions) (*chart.Chart, error)
	WriteChart(ctx context.Context, ch *chart.Chart, opts nelmcommon.HelmOptions) error

	CopyTo(ctx context.Context, to BundleAccessor, opts copyToOptions) error
	CopyFromArchive(ctx context.Context, fromArchive *BundleArchive, opts copyToOptions) error
	CopyFromRemote(ctx context.Context, fromRemote *RemoteBundle, opts copyToOptions) error
}

type BundleAccessorOptions struct {
	BundlesRegistryClient BundlesRegistryClient
	RegistryClient        docker_registry.Interface
}

func NewBundleAccessor(addr *ref.Addr, opts BundleAccessorOptions) BundleAccessor {
	switch {
	case addr.RegistryAddress != nil:
		return NewRemoteBundle(addr.RegistryAddress, opts.BundlesRegistryClient, opts.RegistryClient)
	case addr.ArchiveAddress != nil:
		return NewBundleArchive(NewBundleArchiveFileReader(addr.ArchiveAddress.Path), NewBundleArchiveFileWriter(addr.ArchiveAddress.Path))
	default:
		panic(fmt.Sprintf("invalid address given %#v", addr))
	}
}
