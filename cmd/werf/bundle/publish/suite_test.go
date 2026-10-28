package publish

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCmdBundlePublish(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cmd Bundle Publish Suite")
}
