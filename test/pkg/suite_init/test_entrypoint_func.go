package suite_init

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/test/pkg/utils"
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

		RegisterFailHandler(Fail)
		RunSpecs(t, description)
	}
}
