package true_git

import (
	"github.com/werf/werf/integration/pkg/suite_init"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "True Git Suite")
}

var SuiteData suite_init.SuiteData
var _ = SuiteData.SetupTmp(suite_init.NewTmpDirData())