// +build integration integration_k8s

package ansible_test

import (
	"testing"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/flant/werf/pkg/testing/utils"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Ansible suite")
}

var werfBinPath string

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
	utils.BeforeEachOverrideWerfProjectName()
})

var _ = ginkgo.AfterEach(func() {
	utils.ResetEnviron()
})
