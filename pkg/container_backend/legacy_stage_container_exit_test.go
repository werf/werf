package container_backend

import (
	"errors"
	"fmt"

	"github.com/docker/cli/cli"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Legacy stage container exit status", func() {
	DescribeTable("containerExitCode",
		func(err error, expectedCode int, expectedOK bool) {
			code, ok := containerExitCode(err)
			Expect(ok).To(Equal(expectedOK))
			Expect(code).To(Equal(expectedCode))
		},
		Entry("a bare status error", cli.StatusError{StatusCode: 23}, 23, true),
		Entry("a wrapped status error", fmt.Errorf("container run failed: %w", cli.StatusError{StatusCode: 23}), 23, true),
		Entry("a status error carrying a message", cli.StatusError{StatusCode: 125, Status: "no such image"}, 125, true),
		Entry("any other error", errors.New("no such host"), 0, false),
	)

	DescribeTable("IsStartContainerErr",
		func(err error, expected bool) {
			Expect(IsStartContainerErr(err)).To(Equal(expected))
		},
		Entry("125 is a start failure", cli.StatusError{StatusCode: 125}, true),
		Entry("126 is a start failure", cli.StatusError{StatusCode: 126}, true),
		Entry("127 is a start failure", cli.StatusError{StatusCode: 127}, true),
		Entry("a command exit code is not", cli.StatusError{StatusCode: 127 + 1}, false),
		Entry("a wrapped start failure", fmt.Errorf("container run failed: %w", cli.StatusError{StatusCode: 126}), true),
		Entry("any other error", errors.New("no such host"), false),
	)

	DescribeTable("namedContainerExitErr",
		func(err error, expectedMessage string) {
			named := namedContainerExitErr(err)
			Expect(named.Error()).To(Equal(expectedMessage))
			Expect(errors.Is(named, err)).To(BeTrue())
		},
		Entry("a message-less status error gets its code spelled out", cli.StatusError{StatusCode: 23}, "exit code 23"),
		Entry("a status error with a message keeps it", cli.StatusError{StatusCode: 125, Status: "no such image"}, "no such image"),
		Entry("any other error is left alone", errors.New("no such host"), "no such host"),
	)

	It("reports the exit code through the run wrapping", func() {
		err := fmt.Errorf("container run failed: %w", namedContainerExitErr(cli.StatusError{StatusCode: 23}))

		Expect(err.Error()).To(Equal("container run failed: exit code 23"))
		code, ok := containerExitCode(err)
		Expect(ok).To(BeTrue())
		Expect(code).To(Equal(23))
	})
})
