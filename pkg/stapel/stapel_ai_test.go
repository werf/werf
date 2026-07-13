package stapel

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/image"
)

var _ = Describe("stapel container naming", func() {
	It("uses version-only suffix when platform is not specified", func() {
		container := getContainer("")

		expectedName := fmt.Sprintf("%s%s", image.AssemblingContainerNamePrefix, getVersion())
		Expect(container.Name).To(Equal(expectedName))
	})

	It("adds platform suffix when platform is specified", func() {
		container := getContainer("linux/arm64")

		expectedName := fmt.Sprintf("%s%s_linux_arm64", image.AssemblingContainerNamePrefix, getVersion())
		Expect(container.Name).To(Equal(expectedName))
	})

})
