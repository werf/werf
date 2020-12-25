package suite_init

import (
	"fmt"
	"os"
	"testing"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/werf/werf/integration/utils"
)

type TestSuiteEntrypointFuncOptions struct {
	RequiredSuiteTools []string
	RequiredSuiteEnvs  []string
}

func MakeTestSuiteEntrypointFunc(description string, opts TestSuiteEntrypointFuncOptions) func(t *testing.T) {
	return func(t *testing.T) {
		if !utils.MeetsRequirements(opts.RequiredSuiteTools, opts.RequiredSuiteEnvs) {
			fmt.Println("Missing required tools")
			os.Exit(1)
		}

		gomega.RegisterFailHandler(ginkgo.Fail)
		ginkgo.RunSpecs(t, description)
	}
}
