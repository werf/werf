package common

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/werf/common-go/pkg/util"
	thresholdpkg "github.com/werf/werf/v2/pkg/cleaning/threshold"
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
			AllowedBackendStorageVolumeUsageThreshold:       cmdData.AllowedBackendStorageVolumeUsage,
			AllowedBackendStorageVolumeUsageMarginThreshold: cmdData.AllowedBackendStorageVolumeUsageMargin,
			AllowedBackendStorageVolumeUsageMarginExplicit:  cmdData.AllowedBackendStorageVolumeUsageMarginExplicit != nil && *cmdData.AllowedBackendStorageVolumeUsageMarginExplicit,
			AllowedLocalCacheVolumeUsageThreshold:           cmdData.AllowedLocalCacheVolumeUsage,
			AllowedLocalCacheVolumeUsageMarginThreshold:     cmdData.AllowedLocalCacheVolumeUsageMargin,
			AllowedLocalCacheVolumeUsageMarginExplicit:      cmdData.AllowedLocalCacheVolumeUsageMarginExplicit != nil && *cmdData.AllowedLocalCacheVolumeUsageMarginExplicit,
			BackendStoragePath:                              cmdData.BackendStoragePath,
		},
		TmpDir:      cmdData.TmpDir,
		HomeDir:     cmdData.HomeDir,
		ProjectName: cmdData.ProjectName,
	})
}

type volumeUsageThresholdFlag struct {
	value      *thresholdpkg.Threshold
	onExplicit func()
	afterSet   func(*thresholdpkg.Threshold)
}

func newVolumeUsageThresholdFlag(target *thresholdpkg.Threshold, onExplicit func(), afterSet func(*thresholdpkg.Threshold)) *volumeUsageThresholdFlag {
	return &volumeUsageThresholdFlag{value: target, onExplicit: onExplicit, afterSet: afterSet}
}

func (f *volumeUsageThresholdFlag) String() string {
	if f == nil || f.value == nil {
		return ""
	}
	return f.value.FormatCLIValue()
}

func (f *volumeUsageThresholdFlag) Set(value string) error {
	threshold, err := thresholdpkg.Parse(value)
	if err != nil {
		return err
	}
	*f.value = threshold
	if f.onExplicit != nil {
		f.onExplicit()
	}
	if f.afterSet != nil {
		f.afterSet(f.value)
	}
	return nil
}

func (f *volumeUsageThresholdFlag) Type() string {
	return "volume-usage"
}

func initVolumeUsageThresholdTarget(target **thresholdpkg.Threshold, defaultValue string) {
	if *target != nil {
		return
	}

	parsedDefaultValue, err := thresholdpkg.Parse(defaultValue)
	if err != nil {
		panic(err)
	}

	*target = new(thresholdpkg.Threshold)
	**target = parsedDefaultValue
}

func setupVolumeUsageThresholdFlag(flagSet *pflag.FlagSet, target *thresholdpkg.Threshold, onExplicit func(), afterSet func(*thresholdpkg.Threshold), paramName, defaultValue, usage string) {
	flagSet.Var(newVolumeUsageThresholdFlag(target, onExplicit, afterSet), paramName, usage)

	flag := flagSet.Lookup(paramName)
	if flag == nil {
		panic(fmt.Sprintf("flag %q not found", paramName))
	}
	flag.DefValue = defaultValue
}

