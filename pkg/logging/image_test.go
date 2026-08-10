package logging

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// disablePrettyLogForTest calls DisablePrettyLog (which mutates the
// package-level pretty-print prefixes) and restores the original values on
// cleanup, so these tests don't leak global state to any test that runs
// after them in the same process.
func disablePrettyLogForTest(t *testing.T) {
	t.Helper()

	prevFinal, prevIntermediate := finalImagePrettyPrefix, intermediateImagePrettyPrefix
	t.Cleanup(func() {
		finalImagePrettyPrefix, intermediateImagePrettyPrefix = prevFinal, prevIntermediate
	})

	DisablePrettyLog()
}

func TestImageLogProcessName_WithWorker(t *testing.T) {
	disablePrettyLogForTest(t)

	got := ImageLogProcessName("app", true, "linux/amd64", WithProgress(2, 6), WithWorker(1))
	require.Equal(t, "(2/6) image app [linux/amd64] (worker 1)", got)
}

func TestImageLogProcessName_WithoutWorker(t *testing.T) {
	disablePrettyLogForTest(t)

	got := ImageLogProcessName("app", true, "linux/amd64", WithProgress(2, 6))
	require.Equal(t, "(2/6) image app [linux/amd64]", got)
	require.NotContains(t, got, "worker")
}

func TestImageLogProcessName_WorkerWithoutProgress(t *testing.T) {
	disablePrettyLogForTest(t)

	got := ImageLogProcessName("app", false, "", WithWorker(0))
	require.Equal(t, "image app (worker 0)", got)
}
