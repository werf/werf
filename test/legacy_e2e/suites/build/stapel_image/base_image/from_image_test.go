package base_image_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v3/test/pkg/utils"
)

var _ = Describe("from anywhere", func() {
	BeforeEach(func() {
		SuiteData.TestDirPath = utils.FixturePath("from_anywhere")
	})

	It("should resolve and chain correctly", func(ctx SpecContext) {
		out := utils.SucceedCommandOutputString(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build")
		Expect(out).To(ContainSubstring("Building stage FromImage/from"))
		Expect(out).To(ContainSubstring("Building stage FromImageAlias/from"))
		Expect(out).To(ContainSubstring("Building stage FromAnother/from"))
	})
})
