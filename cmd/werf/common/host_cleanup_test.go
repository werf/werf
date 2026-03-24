package common

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupAllowedBackendStorageVolumeUsageMargin_DefaultKeepsNil(t *testing.T) {
	t.Setenv("WERF_ALLOWED_BACKEND_STORAGE_VOLUME_USAGE_MARGIN", "")
	t.Setenv("WERF_ALLOWED_DOCKER_STORAGE_VOLUME_USAGE_MARGIN", "")

	var cmdData CmdData
	cmd := &cobra.Command{Use: "test"}

	SetupAllowedBackendStorageVolumeUsageMargin(&cmdData, cmd)

	assert.Nil(t, cmdData.AllowedBackendStorageVolumeUsageMargin)
	require.NotNil(t, cmdData.AllowedBackendStorageVolumeUsageMarginExplicit)
	assert.False(t, *cmdData.AllowedBackendStorageVolumeUsageMarginExplicit)
}

func TestSetupAllowedBackendStorageVolumeUsageMargin_EnvMakesValueExplicit(t *testing.T) {
	t.Setenv("WERF_ALLOWED_BACKEND_STORAGE_VOLUME_USAGE_MARGIN", "2GB")
	t.Setenv("WERF_ALLOWED_DOCKER_STORAGE_VOLUME_USAGE_MARGIN", "")

	var cmdData CmdData
	cmd := &cobra.Command{Use: "test"}

	SetupAllowedBackendStorageVolumeUsageMargin(&cmdData, cmd)

	require.NotNil(t, cmdData.AllowedBackendStorageVolumeUsageMargin)
	require.NotNil(t, cmdData.AllowedBackendStorageVolumeUsageMarginExplicit)
	assert.True(t, *cmdData.AllowedBackendStorageVolumeUsageMarginExplicit)
	assert.Equal(t, "2000000000B", cmdData.AllowedBackendStorageVolumeUsageMargin.FormatCLIValue())
}

func TestSetupAllowedBackendStorageVolumeUsageMargin_FlagMakesValueExplicit(t *testing.T) {
	t.Setenv("WERF_ALLOWED_BACKEND_STORAGE_VOLUME_USAGE_MARGIN", "")
	t.Setenv("WERF_ALLOWED_DOCKER_STORAGE_VOLUME_USAGE_MARGIN", "")

	var cmdData CmdData
	cmd := &cobra.Command{Use: "test"}

	SetupAllowedBackendStorageVolumeUsageMargin(&cmdData, cmd)

	require.NoError(t, cmd.Flags().Parse([]string{"--allowed-backend-storage-volume-usage-margin=2GB"}))
	require.NotNil(t, cmdData.AllowedBackendStorageVolumeUsageMargin)
	require.NotNil(t, cmdData.AllowedBackendStorageVolumeUsageMarginExplicit)
	assert.True(t, *cmdData.AllowedBackendStorageVolumeUsageMarginExplicit)
	assert.Equal(t, "2000000000B", cmdData.AllowedBackendStorageVolumeUsageMargin.FormatCLIValue())
}

func TestSetupAllowedLocalCacheVolumeUsageMargin_DefaultKeepsNil(t *testing.T) {
	t.Setenv("WERF_ALLOWED_LOCAL_CACHE_VOLUME_USAGE_MARGIN", "")

	var cmdData CmdData
	cmd := &cobra.Command{Use: "test"}

	SetupAllowedLocalCacheVolumeUsageMargin(&cmdData, cmd)

	assert.Nil(t, cmdData.AllowedLocalCacheVolumeUsageMargin)
	require.NotNil(t, cmdData.AllowedLocalCacheVolumeUsageMarginExplicit)
	assert.False(t, *cmdData.AllowedLocalCacheVolumeUsageMarginExplicit)
}

func TestSetupAllowedLocalCacheVolumeUsageMargin_EnvMakesValueExplicit(t *testing.T) {
	t.Setenv("WERF_ALLOWED_LOCAL_CACHE_VOLUME_USAGE_MARGIN", "2GB")

	var cmdData CmdData
	cmd := &cobra.Command{Use: "test"}

	SetupAllowedLocalCacheVolumeUsageMargin(&cmdData, cmd)

	require.NotNil(t, cmdData.AllowedLocalCacheVolumeUsageMargin)
	require.NotNil(t, cmdData.AllowedLocalCacheVolumeUsageMarginExplicit)
	assert.True(t, *cmdData.AllowedLocalCacheVolumeUsageMarginExplicit)
	assert.Equal(t, "2000000000B", cmdData.AllowedLocalCacheVolumeUsageMargin.FormatCLIValue())
}

func TestSetupAllowedLocalCacheVolumeUsageMargin_FlagMakesValueExplicit(t *testing.T) {
	t.Setenv("WERF_ALLOWED_LOCAL_CACHE_VOLUME_USAGE_MARGIN", "")

	var cmdData CmdData
	cmd := &cobra.Command{Use: "test"}

	SetupAllowedLocalCacheVolumeUsageMargin(&cmdData, cmd)

	require.NoError(t, cmd.Flags().Parse([]string{"--allowed-local-cache-volume-usage-margin=2GB"}))
	require.NotNil(t, cmdData.AllowedLocalCacheVolumeUsageMargin)
	require.NotNil(t, cmdData.AllowedLocalCacheVolumeUsageMarginExplicit)
	assert.True(t, *cmdData.AllowedLocalCacheVolumeUsageMarginExplicit)
	assert.Equal(t, "2000000000B", cmdData.AllowedLocalCacheVolumeUsageMargin.FormatCLIValue())
}
