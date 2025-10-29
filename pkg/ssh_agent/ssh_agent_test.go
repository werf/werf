package ssh_agent

import (
	"context"
	"os"
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"github.com/werf/werf/v2/pkg/logging"
	"github.com/werf/werf/v2/pkg/werf"
)

var _ = Describe("Linux Fallback with SSH Environment", func() {
	type testCase struct {
		name                   string
		skipIfSSHSockEnvNotSet bool
		sshEnv                 string
		testPath               string
		errMatcher             types.GomegaMatcher
		validationFunc         func()
	}

	const (
		longPath = "/tmp/werf-test-agent-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	)

	DescribeTable("should handle different SSH agent fallback scenarios",
		func(ctx context.Context, tc testCase) {
			if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
				Skip("Skipping test on non-linux OS")
			}
			if tc.skipIfSSHSockEnvNotSet && os.Getenv(SSHAuthSockEnv) == "" {
				Skip("Skipping test because SSH_AUTH_SOCK is not set")
			}

			ctx = logging.WithLogger(ctx)

			if tc.sshEnv != "" {
				GinkgoT().Setenv(SSHAuthSockEnv, tc.sshEnv)
			}

			home, _ := os.UserHomeDir()

			err := os.MkdirAll(tc.testPath, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(tc.testPath)

			err = werf.Init(tc.testPath, home)
			Expect(err).NotTo(HaveOccurred())

			err = Init(ctx, []string{})
			Expect(err).To(tc.errMatcher)

			tc.validationFunc()
		},
		Entry("with SSH_ENV set", testCase{
			name:                   "with SSH_ENV set",
			skipIfSSHSockEnvNotSet: true,
			sshEnv:                 "",
			testPath:               "/tmp/werf-test-agent",
			errMatcher:             BeNil(),
			validationFunc: func() {
				valid, err := validateAgentSock(SSHAuthSock)
				Expect(err).NotTo(HaveOccurred())
				Expect(valid).To(BeTrue())
			},
		}),
		Entry("without SSH_ENV", testCase{
			name:                   "without SSH_ENV",
			skipIfSSHSockEnvNotSet: false,
			sshEnv:                 "",
			testPath:               "/tmp/werf-test-agent",
			errMatcher:             BeNil(),
			validationFunc: func() {
				if SSHAuthSock != "" {
					valid, err := validateAgentSock(SSHAuthSock)
					Expect(err).NotTo(HaveOccurred())
					Expect(valid).To(BeTrue())
				}
			},
		}),
		Entry("with SSH_ENV and wild tmp path", testCase{
			name:                   "with SSH_ENV and wild tmp path",
			skipIfSSHSockEnvNotSet: true,
			sshEnv:                 "",
			testPath:               longPath,
			errMatcher:             BeNil(),
			validationFunc: func() {
				valid, err := validateAgentSock(SSHAuthSock)
				Expect(err).NotTo(HaveOccurred())
				Expect(valid).To(BeTrue())
			},
		}),
		Entry("with SSH_ENV and long path", testCase{
			name:                   "with SSH_ENV and long path",
			skipIfSSHSockEnvNotSet: false,
			sshEnv:                 longPath,
			testPath:               "/tmp/werf-test-agent",
			errMatcher:             Not(BeNil()),
			validationFunc:         func() {},
		}),
	)
})
