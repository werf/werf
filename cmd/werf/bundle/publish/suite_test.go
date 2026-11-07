package publish

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/werf"
)

func TestCmdBundlePublish(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cmd Bundle Publish Suite")
}

var _ = BeforeSuite(func() {
	t := GinkgoT()
	Expect(werf.Init(t.TempDir(), t.TempDir())).To(Succeed())
})
