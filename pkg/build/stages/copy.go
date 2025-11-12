package stages

import (
	"context"

	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/ref"
)

type CopyOptions struct {
	ProjectName      string
	ContainerBackend container_backend.ContainerBackend
	CommonCmdData    *common.CmdData
	WerfConfig       *config.WerfConfig
	All              bool
}

func Copy(ctx context.Context, fromAddr, toAddr *ref.Addr, opts CopyOptions) error {
	from, err := NewStorageSrcAccessor(ctx, fromAddr, StorageAccessorOptions{
		RegistryOptions: RegistryStorageOptions{
			ProjectName:                  opts.ProjectName,
			ContainerBackend:             opts.ContainerBackend,
			CommonCmdData:                opts.CommonCmdData,
			DisableCleanup:               opts.WerfConfig.Meta.Cleanup.DisableCleanup,
			DisableGitHistoryBasedPolicy: opts.WerfConfig.Meta.Cleanup.DisableGitHistoryBasedPolicy,
		},
	})
	if err != nil {
		return err
	}

	to, err := NewStorageDstAccessor(ctx, toAddr, StorageAccessorOptions{
		RegistryOptions: RegistryStorageOptions{
			ProjectName:                  opts.ProjectName,
			ContainerBackend:             opts.ContainerBackend,
			CommonCmdData:                opts.CommonCmdData,
			DisableCleanup:               opts.WerfConfig.Meta.Cleanup.DisableCleanup,
			DisableGitHistoryBasedPolicy: opts.WerfConfig.Meta.Cleanup.DisableGitHistoryBasedPolicy,
		},
	})
	if err != nil {
		return err
	}

	return from.CopyTo(ctx, to, copyToOptions{
		All:         opts.All,
		ProjectName: opts.ProjectName,
	})
}
