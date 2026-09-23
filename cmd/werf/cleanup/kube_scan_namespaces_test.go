package cleanup

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("cleanup command flags", func() {
	It("registers --kube-scan-namespaces", func() {
		flag := NewCmd(context.Background()).Flags().Lookup("kube-scan-namespaces")

		Expect(flag).NotTo(BeNil())
		Expect(flag.Value.Type()).To(Equal("stringArray"))
	})
})
