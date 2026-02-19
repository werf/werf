package cyclonedxutil

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCyclonedxUtil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CyclonedxUtil Suite")
}
