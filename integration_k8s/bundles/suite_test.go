package bundles_test

import (
	"testing"

	"github.com/werf/werf/integration/suite_init"
)

var testSuiteEntrypointFunc = suite_init.MakeTestSuiteEntrypointFunc("Bundles suite", suite_init.TestSuiteEntrypointFuncOptions{
	RequiredSuiteTools: []string{"git", "docker"},
	RequiredSuiteEnvs:  []string{},
})

func TestSuite(t *testing.T) {
	testSuiteEntrypointFunc(t)
}

var SuiteData suite_init.SuiteData
