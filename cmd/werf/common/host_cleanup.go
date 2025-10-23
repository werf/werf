package common

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/host_cleaning"
	"github.com/werf/werf/v2/pkg/util/option"
)

func RunAutoHostCleanup(ctx context.Context, cmdData *CmdData, containerBackend container_backend.ContainerBackend) error {
	if *cmdData.DisableAutoHostCleanup {
		return nil
	}

	return host_cleaning.RunAutoHostCleanup(ctx, containerBackend, host_cleaning.AutoHostCleanupOptions{
		HostCleanupOptions: host_cleaning.HostCleanupOptions{
			DryRun: false,
			Force:  false,
			AllowedBackendStorageVolumeUsagePercentage:       cmdData.AllowedBackendStorageVolumeUsage,
			AllowedBackendStorageVolumeUsageMarginPercentage: cmdData.AllowedBackendStorageVolumeUsageMargin,
			AllowedLocalCacheVolumeUsagePercentage:           cmdData.AllowedLocalCacheVolumeUsage,
			AllowedLocalCacheVolumeUsageMarginPercentage:     cmdData.AllowedLocalCacheVolumeUsageMargin,
			BackendStoragePath:                               cmdData.BackendStoragePath,
		},
		TmpDir:      cmdData.TmpDir,
		HomeDir:     cmdData.HomeDir,
		ProjectName: cmdData.ProjectName,
	})
}

func SetupDisableAutoHostCleanup(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.DisableAutoHostCleanup = new(bool)
	cmd.Flags().BoolVarP(cmdData.DisableAutoHostCleanup, "disable-auto-host-cleanup", "", util.GetBoolEnvironmentDefaultFalse("WERF_DISABLE_AUTO_HOST_CLEANUP"), "Disable auto host cleanup procedure in main werf commands like werf-build, werf-converge and other (default disabled or WERF_DISABLE_AUTO_HOST_CLEANUP)")
}

func SetupAllowedBackendStorageVolumeUsage(cmdData *CmdData, cmd *cobra.Command) {
	aliases := []struct {
		ParamName string
		EnvName   string
	}{
		{"allowed-backend-storage-volume-usage", "WERF_ALLOWED_BACKEND_STORAGE_VOLUME_USAGE"},
		{"allowed-docker-storage-volume-usage", "WERF_ALLOWED_DOCKER_STORAGE_VOLUME_USAGE"},
	}

	defaultValUint64 := option.PtrValueOrDefault(util.GetUint64EnvVarStrict(aliases[0].EnvName),
		// keep backward compatibility
		option.PtrValueOrDefault(util.GetUint64EnvVarStrict(aliases[1].EnvName),
			uint64(host_cleaning.DefaultAllowedBackendStorageVolumeUsagePercentage)))

	defaultVal := uint(defaultValUint64)

	if defaultVal > 100 {
		panic(fmt.Sprintf("bad %s value: specify percentage between 0 and 100", aliases[0].EnvName))
	}

	cmdData.AllowedBackendStorageVolumeUsage = new(uint)

	for _, alias := range aliases {
		cmd.Flags().UintVarP(
			cmdData.AllowedBackendStorageVolumeUsage,
			alias.ParamName,
			"",
			defaultVal,
			fmt.Sprintf("Set allowed percentage of backend (Docker or Buildah) storage volume usage which will cause cleanup of least recently used local backend images (default %d%% or $%s)", uint(host_cleaning.DefaultAllowedBackendStorageVolumeUsagePercentage), alias.EnvName),
		)
	}

	if err := cmd.Flags().MarkHidden(aliases[1].ParamName); err != nil {
		panic(err)
	}
}

func SetupAllowedBackendStorageVolumeUsageMargin(cmdData *CmdData, cmd *cobra.Command) {
	aliases := []struct {
		ParamName string
		EnvName   string
	}{
		{"allowed-backend-storage-volume-usage-margin", "WERF_ALLOWED_BACKEND_STORAGE_VOLUME_USAGE_MARGIN"},
		{"allowed-docker-storage-volume-usage-margin", "WERF_ALLOWED_DOCKER_STORAGE_VOLUME_USAGE_MARGIN"},
	}

	defaultValUint64 := option.PtrValueOrDefault(util.GetUint64EnvVarStrict(aliases[0].EnvName),
		// keep backward compatibility
		option.PtrValueOrDefault(util.GetUint64EnvVarStrict(aliases[1].EnvName),
			uint64(host_cleaning.DefaultAllowedBackendStorageVolumeUsageMarginPercentage)))

	defaultVal := uint(defaultValUint64)

	if defaultVal > 100 {
		panic(fmt.Sprintf("bad %s value: specify percentage between 0 and 100", aliases[0].EnvName))
	}

	cmdData.AllowedBackendStorageVolumeUsageMargin = new(uint)

	for _, alias := range aliases {
		cmd.Flags().UintVarP(
			cmdData.AllowedBackendStorageVolumeUsageMargin,
			alias.ParamName,
			"",
			defaultVal,
			fmt.Sprintf("During cleanup of least recently used local backend (Docker or Buildah) images werf would delete images until volume usage becomes below \"allowed-backend-storage-volume-usage - allowed-backend-storage-volume-usage-margin\" level (default %d%% or $%s)", uint(host_cleaning.DefaultAllowedBackendStorageVolumeUsageMarginPercentage), alias.EnvName),
		)
	}

	if err := cmd.Flags().MarkHidden(aliases[1].ParamName); err != nil {
		panic(err)
	}
}

