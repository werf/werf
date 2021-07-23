package true_git

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/integration/pkg/suite_init"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "True Git Suite")
}

var (
	SuiteData suite_init.SuiteData
	_         = SuiteData.SetupTmp(suite_init.NewTmpDirData())
)
