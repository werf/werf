package stages

import (
	"context"

	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/build"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/giterminism_manager"
	"github.com/werf/werf/v2/pkg/ref"
)

type CopyOptions struct {
	AllStages          bool
	ProjectName        string
	BaseTmpDir         string
	ContainerBackend   container_backend.ContainerBackend
	CommonCmdData      *common.CmdData
	WerfConfig         *config.WerfConfig
	GiterminismManager *giterminism_manager.Manager
	BuildOptions       build.BuildOptions
}

func Copy(ctx context.Context, fromAddr, toAddr *ref.Addr, opts CopyOptions) error {
	from, err := NewStorageSrcAccessor(ctx, fromAddr, StorageAccessorOptions{
		InsecureRegistry:             *opts.CommonCmdData.InsecureRegistry,
		SkipTlsVerifyRegistry:        *opts.CommonCmdData.SkipTlsVerifyRegistry,
		DisableCleanup:               opts.WerfConfig.Meta.Cleanup.DisableCleanup,
		DisableGitHistoryBasedPolicy: opts.WerfConfig.Meta.Cleanup.DisableGitHistoryBasedPolicy,
		AllStages:                    opts.AllStages,
		BaseTmpDir:                   opts.BaseTmpDir,
		ContainerBackend:             opts.ContainerBackend,
		CommonCmdData:                opts.CommonCmdData,
		WerfConfig:                   opts.WerfConfig,
		GiterminismManager:           opts.GiterminismManager,
	})
	if err != nil {
		return err
	}

	to, err := NewStorageDstAccessor(ctx, toAddr, StorageAccessorOptions{
		InsecureRegistry:             *opts.CommonCmdData.InsecureRegistry,
		SkipTlsVerifyRegistry:        *opts.CommonCmdData.SkipTlsVerifyRegistry,
		DisableCleanup:               opts.WerfConfig.Meta.Cleanup.DisableCleanup,
		DisableGitHistoryBasedPolicy: opts.WerfConfig.Meta.Cleanup.DisableGitHistoryBasedPolicy,
		AllStages:                    opts.AllStages,
		BaseTmpDir:                   opts.BaseTmpDir,
		ContainerBackend:             opts.ContainerBackend,
		CommonCmdData:                opts.CommonCmdData,
		WerfConfig:                   opts.WerfConfig,
		GiterminismManager:           opts.GiterminismManager,
	})
	if err != nil {
		return err
	}

	return from.CopyTo(ctx, to, copyToOptions{
		All:          opts.AllStages,
		ProjectName:  opts.ProjectName,
		BuildOptions: opts.BuildOptions,
	})
}
