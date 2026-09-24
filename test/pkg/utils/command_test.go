package utils

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Utils Suite")
}

var _ = Describe("RunCommandWithOptions", func() {
	It("drains stderr while command writes", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		output, err := RunCommandWithOptions(ctx, "", "sh", []string{"-c", "dd if=/dev/zero bs=1024 count=128 >&2; printf done"}, RunCommandOptions{})

		Expect(err).NotTo(HaveOccurred())
		Expect(string(output)).To(ContainSubstring("done"))
	})
})
