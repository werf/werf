package common

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/host_cleaning"
)

func RunHostStorageGC(ctx context.Context, cmdData *CmdData) error {
	if os.Getenv("WERF_ENABLE_HOST_STORAGE_GC") != "1" {
		return nil
	}

	opts := host_cleaning.LocalDockerServerGCOptions{
		DryRun:                                false,
		Force:                                 false,
		AllowedVolumeUsagePercentageThreshold: *cmdData.AllowedVolumeUsage,
		DockerServerStoragePath:               *cmdData.DockerServerStoragePath,
	}

	shouldRun, err := host_cleaning.ShouldRunGCForLocalDockerServer(ctx, opts)
	if err != nil {
		return err
	}
	if !shouldRun {
		return nil
	}

	return logboek.Context(ctx).Default().LogProcess("Running host storage GC").DoError(func() error {
		return host_cleaning.RunGCForLocalDockerServer(ctx, opts)
	})
}

func SetupAllowedVolumeUsage(cmdData *CmdData, cmd *cobra.Command) {
	var defaultAllowedVolumeUsage int64
	if v := GetIntEnvVarStrict("WERF_ALLOWED_VOLUME_USAGE"); v != nil {
		defaultAllowedVolumeUsage = *v
	}

	cmdData.AllowedVolumeUsage = new(int64)
	cmd.Flags().Int64VarP(cmdData.AllowedVolumeUsage, "allowed-volume-usage", "", defaultAllowedVolumeUsage, "Set allowed percentage threshold of docker storage volume usage which will cause garbage collection of local docker images (default 80% or $WERF_ALLOWED_VOLUME_USAGE)")
}

func SetupDockerServerStoragePath(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.DockerServerStoragePath = new(string)
	cmd.Flags().StringVarP(cmdData.DockerServerStoragePath, "docker-server-storage-path", "", os.Getenv("WERF_DOCKER_SERVER_STORAGE_PATH"), "Use specified path to the local docker server storage to check docker storage volume usage while performing garbage collection of local docker images (detect local docker server storage path by default or use $WERF_DOCKER_SERVER_STORAGE_PATH)")
}
