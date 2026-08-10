package slug

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSlugSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Slug Suite")
}
