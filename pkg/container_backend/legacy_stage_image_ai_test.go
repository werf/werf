package container_backend

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAI_IsArm64Platform(t *testing.T) {
	require.True(t, isArm64Platform("linux/arm64"))
	require.True(t, isArm64Platform("linux/arm64/v8"))
	require.False(t, isArm64Platform("linux/amd64"))
	require.False(t, isArm64Platform("linux/arm/v7"))
}
