package build

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func TestBuildCommand(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Build Command Suite")
}

var _ = ginkgo.Describe("Promote built images flags", func() {
	ginkgo.BeforeEach(func() {
		for _, entry := range os.Environ() {
			key, _, _ := strings.Cut(entry, "=")
			if strings.HasPrefix(key, "WERF_") {
				ginkgo.GinkgoT().Setenv(key, "")
			}
		}
	})

	ginkgo.DescribeTable("rejects invalid mode combinations before initializing the build",
		func(ctx ginkgo.SpecContext, args []string, env map[string]string, expected string) {
			for key, value := range env {
				ginkgo.GinkgoT().Setenv(key, value)
			}
			cmd := NewCmd(ctx)
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)
			cmd.SetArgs(append([]string{}, args...))
			gomega.Expect(cmd.ExecuteContext(ctx)).To(gomega.MatchError(gomega.ContainSubstring(expected)))
		},
		ginkgo.Entry("check flag", []string{"--promote-built-images", "--check-built-images"}, nil, "cannot be used with --check-built-images"),
		ginkgo.Entry("legacy check flag", []string{"--promote-built-images", "--require-built-images"}, nil, "cannot be used with --check-built-images"),
		ginkgo.Entry("short check flag", []string{"--promote-built-images", "-Z"}, nil, "cannot be used with --check-built-images"),
		ginkgo.Entry("check environment", []string{"--promote-built-images"}, map[string]string{"WERF_CHECK_BUILT_IMAGES": "true"}, "cannot be used with --check-built-images"),
		ginkgo.Entry("legacy environment does not select check mode", []string{"--promote-built-images"}, map[string]string{"WERF_REQUIRE_BUILT_IMAGES": "true"}, "requires at least one --final-repo"),
		ginkgo.Entry("no destination", []string{"--promote-built-images"}, nil, "requires at least one --final-repo"),
		ginkgo.Entry("empty destination", []string{"--promote-built-images", "--final-repo="}, nil, "requires at least one --final-repo"),
		ginkgo.Entry("promotion environment without destination", nil, map[string]string{"WERF_PROMOTE_BUILT_IMAGES": "true"}, "requires at least one --final-repo"),
	)
})