func SetupBackendStoragePath(cmdData *CmdData, cmd *cobra.Command) {
	aliases := []struct {
		ParamName string
		EnvName   string
	}{
		{"backend-storage-path", "WERF_BACKEND_STORAGE_PATH"},
		{"docker-server-storage-path", "WERF_DOCKER_SERVER_STORAGE_PATH"},
	}

	defaultVal := option.ValueOrDefault(os.Getenv(aliases[0].EnvName),
		// keep backward compatibility
		os.Getenv(aliases[1].EnvName))

	cmdData.BackendStoragePath = new(string)

	for _, alias := range aliases {
		cmd.Flags().StringVarP(
			cmdData.BackendStoragePath,
			alias.ParamName,
			"",
			defaultVal,
			fmt.Sprintf("Use specified path to the local backend (Docker or Buildah) storage to check backend storage volume usage while performing garbage collection of local backend images (detect local backend storage path by default or use $%s)", alias.EnvName),
		)
	}

	if err := cmd.Flags().MarkHidden(aliases[1].ParamName); err != nil {
		panic(err)
	}
}

func SetupAllowedLocalCacheVolumeUsage(cmdData *CmdData, cmd *cobra.Command) {
	envVarName := "WERF_ALLOWED_LOCAL_CACHE_VOLUME_USAGE"

	var defaultVal uint
	if v := util.GetUint64EnvVarStrict(envVarName); v != nil {
		defaultVal = uint(*v)
	} else {
		defaultVal = uint(host_cleaning.DefaultAllowedLocalCacheVolumeUsagePercentage)
	}
	if defaultVal > 100 {
		panic(fmt.Sprintf("bad %s value: specify percentage between 0 and 100", envVarName))
	}

	cmdData.AllowedLocalCacheVolumeUsage = new(uint)
	cmd.Flags().UintVarP(cmdData.AllowedLocalCacheVolumeUsage, "allowed-local-cache-volume-usage", "", defaultVal, fmt.Sprintf("Set allowed percentage of local cache (~/.werf/local_cache by default) volume usage which will cause cleanup of least recently used data from the local cache (default %d%% or $%s)", uint(host_cleaning.DefaultAllowedLocalCacheVolumeUsagePercentage), envVarName))
}

func SetupAllowedLocalCacheVolumeUsageMargin(cmdData *CmdData, cmd *cobra.Command) {
	envVarName := "WERF_ALLOWED_LOCAL_CACHE_VOLUME_USAGE_MARGIN"

	var defaultVal uint
	if v := util.GetUint64EnvVarStrict(envVarName); v != nil {
		defaultVal = uint(*v)
	} else {
		defaultVal = uint(host_cleaning.DefaultAllowedLocalCacheVolumeUsageMarginPercentage)
	}
	if defaultVal > 100 {
		panic(fmt.Sprintf("bad %s value: specify percentage between 0 and 100", envVarName))
	}

	cmdData.AllowedLocalCacheVolumeUsageMargin = new(uint)
	cmd.Flags().UintVarP(cmdData.AllowedLocalCacheVolumeUsageMargin, "allowed-local-cache-volume-usage-margin", "", defaultVal, fmt.Sprintf("During cleanup of local cache werf would delete local cache data until volume usage becomes below \"allowed-local-cache-volume-usage - allowed-local-cache-volume-usage-margin\" level (default %d%% or $%s)", uint(host_cleaning.DefaultAllowedLocalCacheVolumeUsageMarginPercentage), envVarName))
}

func SetupProjectName(cmdData *CmdData, cmd *cobra.Command, visible bool) {
	const name = "project-name"

	cmdData.ProjectName = new(string)
	cmd.Flags().StringVarP(cmdData.ProjectName, name, "N", os.Getenv("WERF_PROJECT_NAME"), "Set a specific project name (default $WERF_PROJECT_NAME)")

	if !visible {
		if err := cmd.Flags().MarkHidden(name); err != nil {
			panic(err)
		}
	}
}
