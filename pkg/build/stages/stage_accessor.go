package stages

import (
	"context"
	"fmt"

	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/build"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/giterminism_manager"
	"github.com/werf/werf/v2/pkg/ref"
)

type copyToOptions struct {
	All         bool
	ProjectName string
}

type StorageAccessor interface {
	CopyTo(ctx context.Context, to StorageAccessor, opts copyToOptions) error
	CopyFromArchive(ctx context.Context, fromArchive *ArchiveStorage, opts copyToOptions) error
	CopyFromRemote(ctx context.Context, fromRemote *RemoteStorage, opts copyToOptions) error
}

type StorageAccessorOptions struct {
	InsecureRegistry             bool
	SkipTlsVerifyRegistry        bool
	DisableCleanup               bool
	DisableGitHistoryBasedPolicy bool
	AllStages                    bool
	BaseTmpDir                   string
	ContainerBackend             container_backend.ContainerBackend
	CommonCmdData                *common.CmdData
	WerfConfig                   *config.WerfConfig
	GiterminismManager           *giterminism_manager.Manager
}

func NewStorageSrcAccessor(ctx context.Context, addr *ref.Addr, opts StorageAccessorOptions) (StorageAccessor, error) {
	switch {
	case addr.RegistryAddress != nil:
		return createSrcRemoteStorage(ctx, addr, opts)
	case addr.ArchiveAddress != nil:
		return createArchiveStorage(addr.ArchiveAddress)
	default:
		panic(fmt.Sprintf("invalid address given %#v", addr))
	}
}

func NewStorageDstAccessor(ctx context.Context, addr *ref.Addr, opts StorageAccessorOptions) (StorageAccessor, error) {
	switch {
	case addr.RegistryAddress != nil:
		return createDstRemoteStorage(ctx, addr, opts.InsecureRegistry, opts.SkipTlsVerifyRegistry, opts.AllStages)
	case addr.ArchiveAddress != nil:
		return createArchiveStorage(addr.ArchiveAddress)
	default:
		panic(fmt.Sprintf("invalid address given %#v", addr))
	}
}

func createSrcRemoteStorage(ctx context.Context, addr *ref.Addr, opts StorageAccessorOptions) (*RemoteStorage, error) {
	opts.CommonCmdData.Repo.Address = &addr.RegistryAddress.Repo //FIXME выдумать что-нить симпатичнее

	storageManager, err := common.NewStorageManager(ctx, &common.NewStorageManagerConfig{
		ProjectName:                    opts.WerfConfig.Meta.Project,
		ContainerBackend:               opts.ContainerBackend,
		CmdData:                        opts.CommonCmdData,
		CleanupDisabled:                opts.DisableCleanup,
		GitHistoryBasedCleanupDisabled: opts.DisableGitHistoryBasedPolicy,
	})
	if err != nil {
		return nil, err
	}

	dockerRegistry, err := common.CreateDockerRegistry(ctx, addr.Repo, opts.InsecureRegistry, opts.SkipTlsVerifyRegistry)
	if err != nil {
		return nil, err
	}

	conveyorWithRetry := build.NewConveyorWithRetryWrapper(opts.WerfConfig, opts.GiterminismManager, opts.GiterminismManager.ProjectDir(), opts.BaseTmpDir, opts.ContainerBackend, storageManager, storageManager.StorageLockManager, build.ConveyorOptions{})
	return NewRemoteStorage(addr.RegistryAddress, dockerRegistry, storageManager, conveyorWithRetry, opts.AllStages), nil
}

func createDstRemoteStorage(ctx context.Context, addr *ref.Addr, insecureRegistry, skipTlsVerifyRegistry, allStages bool) (*RemoteStorage, error) {
	dockerRegistry, err := common.CreateDockerRegistry(ctx, addr.Repo, insecureRegistry, skipTlsVerifyRegistry)
	if err != nil {
		return nil, err
	}

	return NewRemoteStorage(addr.RegistryAddress, dockerRegistry, nil, nil, allStages), nil
}

func createArchiveStorage(addr *ref.ArchiveAddress) (*ArchiveStorage, error) {
	return NewArchiveStorage(NewArchiveStorageFileReader(addr.Path), NewArchiveStorageFileWriter(addr.Path)), nil
}
