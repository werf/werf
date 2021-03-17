package common

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/host_cleaning"
)

func RunHostStorageGC(ctx context.Context, cmdData *CmdData) error {
	if os.Getenv("WERF_ENABLE_HOST_STORAGE_GC") != "1" {
		return nil
	}

	opts := host_cleaning.HostCleanupOptions{
		DryRun:                             false,
		Force:                              false,
		AllowedVolumeUsagePercentage:       cmdData.AllowedVolumeUsage,
		AllowedVolumeUsageMarginPercentage: cmdData.AllowedVolumeUsageMargin,
		DockerServerStoragePath:            *cmdData.DockerServerStoragePath,
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
	envVarName := "WERF_ALLOWED_VOLUME_USAGE"
	var defaultVal uint
	if v := GetUint64EnvVarStrict(envVarName); v != nil {
		defaultVal = uint(*v)
	} else {
		defaultVal = uint(host_cleaning.DefaultAllowedVolumeUsagePercentage)
	}
	if defaultVal > 100 {
		TerminateWithError(fmt.Sprintf("bad %s value: specify percentage between 0 and 100", envVarName), 1)
	}

	cmdData.AllowedVolumeUsage = new(uint)
	cmd.Flags().UintVarP(cmdData.AllowedVolumeUsage, "allowed-volume-usage", "", defaultVal, fmt.Sprintf("Set allowed percentage of docker storage volume usage which will cause garbage collection of local docker images (default %d%% or $WERF_ALLOWED_VOLUME_USAGE)", uint(host_cleaning.DefaultAllowedVolumeUsagePercentage)))
}

func SetupAllowedVolumeUsageMargin(cmdData *CmdData, cmd *cobra.Command) {
	envVarName := "WERF_ALLOWED_VOLUME_USAGE_MARGIN"
	var defaultVal uint
	if v := GetUint64EnvVarStrict(envVarName); v != nil {
		defaultVal = uint(*v)
	} else {
		defaultVal = uint(host_cleaning.DefaultAllowedVolumeUsageMarginPercentage)
	}
	if defaultVal > 100 {
		TerminateWithError(fmt.Sprintf("bad %s value: specify percentage between 0 and 100", envVarName), 1)
	}

	cmdData.AllowedVolumeUsageMargin = new(uint)
	cmd.Flags().UintVarP(cmdData.AllowedVolumeUsageMargin, "allowed-volume-usage-margin", "", defaultVal, fmt.Sprintf("During garbage collection werf would delete images until volume usage becomes below \"allowed-volume-usage - allowed-volume-usage-margin\" level (default %d%% or $WERF_ALLOWED_VOLUME_USAGE_MARGIN)", uint(host_cleaning.DefaultAllowedVolumeUsageMarginPercentage)))
}

func SetupDockerServerStoragePath(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.DockerServerStoragePath = new(string)
	cmd.Flags().StringVarP(cmdData.DockerServerStoragePath, "docker-server-storage-path", "", os.Getenv("WERF_DOCKER_SERVER_STORAGE_PATH"), "Use specified path to the local docker server storage to check docker storage volume usage while performing garbage collection of local docker images (detect local docker server storage path by default or use $WERF_DOCKER_SERVER_STORAGE_PATH)")
}
