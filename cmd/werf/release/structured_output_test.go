package release_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"sigs.k8s.io/yaml"
)

func TestRelease(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Release Structured Output Suite")
}

func TestMain(m *testing.M) {
	if subCommand := os.Getenv(subCommandEnv); subCommand != "" {
		os.Exit(runSubCommand(subCommand, strings.Split(os.Getenv(subCommandArgsEnv), "\n")))
	}

	os.Exit(m.Run())
}

var _ = ginkgo.Describe("release structured output", func() {
	var kubeConfigPath string

	ginkgo.BeforeEach(func() {
		kubeConfigPath = startKubeStub()
	})

	commandArgs := func(subCommand, outputFormat string) []string {
		args := []string{
			"--kube-config", kubeConfigPath,
			"--kube-request-timeout=5s",
			"--namespace", testReleaseNamespace,
			"--output-format", outputFormat,
		}

		if subCommand != "list" {
			args = append(args, "--release", testReleaseName)
		}

		return args
	}

	ginkgo.DescribeTable("stays machine-readable with default log options in CI",
		func(subCommand, ciEnvVar, outputFormat string) {
			result := runReleaseCommand(subCommand, ciEnvVar, commandArgs(subCommand, outputFormat))

			gomega.Expect(result.ExitCode).To(gomega.Equal(0), "stderr: %s", result.Stderr)
			gomega.Expect(result.Stdout).NotTo(gomega.ContainSubstring("\x1b["), "stdout must not contain ANSI escapes")

			var parsed map[string]interface{}
			if outputFormat == "json" {
				gomega.Expect(json.Unmarshal([]byte(result.Stdout), &parsed)).To(gomega.Succeed(), "stdout: %q", result.Stdout)
			} else {
				gomega.Expect(yaml.Unmarshal([]byte(result.Stdout), &parsed)).To(gomega.Succeed(), "stdout: %q", result.Stdout)
			}

			gomega.Expect(parsed).NotTo(gomega.BeEmpty())
			gomega.Expect(result.Stdout).To(gomega.ContainSubstring(testReleaseName))
		},
		ginkgo.Entry("list json in GitHub Actions", "list", "GITHUB_ACTIONS", "json"),
		ginkgo.Entry("list json in GitLab CI", "list", "GITLAB_CI", "json"),
		ginkgo.Entry("list yaml in GitHub Actions", "list", "GITHUB_ACTIONS", "yaml"),
		ginkgo.Entry("get json in GitHub Actions", "get", "GITHUB_ACTIONS", "json"),
		ginkgo.Entry("get yaml in GitLab CI", "get", "GITLAB_CI", "yaml"),
		ginkgo.Entry("history json in GitHub Actions", "history", "GITHUB_ACTIONS", "json"),
		ginkgo.Entry("history yaml in GitLab CI", "history", "GITLAB_CI", "yaml"),
	)

	ginkgo.DescribeTable("keeps the log flag precedence of the other commands",
		func(logArgs []string, expectProgressLogs bool) {
			result := runReleaseCommand("list", "GITHUB_ACTIONS", append(commandArgs("list", "json"), logArgs...))

			gomega.Expect(result.ExitCode).To(gomega.Equal(0), "stderr: %s", result.Stderr)

			progressLogs := gomega.Expect(result.Stdout + result.Stderr)
			if expectProgressLogs {
				progressLogs.To(gomega.ContainSubstring("List releases"))
			} else {
				progressLogs.NotTo(gomega.ContainSubstring("List releases"))
			}
		},
		ginkgo.Entry("debug logs progress", []string{"--log-debug"}, true),
		ginkgo.Entry("verbose logs progress", []string{"--log-verbose"}, true),
		ginkgo.Entry("quiet stays silent", []string{"--log-quiet"}, false),
		ginkgo.Entry("quiet wins over verbose", []string{"--log-quiet", "--log-verbose"}, false),
		ginkgo.Entry("debug wins over quiet", []string{"--log-quiet", "--log-debug"}, true),
	)
})
