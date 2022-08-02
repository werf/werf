package common

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/host_cleaning"
	"github.com/werf/werf/pkg/util"
)

func RunAutoHostCleanup(ctx context.Context, cmdData *CmdData, containerBackend container_backend.ContainerBackend) error {
	if *cmdData.DisableAutoHostCleanup {
		return nil
	}

	cleanupDockerServer := false
	if _, match := containerBackend.(*container_backend.DockerServerBackend); match {
		cleanupDockerServer = true
	}

	if *cmdData.AllowedDockerStorageVolumeUsageMargin >= *cmdData.AllowedDockerStorageVolumeUsage {
		return fmt.Errorf("incompatible params --allowed-docker-storage-volume-usage=%d and --allowed-docker-storage-volume-usage-margin=%d: margin percentage should be less than allowed volume usage level percentage", *cmdData.AllowedDockerStorageVolumeUsage, *cmdData.AllowedDockerStorageVolumeUsageMargin)
	}

	if *cmdData.AllowedLocalCacheVolumeUsageMargin >= *cmdData.AllowedLocalCacheVolumeUsage {
		return fmt.Errorf("incompatible params --allowed-local-cache-volume-usage=%d and --allowed-local-cache-volume-usage-margin=%d: margin percentage should be less than allowed volume usage level percentage", *cmdData.AllowedLocalCacheVolumeUsage, *cmdData.AllowedLocalCacheVolumeUsageMargin)
	}

	return host_cleaning.RunAutoHostCleanup(ctx, host_cleaning.AutoHostCleanupOptions{
		HostCleanupOptions: host_cleaning.HostCleanupOptions{
			DryRun:              false,
			Force:               false,
			CleanupDockerServer: cleanupDockerServer,
			AllowedDockerStorageVolumeUsagePercentage:       cmdData.AllowedDockerStorageVolumeUsage,
			AllowedDockerStorageVolumeUsageMarginPercentage: cmdData.AllowedDockerStorageVolumeUsageMargin,
			AllowedLocalCacheVolumeUsagePercentage:          cmdData.AllowedLocalCacheVolumeUsage,
			AllowedLocalCacheVolumeUsageMarginPercentage:    cmdData.AllowedLocalCacheVolumeUsageMargin,
			DockerServerStoragePath:                         cmdData.DockerServerStoragePath,
		},
	})
}

func SetupDisableAutoHostCleanup(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.DisableAutoHostCleanup = new(bool)
	cmd.Flags().BoolVarP(cmdData.DisableAutoHostCleanup, "disable-auto-host-cleanup", "", util.GetBoolEnvironmentDefaultFalse("WERF_DISABLE_AUTO_HOST_CLEANUP"), "Disable auto host cleanup procedure in main werf commands like werf-build, werf-converge and other (default disabled or WERF_DISABLE_AUTO_HOST_CLEANUP)")
}

func SetupAllowedDockerStorageVolumeUsage(cmdData *CmdData, cmd *cobra.Command) {
	envVarName := "WERF_ALLOWED_DOCKER_STORAGE_VOLUME_USAGE"

	var defaultVal uint
	if v := GetUint64EnvVarStrict(envVarName); v != nil {
		defaultVal = uint(*v)
	} else {
		defaultVal = uint(host_cleaning.DefaultAllowedDockerStorageVolumeUsagePercentage)
	}
	if defaultVal > 100 {
		TerminateWithError(fmt.Sprintf("bad %s value: specify percentage between 0 and 100", envVarName), 1)
	}

	cmdData.AllowedDockerStorageVolumeUsage = new(uint)
	cmd.Flags().UintVarP(cmdData.AllowedDockerStorageVolumeUsage, "allowed-docker-storage-volume-usage", "", defaultVal, fmt.Sprintf("Set allowed percentage of docker storage volume usage which will cause cleanup of least recently used local docker images (default %d%% or $%s)", uint(host_cleaning.DefaultAllowedDockerStorageVolumeUsagePercentage), envVarName))
}