func firstEnvValue(envNames ...string) string {
	for _, envName := range envNames {
		if value := os.Getenv(envName); value != "" {
			return value
		}
	}
	return ""
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

	defaultValue := option.ValueOrDefault(firstEnvValue(aliases[0].EnvName, aliases[1].EnvName), host_cleaning.DefaultAllowedBackendStorageVolumeUsageThreshold().FormatCLIValue())

	initVolumeUsageThresholdTarget(&cmdData.AllowedBackendStorageVolumeUsage, defaultValue)

	for _, alias := range aliases {
		setupVolumeUsageThresholdFlag(cmd.Flags(), cmdData.AllowedBackendStorageVolumeUsage, nil, nil, alias.ParamName, defaultValue, fmt.Sprintf("Set the cleanup threshold for backend (Docker or Buildah) storage. Plain numbers define the maximum allowed usage percentage, while values with units define the minimum required free space, e.g. 70 or 10GB (default %s or $%s)", host_cleaning.DefaultAllowedBackendStorageVolumeUsageThreshold().String(), alias.EnvName))
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

	defaultValue := option.ValueOrDefault(firstEnvValue(aliases[0].EnvName, aliases[1].EnvName), host_cleaning.DefaultAllowedBackendStorageVolumeUsageMarginThreshold().FormatCLIValue())
	marginEnvValue := firstEnvValue(aliases[0].EnvName, aliases[1].EnvName)
	cmdData.AllowedBackendStorageVolumeUsageMarginExplicit = new(bool)
	*cmdData.AllowedBackendStorageVolumeUsageMarginExplicit = marginEnvValue != ""

	initVolumeUsageThresholdTarget(&cmdData.AllowedBackendStorageVolumeUsageMargin, defaultValue)
	marginTarget := cmdData.AllowedBackendStorageVolumeUsageMargin

	for _, alias := range aliases {
		setupVolumeUsageThresholdFlag(cmd.Flags(), marginTarget, func() { *cmdData.AllowedBackendStorageVolumeUsageMarginExplicit = true }, func(target *thresholdpkg.Threshold) {
			cmdData.AllowedBackendStorageVolumeUsageMargin = target
		}, alias.ParamName, defaultValue, fmt.Sprintf("Set the cleanup margin for backend (Docker or Buildah) storage. In percentage mode the margin is subtracted from the usage threshold, while in bytes mode it is added to the minimum required free space threshold, e.g. 5 or 2GB (default %s or $%s)", host_cleaning.DefaultAllowedBackendStorageVolumeUsageMarginThreshold().String(), alias.EnvName))
	}

	if marginEnvValue == "" {
		cmdData.AllowedBackendStorageVolumeUsageMargin = nil
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
	defaultValue := option.ValueOrDefault(os.Getenv(envVarName), host_cleaning.DefaultAllowedLocalCacheVolumeUsageThreshold().FormatCLIValue())

	initVolumeUsageThresholdTarget(&cmdData.AllowedLocalCacheVolumeUsage, defaultValue)
	setupVolumeUsageThresholdFlag(cmd.Flags(), cmdData.AllowedLocalCacheVolumeUsage, nil, nil, "allowed-local-cache-volume-usage", defaultValue, fmt.Sprintf("Set the cleanup threshold for local cache (~/.werf/local_cache by default). Plain numbers define the maximum allowed usage percentage, while values with units define the minimum required free space, e.g. 70 or 10GB (default %s or $%s)", host_cleaning.DefaultAllowedLocalCacheVolumeUsageThreshold().String(), envVarName))
}

func SetupAllowedLocalCacheVolumeUsageMargin(cmdData *CmdData, cmd *cobra.Command) {
	envVarName := "WERF_ALLOWED_LOCAL_CACHE_VOLUME_USAGE_MARGIN"
	marginEnvValue := os.Getenv(envVarName)
	defaultValue := option.ValueOrDefault(marginEnvValue, host_cleaning.DefaultAllowedLocalCacheVolumeUsageMarginThreshold().FormatCLIValue())
	cmdData.AllowedLocalCacheVolumeUsageMarginExplicit = new(bool)
	*cmdData.AllowedLocalCacheVolumeUsageMarginExplicit = marginEnvValue != ""

	initVolumeUsageThresholdTarget(&cmdData.AllowedLocalCacheVolumeUsageMargin, defaultValue)
	marginTarget := cmdData.AllowedLocalCacheVolumeUsageMargin
	setupVolumeUsageThresholdFlag(cmd.Flags(), marginTarget, func() { *cmdData.AllowedLocalCacheVolumeUsageMarginExplicit = true }, func(target *thresholdpkg.Threshold) { cmdData.AllowedLocalCacheVolumeUsageMargin = target }, "allowed-local-cache-volume-usage-margin", defaultValue, fmt.Sprintf("Set the cleanup margin for local cache. In percentage mode the margin is subtracted from the usage threshold, while in bytes mode it is added to the minimum required free space threshold, e.g. 5 or 2GB (default %s or $%s)", host_cleaning.DefaultAllowedLocalCacheVolumeUsageMarginThreshold().String(), envVarName))

	if marginEnvValue == "" {
		cmdData.AllowedLocalCacheVolumeUsageMargin = nil
	}
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
