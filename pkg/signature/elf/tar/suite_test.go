package tar_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTar(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tar Suite")
}
