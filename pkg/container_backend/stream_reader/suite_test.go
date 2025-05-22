package stream_reader

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestImageReader(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Image Reader Suite")
}
