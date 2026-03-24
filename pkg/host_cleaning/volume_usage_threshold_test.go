package host_cleaning

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/werf/werf/v2/pkg/volumeutils"
)

func TestParseVolumeUsageThreshold(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		value        string
		expected     VolumeUsageThreshold
		expectsError bool
	}{
		{name: "percentage 70", value: "70", expected: NewVolumeUsageThresholdPercentage(70)},
		{name: "percentage 0", value: "0", expected: NewVolumeUsageThresholdPercentage(0)},
		{name: "percentage 100", value: "100", expected: NewVolumeUsageThresholdPercentage(100)},
		{name: "bytes gigabytes", value: "10GB", expected: NewVolumeUsageThresholdBytes(10_000_000_000)},
		{name: "bytes megabytes short b", value: "10Mb", expected: NewVolumeUsageThresholdBytes(10_000_000)},
		{name: "invalid text", value: "wat", expectsError: true},
		{name: "invalid percent sign", value: "101%", expectsError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			threshold, err := ParseVolumeUsageThreshold(tt.value)
			if tt.expectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, threshold)
		})
	}
}

func TestExceedsVolumeUsageThreshold(t *testing.T) {
	t.Parallel()

	vu := volumeutils.VolumeUsage{UsedBytes: 80, TotalBytes: 100}

	assert.False(t, exceedsVolumeUsageThreshold(vu, NewVolumeUsageThresholdPercentage(80)))
	assert.True(t, exceedsVolumeUsageThreshold(vu, NewVolumeUsageThresholdPercentage(70)))
	assert.False(t, exceedsVolumeUsageThreshold(vu, NewVolumeUsageThresholdBytes(20)))
	assert.True(t, exceedsVolumeUsageThreshold(vu, NewVolumeUsageThresholdBytes(21)))
}

func TestTargetBytesToFree(t *testing.T) {
	t.Parallel()

	vu := volumeutils.VolumeUsage{UsedBytes: 80, TotalBytes: 100}

	assert.Equal(t, uint64(15), targetBytesToFree(vu, NewVolumeUsageThresholdPercentage(70), NewVolumeUsageThresholdPercentage(5)))
	assert.Equal(t, uint64(5), targetBytesToFree(vu, NewVolumeUsageThresholdBytes(20), NewVolumeUsageThresholdBytes(5)))
	assert.Equal(t, uint64(10), targetBytesToFree(vu, NewVolumeUsageThresholdBytes(25), NewVolumeUsageThresholdBytes(5)))
}

func TestTargetBytesToFree_MixedTypesPanics(t *testing.T) {
	t.Parallel()

	vu := volumeutils.VolumeUsage{UsedBytes: 80, TotalBytes: 100}

	assert.Panics(t, func() {
		targetBytesToFree(vu, NewVolumeUsageThresholdPercentage(70), NewVolumeUsageThresholdBytes(5))
	})
	assert.Panics(t, func() {
		targetBytesToFree(vu, NewVolumeUsageThresholdBytes(20), NewVolumeUsageThresholdPercentage(5))
	})
}
