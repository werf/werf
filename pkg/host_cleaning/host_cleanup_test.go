package host_cleaning

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveBackendStorageVolumeUsageThresholds_DefaultPercentageMargin(t *testing.T) {
	t.Parallel()

	threshold, margin, err := resolveBackendStorageVolumeUsageThresholds(loToPtr(NewVolumeUsageThresholdPercentage(70)), nil, false)
	require.NoError(t, err)
	assert.Equal(t, NewVolumeUsageThresholdPercentage(70), threshold)
	assert.Equal(t, DefaultAllowedBackendStorageVolumeUsageMarginThreshold(), margin)
}

func TestResolveBackendStorageVolumeUsageThresholds_DefaultBytesMarginBecomesZeroBytes(t *testing.T) {
	t.Parallel()

	threshold, margin, err := resolveBackendStorageVolumeUsageThresholds(loToPtr(NewVolumeUsageThresholdBytes(10_000_000_000)), nil, false)
	require.NoError(t, err)
	assert.Equal(t, NewVolumeUsageThresholdBytes(10_000_000_000), threshold)
	assert.Equal(t, NewVolumeUsageThresholdBytes(0), margin)
}

func TestResolveBackendStorageVolumeUsageThresholds_ExplicitMixedFormatsReturnError(t *testing.T) {
	t.Parallel()

	_, _, err := resolveBackendStorageVolumeUsageThresholds(
		loToPtr(NewVolumeUsageThresholdBytes(10_000_000_000)),
		loToPtr(NewVolumeUsageThresholdPercentage(5)),
		true,
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must use the same format")
}

func TestResolveBackendStorageVolumeUsageThresholds_ImplicitMixedFormatsUseDefaultForThresholdType(t *testing.T) {
	t.Parallel()

	threshold, margin, err := resolveBackendStorageVolumeUsageThresholds(
		loToPtr(NewVolumeUsageThresholdBytes(10_000_000_000)),
		loToPtr(NewVolumeUsageThresholdPercentage(5)),
		false,
	)
	require.NoError(t, err)
	assert.Equal(t, NewVolumeUsageThresholdBytes(10_000_000_000), threshold)
	assert.Equal(t, NewVolumeUsageThresholdBytes(0), margin)
}

func TestResolveBackendStorageVolumeUsageThresholds_ExplicitSameFormatsPass(t *testing.T) {
	t.Parallel()

	threshold, margin, err := resolveBackendStorageVolumeUsageThresholds(
		loToPtr(NewVolumeUsageThresholdBytes(10_000_000_000)),
		loToPtr(NewVolumeUsageThresholdBytes(2_000_000_000)),
		true,
	)
	require.NoError(t, err)
	assert.Equal(t, NewVolumeUsageThresholdBytes(10_000_000_000), threshold)
	assert.Equal(t, NewVolumeUsageThresholdBytes(2_000_000_000), margin)
}

func TestResolveLocalCacheVolumeUsageThresholds_DefaultPercentageMargin(t *testing.T) {
	t.Parallel()

	threshold, margin, err := resolveLocalCacheVolumeUsageThresholds(loToPtr(NewVolumeUsageThresholdPercentage(70)), nil, false)
	require.NoError(t, err)
	assert.Equal(t, NewVolumeUsageThresholdPercentage(70), threshold)
	assert.Equal(t, DefaultAllowedLocalCacheVolumeUsageMarginThreshold(), margin)
}

func TestResolveLocalCacheVolumeUsageThresholds_DefaultBytesMarginBecomesZeroBytes(t *testing.T) {
	t.Parallel()

	threshold, margin, err := resolveLocalCacheVolumeUsageThresholds(loToPtr(NewVolumeUsageThresholdBytes(10_000_000_000)), nil, false)
	require.NoError(t, err)
	assert.Equal(t, NewVolumeUsageThresholdBytes(10_000_000_000), threshold)
	assert.Equal(t, NewVolumeUsageThresholdBytes(0), margin)
}

func TestResolveLocalCacheVolumeUsageThresholds_ExplicitMixedFormatsReturnError(t *testing.T) {
	t.Parallel()

	_, _, err := resolveLocalCacheVolumeUsageThresholds(
		loToPtr(NewVolumeUsageThresholdBytes(10_000_000_000)),
		loToPtr(NewVolumeUsageThresholdPercentage(5)),
		true,
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--allowed-local-cache-volume-usage")
}

func loToPtr[T any](v T) *T {
	return &v
}
