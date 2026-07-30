package logging

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAI_ImageLogProcessName_WithWorker(t *testing.T) {
	DisablePrettyLog()

	got := ImageLogProcessName("app", true, "linux/amd64", WithProgress(2, 6), WithWorker(1))
	require.Equal(t, "(2/6) image app [linux/amd64] (worker 1)", got)
}

func TestAI_ImageLogProcessName_WithoutWorker(t *testing.T) {
	DisablePrettyLog()

	got := ImageLogProcessName("app", true, "linux/amd64", WithProgress(2, 6))
	require.Equal(t, "(2/6) image app [linux/amd64]", got)
	require.NotContains(t, got, "worker")
}

func TestAI_ImageLogProcessName_WorkerWithoutProgress(t *testing.T) {
	DisablePrettyLog()

	got := ImageLogProcessName("app", false, "", WithWorker(0))
	require.Equal(t, "image app (worker 0)", got)
}
