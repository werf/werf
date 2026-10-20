package build

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/cmd/werf/common"
)

func TestBuildCommand(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Build Command Suite")
}

var _ = ginkgo.Describe("Require built images flags", func() {
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
		ginkgo.Entry("check flag", []string{"--require-built-images", "--check-built-images"}, nil, "cannot be used with --check-built-images"),
		ginkgo.Entry("short flag", []string{"-Z", "--check-built-images"}, nil, "cannot be used with --check-built-images"),
		ginkgo.Entry("check environment", []string{"--require-built-images"}, map[string]string{"WERF_CHECK_BUILT_IMAGES": "true"}, "cannot be used with --check-built-images"),
		ginkgo.Entry("require environment", []string{"--check-built-images"}, map[string]string{"WERF_REQUIRE_BUILT_IMAGES": "true"}, "cannot be used with --check-built-images"),
		ginkgo.Entry("both environments", nil, map[string]string{"WERF_REQUIRE_BUILT_IMAGES": "true", "WERF_CHECK_BUILT_IMAGES": "true"}, "cannot be used with --check-built-images"),
	)
	ginkgo.DescribeTable("selects independent check and require modes", func(ctx ginkgo.SpecContext, args []string, env map[string]string, require, check bool) {
		for key, value := range env {
			ginkgo.GinkgoT().Setenv(key, value)
		}
		cmd := NewCmd(ctx)
		gomega.Expect(cmd.ParseFlags(args)).To(gomega.Succeed())
		gomega.Expect(common.GetRequireBuiltImages(&commonCmdData)).To(gomega.Equal(require))
		gomega.Expect(common.GetCheckBuiltImages(&commonCmdData)).To(gomega.Equal(check))
	},
		ginkgo.Entry("ordinary build", nil, nil, false, false),
		ginkgo.Entry("require flag", []string{"--require-built-images"}, nil, true, false),
		ginkgo.Entry("short flag", []string{"-Z"}, nil, true, false),
		ginkgo.Entry("require environment", nil, map[string]string{"WERF_REQUIRE_BUILT_IMAGES": "true"}, true, false),
		ginkgo.Entry("check flag", []string{"--check-built-images"}, nil, false, true),
		ginkgo.Entry("check environment", nil, map[string]string{"WERF_CHECK_BUILT_IMAGES": "true"}, false, true),
		ginkgo.Entry("explicit require override", []string{"--require-built-images=false"}, map[string]string{"WERF_REQUIRE_BUILT_IMAGES": "true"}, false, false),
		ginkgo.Entry("explicit check override", []string{"--check-built-images=false", "-Z"}, map[string]string{"WERF_CHECK_BUILT_IMAGES": "true"}, true, false),
	)
})
