package container_backend

import (
	"testing"

	ocispecs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
)

func TestAI_ValidateWorkerPlatform(t *testing.T) {
	workerPlatforms := []ocispecs.Platform{
		{OS: "linux", Architecture: "arm64"},
		{OS: "linux", Architecture: "amd64"},
		{OS: "linux", Architecture: "arm", Variant: "v7"},
	}

	assert.NoError(t, validateWorkerPlatform(ocispecs.Platform{OS: "linux", Architecture: "amd64"}, workerPlatforms))
	assert.NoError(t, validateWorkerPlatform(ocispecs.Platform{OS: "linux", Architecture: "arm", Variant: "v7"}, workerPlatforms))
	assert.NoError(t, validateWorkerPlatform(ocispecs.Platform{OS: "linux", Architecture: "amd64"}, nil))

	err := validateWorkerPlatform(ocispecs.Platform{OS: "linux", Architecture: "s390x"}, workerPlatforms)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "linux/s390x")
	assert.Contains(t, err.Error(), "tonistiigi/binfmt")
}
