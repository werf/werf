package stapel

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAI_GetVersion_Default(t *testing.T) {
	t.Setenv("WERF_STAPEL_IMAGE_VERSION", "")
	t.Setenv("WERF_EXPERIMENTAL_STAPEL_ARM", "")

	require.Equal(t, VERSION, getVersion())
}

func TestAI_GetVersion_ExperimentalArmUsesDev(t *testing.T) {
	t.Setenv("WERF_STAPEL_IMAGE_VERSION", "")
	t.Setenv("WERF_EXPERIMENTAL_STAPEL_ARM", "1")

	require.Equal(t, "dev", getVersion())
}

func TestAI_GetVersion_ExplicitVersionHasPriority(t *testing.T) {
	t.Setenv("WERF_STAPEL_IMAGE_VERSION", "custom")
	t.Setenv("WERF_EXPERIMENTAL_STAPEL_ARM", "1")

	require.Equal(t, "custom", getVersion())
}
