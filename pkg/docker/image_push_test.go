package docker

import (
	"context"

	"github.com/docker/cli/cli/command"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type fakeVersionCli struct {
	command.Cli
	version string
}

func (c fakeVersionCli) CurrentVersion() string {
	return c.version
}

var _ = Describe("docker image push", func() {
	DescribeTable("pushArgs",
		func(targetPlatform string, platformSupported bool, expected []string) {
			Expect(pushArgs("repo/img:tag", targetPlatform, platformSupported)).To(Equal(expected))
		},
		Entry("platform set and supported", "linux/amd64", true, []string{"--platform", "linux/amd64", "repo/img:tag"}),
		Entry("platform set but unsupported by the daemon", "linux/amd64", false, []string{"repo/img:tag"}),
		Entry("no platform", "", true, []string{"repo/img:tag"}),
	)

	DescribeTable("daemonSupports",
		func(current string, expected bool) {
			ctx := context.WithValue(context.Background(), ctxDockerCliKey, fakeVersionCli{version: current})
			Expect(daemonSupports(ctx, pushPlatformAPIVersion)).To(Equal(expected))
		},
		Entry("older daemon", "1.45", false),
		Entry("exact version", "1.46", true),
		Entry("newer daemon", "1.51", true),
	)
})
