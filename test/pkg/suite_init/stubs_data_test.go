package suite_init

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/prashantv/gostub"
)

func TestSuiteInit(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Suite initialization")
}

var _ = ginkgo.Describe("Project cleanup backend selection", func() {
	data := SuiteData{StubsData: NewStubsData(), ProjectNameData: &ProjectNameData{ProjectName: "werf-test-one"}}
	var wantDocker, wantBuildah bool
	canStubBuildah := runtime.GOOS == "linux" && os.Geteuid() != 0
	ginkgo.BeforeEach(func() {
		dir := ginkgo.GinkgoT().TempDir()
		stub := []byte("#!/bin/sh\n: > \"$0.called\"\n")
		gomega.Expect(os.WriteFile(filepath.Join(dir, "docker"), stub, 0o700)).To(gomega.Succeed())
		if canStubBuildah {
			gomega.Expect(os.WriteFile(filepath.Join(dir, "buildah"), stub, 0o700)).To(gomega.Succeed())
		}
		data.Stubs.SetEnv("PATH", dir)
		data.Stubs.UnsetEnv("WERF_BUILDAH_MODE")
		ginkgo.DeferCleanup(func() {
			for backend, expected := range map[string]bool{"docker": wantDocker, "buildah": canStubBuildah && wantBuildah} {
				marker := filepath.Join(dir, backend+".called")
				if expected {
					gomega.Expect(marker).To(gomega.BeAnExistingFile())
				} else {
					gomega.Expect(marker).NotTo(gomega.BeAnExistingFile())
				}
			}
		})
	})
	data.SetupProjectCleanup()
	ginkgo.DescribeTable("cleans the stores used by the spec", func(inherited string, modes []string, expectDocker, expectBuildah bool) {
		wantDocker, wantBuildah = expectDocker, expectBuildah
		data.Stubs.SetEnv("WERF_BUILDAH_MODE", inherited)
		for _, mode := range modes {
			data.SetContainerBackendMode(mode)
		}
	}, ginkgo.Entry("ordinary Docker spec", "", nil, true, false),
		ginkgo.Entry("Buildah-only spec", "", []string{"native-rootless"}, false, true),
		ginkgo.Entry("Docker before switching to Buildah", "", []string{"docker", "native-rootless"}, true, true),
		ginkgo.Entry("untracked native build may also use Docker Compose", "native-rootless", nil, true, true))
	ginkgo.Context("consecutive specs on one worker", ginkgo.Ordered, func() {
		ginkgo.It("records Buildah for the first spec", func() {
			wantDocker, wantBuildah = false, true
			data.SetContainerBackendMode("native-rootless")
		})
		ginkgo.It("cleans Docker without inheriting the first spec's backend", func() {
			wantDocker, wantBuildah = true, false
		})
	})
})

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
