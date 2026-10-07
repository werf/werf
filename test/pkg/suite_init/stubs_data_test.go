package suite_init

import (
	"os"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/prashantv/gostub"
)

func TestSuiteInit(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Suite initialization")
}

var _ = ginkgo.Describe("Stub cleanup ordering", func() {
	const variable = "WERF_TEST_STUB_CLEANUP_ORDER"
	stubs := gostub.New()

	ginkgo.BeforeEach(func() {
		original, present := os.LookupEnv(variable)
		gomega.Expect(os.Setenv(variable, "original")).To(gomega.Succeed())
		ginkgo.DeferCleanup(func() {
			value := os.Getenv(variable)
			if present {
				gomega.Expect(os.Setenv(variable, original)).To(gomega.Succeed())
			} else {
				gomega.Expect(os.Unsetenv(variable)).To(gomega.Succeed())
			}
			gomega.Expect(value).To(gomega.Equal("original"))
		})
	})
	SetupStubs(stubs)
	ginkgo.AfterEach(func() {
		gomega.Expect(os.Getenv(variable)).To(gomega.Equal("test-backend"))
	})
	ginkgo.It("retains the test environment through resource cleanup and then restores it", func() {
		stubs.SetEnv(variable, "test-backend")
		ginkgo.DeferCleanup(func() {
			gomega.Expect(os.Getenv(variable)).To(gomega.Equal("test-backend"))
		})
	})
})
