package buildah

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBuildah(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Buildah Suite")
}
