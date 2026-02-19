package sbom

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSbom(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sbom Suite")
}
