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
	"github.com/werf/werf/v2/cmd/werf/common/units"
)

func RunAutoHostCleanup(ctx context.Context, cmdData *CmdData, containerBackend container_backend.ContainerBackend) error {
	if *cmdData.DisableAutoHostCleanup {
		return nil
	}

	return host_cleaning.RunAutoHostCleanup(ctx, containerBackend, host_cleaning.AutoHostCleanupOptions{
		HostCleanupOptions: host_cleaning.HostCleanupOptions{
			DryRun:                                 false,
			Force:                                  false,
			AllowedBackendStorageVolumeUsage:       cmdData.AllowedBackendStorageVolumeUsage,
			AllowedBackendStorageVolumeUsageMargin: cmdData.AllowedBackendStorageVolumeUsageMargin,
			AllowedLocalCacheVolumeUsage:           cmdData.AllowedLocalCacheVolumeUsage,
			AllowedLocalCacheVolumeUsageMargin:     cmdData.AllowedLocalCacheVolumeUsageMargin,
			BackendStoragePath:                     cmdData.BackendStoragePath,
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

	defaultValStr := option.ValueOrDefault(os.Getenv(aliases[0].EnvName),
		option.ValueOrDefault(os.Getenv(aliases[1].EnvName),
			fmt.Sprintf("%d", host_cleaning.DefaultAllowedBackendStorageVolumeUsagePercentage)))

	cmdData.AllowedBackendStorageVolumeUsage = new(units.UnitValue)
	if err := cmdData.AllowedBackendStorageVolumeUsage.Set(defaultValStr); err != nil {
		panic(fmt.Errorf("invalid default value for %s: %w", aliases[0].ParamName, err))
	}

	for _, alias := range aliases {
		cmd.Flags().VarP(
			cmdData.AllowedBackendStorageVolumeUsage,
			alias.ParamName,
			"",
			fmt.Sprintf("Set allowed percentage or absolute value (e.g. 10GB) of backend (Docker or Buildah) storage volume usage which will cause cleanup of least recently used local backend images (default %d%% or $%s)", host_cleaning.DefaultAllowedBackendStorageVolumeUsagePercentage, alias.EnvName),
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

	defaultValStr := option.ValueOrDefault(os.Getenv(aliases[0].EnvName),
		option.ValueOrDefault(os.Getenv(aliases[1].EnvName),
			fmt.Sprintf("%d", host_cleaning.DefaultAllowedBackendStorageVolumeUsageMarginPercentage)))

	cmdData.AllowedBackendStorageVolumeUsageMargin = new(units.UnitValue)
	if err := cmdData.AllowedBackendStorageVolumeUsageMargin.Set(defaultValStr); err != nil {
		panic(fmt.Errorf("invalid default value for %s: %w", aliases[0].ParamName, err))
	}

	for _, alias := range aliases {
		cmd.Flags().VarP(
			cmdData.AllowedBackendStorageVolumeUsageMargin,
			alias.ParamName,
			"",
			fmt.Sprintf("During cleanup of least recently used local backend (Docker or Buildah) images werf would delete images until volume usage becomes below \"allowed-backend-storage-volume-usage - allowed-backend-storage-volume-usage-margin\" level (default %d%% or $%s)", host_cleaning.DefaultAllowedBackendStorageVolumeUsageMarginPercentage, alias.EnvName),
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

	defaultValStr := option.ValueOrDefault(os.Getenv(envVarName),
		fmt.Sprintf("%d", host_cleaning.DefaultAllowedLocalCacheVolumeUsagePercentage))

	cmdData.AllowedLocalCacheVolumeUsage = new(units.UnitValue)
	if err := cmdData.AllowedLocalCacheVolumeUsage.Set(defaultValStr); err != nil {
		panic(fmt.Errorf("invalid default value for allowed-local-cache-volume-usage: %w", err))
	}

	cmd.Flags().VarP(cmdData.AllowedLocalCacheVolumeUsage, "allowed-local-cache-volume-usage", "", fmt.Sprintf("Set allowed percentage or absolute value (e.g. 10GB) of local cache (~/.werf/local_cache by default) volume usage which will cause cleanup of least recently used data from the local cache (default %d%% or $%s)", host_cleaning.DefaultAllowedLocalCacheVolumeUsagePercentage, envVarName))
}

func SetupAllowedLocalCacheVolumeUsageMargin(cmdData *CmdData, cmd *cobra.Command) {
	envVarName := "WERF_ALLOWED_LOCAL_CACHE_VOLUME_USAGE_MARGIN"

	defaultValStr := option.ValueOrDefault(os.Getenv(envVarName),
		fmt.Sprintf("%d", host_cleaning.DefaultAllowedLocalCacheVolumeUsageMarginPercentage))

	cmdData.AllowedLocalCacheVolumeUsageMargin = new(units.UnitValue)
	if err := cmdData.AllowedLocalCacheVolumeUsageMargin.Set(defaultValStr); err != nil {
		panic(fmt.Errorf("invalid default value for allowed-local-cache-volume-usage-margin: %w", err))
	}

	cmd.Flags().VarP(cmdData.AllowedLocalCacheVolumeUsageMargin, "allowed-local-cache-volume-usage-margin", "", fmt.Sprintf("During cleanup of local cache werf would delete local cache data until volume usage becomes below \"allowed-local-cache-volume-usage - allowed-local-cache-volume-usage-margin\" level (default %d%% or $%s)", host_cleaning.DefaultAllowedLocalCacheVolumeUsageMarginPercentage, envVarName))
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