func SetupAllowedDockerStorageVolumeUsageMargin(cmdData *CmdData, cmd *cobra.Command) {
	envVarName := "WERF_ALLOWED_DOCKER_STORAGE_VOLUME_USAGE_MARGIN"

	var defaultVal uint
	if v := GetUint64EnvVarStrict(envVarName); v != nil {
		defaultVal = uint(*v)
	} else {
		defaultVal = uint(host_cleaning.DefaultAllowedDockerStorageVolumeUsageMarginPercentage)
	}
	if defaultVal > 100 {
		TerminateWithError(fmt.Sprintf("bad %s value: specify percentage between 0 and 100", envVarName), 1)
	}

	cmdData.AllowedDockerStorageVolumeUsageMargin = new(uint)
	cmd.Flags().UintVarP(cmdData.AllowedDockerStorageVolumeUsageMargin, "allowed-docker-storage-volume-usage-margin", "", defaultVal, fmt.Sprintf("During cleanup of least recently used local docker images werf would delete images until volume usage becomes below \"allowed-docker-storage-volume-usage - allowed-docker-storage-volume-usage-margin\" level (default %d%% or $%s)", uint(host_cleaning.DefaultAllowedDockerStorageVolumeUsageMarginPercentage), envVarName))
}

func SetupDockerServerStoragePath(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.DockerServerStoragePath = new(string)
	cmd.Flags().StringVarP(cmdData.DockerServerStoragePath, "docker-server-storage-path", "", os.Getenv("WERF_DOCKER_SERVER_STORAGE_PATH"), "Use specified path to the local docker server storage to check docker storage volume usage while performing garbage collection of local docker images (detect local docker server storage path by default or use $WERF_DOCKER_SERVER_STORAGE_PATH)")
}

func SetupAllowedLocalCacheVolumeUsage(cmdData *CmdData, cmd *cobra.Command) {
	envVarName := "WERF_ALLOWED_LOCAL_CACHE_VOLUME_USAGE"

	var defaultVal uint
	if v := GetUint64EnvVarStrict(envVarName); v != nil {
		defaultVal = uint(*v)
	} else {
		defaultVal = uint(host_cleaning.DefaultAllowedLocalCacheVolumeUsagePercentage)
	}
	if defaultVal > 100 {
		TerminateWithError(fmt.Sprintf("bad %s value: specify percentage between 0 and 100", envVarName), 1)
	}

	cmdData.AllowedLocalCacheVolumeUsage = new(uint)
	cmd.Flags().UintVarP(cmdData.AllowedLocalCacheVolumeUsage, "allowed-local-cache-volume-usage", "", defaultVal, fmt.Sprintf("Set allowed percentage of local cache (~/.werf/local_cache by default) volume usage which will cause cleanup of least recently used data from the local cache (default %d%% or $%s)", uint(host_cleaning.DefaultAllowedLocalCacheVolumeUsagePercentage), envVarName))
}

func SetupAllowedLocalCacheVolumeUsageMargin(cmdData *CmdData, cmd *cobra.Command) {
	envVarName := "WERF_ALLOWED_LOCAL_CACHE_VOLUME_USAGE_MARGIN"

	var defaultVal uint
	if v := GetUint64EnvVarStrict(envVarName); v != nil {
		defaultVal = uint(*v)
	} else {
		defaultVal = uint(host_cleaning.DefaultAllowedLocalCacheVolumeUsageMarginPercentage)
	}
	if defaultVal > 100 {
		TerminateWithError(fmt.Sprintf("bad %s value: specify percentage between 0 and 100", envVarName), 1)
	}

	cmdData.AllowedLocalCacheVolumeUsageMargin = new(uint)
	cmd.Flags().UintVarP(cmdData.AllowedLocalCacheVolumeUsageMargin, "allowed-local-cache-volume-usage-margin", "", defaultVal, fmt.Sprintf("During cleanup of least recently used local docker images werf would delete images until volume usage becomes below \"allowed-docker-storage-volume-usage - allowed-docker-storage-volume-usage-margin\" level (default %d%% or $%s)", uint(host_cleaning.DefaultAllowedLocalCacheVolumeUsageMarginPercentage), envVarName))
}
