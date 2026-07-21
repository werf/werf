package container_backend

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Locks in the sentinel behavior consumed by pkg/host_cleaning/local_backend_cleaner.go via
// errors.Is, so the go.podman.io/storage/types alias can be replaced with a locally defined
// sentinel without regressing the cleanup error handling.
func TestAI_ErrImageUsedByContainer_Sentinel(t *testing.T) {
	assert.EqualError(t, ErrImageUsedByContainer, "image is in use by a container")
	assert.True(t, errors.Is(fmt.Errorf("prune images: %w", ErrImageUsedByContainer), ErrImageUsedByContainer))
}
