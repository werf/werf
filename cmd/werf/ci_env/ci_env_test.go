package ci_env

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/joho/godotenv"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prashantv/gostub"
	"github.com/samber/lo"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CI env suite")
}

var _ = Describe("ci-env env file", func() {
	DescribeTable("exports WERF_DOCKER_CONFIG alongside DOCKER_CONFIG",
		func(ctx context.Context, generate func(context.Context, io.Writer, string) error) {
			stubs := gostub.New()
			defer stubs.Reset()

			for _, key := range []string{"CI_REGISTRY_IMAGE", "CI_JOB_TOKEN", "GITHUB_REPOSITORY", "GITHUB_TOKEN"} {
				stubs.SetEnv(key, "")
			}
			stubs.SetEnv("DOCKER_CONFIG", "")
			stubs.SetEnv("WERF_DOCKER_CONFIG", "")

			commonCmdData.LogVerbose = lo.ToPtr(false)
			cmdData.AsEnvFile = true
			defer func() { cmdData.AsEnvFile = false }()

			var buf bytes.Buffer
			Expect(generate(ctx, &buf, "/tmp/werf-docker-config")).To(Succeed())

			envs, err := godotenv.Parse(&buf)
			Expect(err).NotTo(HaveOccurred())

			Expect(envs).To(HaveKeyWithValue("DOCKER_CONFIG", "/tmp/werf-docker-config"))
			Expect(envs).To(HaveKeyWithValue("WERF_DOCKER_CONFIG", "/tmp/werf-docker-config"))
		},
		Entry("gitlab", generateGitlabEnvs),
		Entry("github", generateGithubEnvs),
	)
})
