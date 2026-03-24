package threshold

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		value        string
		expected     Threshold
		expectsError bool
	}{
		{name: "percentage 70", value: "70", expected: NewPercentage(70)},
		{name: "percentage 0", value: "0", expected: NewPercentage(0)},
		{name: "percentage 100", value: "100", expected: NewPercentage(100)},
		{name: "bytes gigabytes", value: "10GB", expected: NewBytes(10_000_000_000)},
		{name: "bytes megabytes short b", value: "10Mb", expected: NewBytes(10_000_000)},
		{name: "invalid text", value: "wat", expectsError: true},
		{name: "invalid percent sign", value: "101%", expectsError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th, err := Parse(tt.value)
			if tt.expectsError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, th)
		})
	}
}

func TestResolve(t *testing.T) {
	t.Parallel()

	threshold, margin, err := Resolve(ptr(NewPercentage(70)), nil, NewPercentage(70), NewPercentage(5), false, "--foo", "--bar")
	require.NoError(t, err)
	assert.Equal(t, NewPercentage(70), threshold)
	assert.Equal(t, NewPercentage(5), margin)

	threshold, margin, err = Resolve(ptr(NewBytes(10_000_000_000)), nil, NewPercentage(70), NewPercentage(5), false, "--foo", "--bar")
	require.NoError(t, err)
	assert.Equal(t, NewBytes(10_000_000_000), threshold)
	assert.Equal(t, NewBytes(0), margin)

	_, _, err = Resolve(ptr(NewBytes(10_000_000_000)), ptr(NewPercentage(5)), NewPercentage(70), NewPercentage(5), true, "--foo", "--bar")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must use the same format")
}

func ptr[T any](v T) *T { return &v }
