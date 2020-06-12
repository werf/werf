package imports_test

import (
	"testing"

	"github.com/prashantv/gostub"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/werf/werf/pkg/testing/utils"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Imports suite")
}

var werfBinPath string
var stubs = gostub.New()

var _ = ginkgo.SynchronizedBeforeSuite(func() []byte {
	computedPathToWerf := utils.ProcessWerfBinPath()
	return []byte(computedPathToWerf)
}, func(computedPathToWerf []byte) {
	werfBinPath = string(computedPathToWerf)
})

var _ = ginkgo.SynchronizedAfterSuite(func() {}, func() {
	gexec.CleanupBuildArtifacts()
})

var _ = ginkgo.BeforeEach(func() {
	utils.BeforeEachOverrideWerfProjectName(stubs)
	stubs.SetEnv("WERF_DISABLE_AUTO_GC", "1")
})

var _ = ginkgo.AfterEach(func() {
	stubs.Reset()
})
