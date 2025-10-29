package volumeutils_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestVolumeUsage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "VolumeUsage Suite")
}
