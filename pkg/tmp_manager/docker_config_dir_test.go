package tmp_manager

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prashantv/gostub"

	"github.com/werf/werf/v3/pkg/werf"
)

var _ = Describe("docker config dir", func() {
	// `werf ci-env` hands its docker config dir to the build that follows, so
	// a concurrent `werf host cleanup` must leave a fresh one alone.
	DescribeTable("survives host cleanup until it is stale",
		func(age time.Duration, expectCollected bool) {
			stubs := gostub.New()
			defer stubs.Reset()
			stubs.SetEnv("WERF_TMP_DIR", GinkgoT().TempDir())
			stubs.SetEnv("WERF_HOME", GinkgoT().TempDir())
			Expect(werf.Init("", "")).To(Succeed())

			fromDockerConfig := GinkgoT().TempDir()
			Expect(os.WriteFile(filepath.Join(fromDockerConfig, "config.json"), []byte("{}"), 0o600)).To(Succeed())

			dir, err := CreateDockerConfigDir(GinkgoT().Context(), fromDockerConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(DelegateCleanup(GinkgoT().Context())).To(Succeed())

			stubs.StubFunc(&timeSince, age)

			dirsToRemove, symlinksToRemove, err := collectPaths()
			Expect(err).NotTo(HaveOccurred())

			if expectCollected {
				Expect(dirsToRemove).To(ContainElement(dir))
				Expect(symlinksToRemove).NotTo(BeEmpty())
			} else {
				Expect(dirsToRemove).NotTo(ContainElement(dir))
				Expect(symlinksToRemove).To(BeEmpty())
			}
		},
		Entry("just created", time.Minute, false),
		Entry("created an hour ago", time.Hour, false),
		Entry("created seven hours ago", time.Hour*7, true),
	)
})
