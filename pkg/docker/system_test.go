package docker

import (
	"context"
	"errors"

	"github.com/docker/cli/cli/command"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("docker system", func() {
	Describe("cliWithCustomOptions", func() {
		It("uses a separate Docker CLI", func() {
			ctx, err := NewContext(context.Background())
			Expect(err).NotTo(HaveOccurred())

			originalCli := cli(ctx)
			var customCli command.Cli
			err = cliWithCustomOptions(ctx, nil, func(c command.Cli) error {
				customCli = c
				return nil
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(customCli).NotTo(Equal(originalCli))
			Expect(cli(ctx)).To(Equal(originalCli))
		})
	})

	Describe("isDaemonUnavailableErr", func() {
		It("should return false for nil error", func() {
			Expect(isDaemonUnavailableErr(nil)).To(BeFalse())
		})

		It("should return true for connection refused", func() {
			err := errors.New("connect: connection refused")
			Expect(isDaemonUnavailableErr(err)).To(BeTrue())
		})

		It("should return true for no such file", func() {
			err := errors.New("connect: no such file or directory")
			Expect(isDaemonUnavailableErr(err)).To(BeTrue())
		})

		It("should return true for cannot connect", func() {
			err := errors.New("Cannot connect to the Docker daemon")
			Expect(isDaemonUnavailableErr(err)).To(BeTrue())
		})

		It("should return false for other errors", func() {
			err := errors.New("some other error")
			Expect(isDaemonUnavailableErr(err)).To(BeFalse())
		})
	})
})
