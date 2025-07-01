package scanner

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSbomScanner(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sbom Scanner Suite")
}
