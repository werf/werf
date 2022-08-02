package image

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("InfoGetter", func() {
	DescribeTable("TestInfoGetter",
		func(data TestInfoGetter) {
			getter := NewInfoGetter(data.ImageName, data.Ref, data.Opts)

			Expect(getter.IsNameless()).To(Equal(data.ExpectIsNameless))
			Expect(getter.GetWerfImageName()).To(Equal(data.ExpectWerfImageName))
			Expect(getter.GetName()).To(Equal(data.ExpectName))
			Expect(getter.GetTag()).To(Equal(data.ExpectTag))
		},

		Entry("nameless image",
			TestInfoGetter{
				ImageName:           "",
				Ref:                 "myregistry.domain.com/group/project:abcd",
				Opts:                InfoGetterOptions{},
				ExpectIsNameless:    true,
				ExpectWerfImageName: "",
				ExpectName:          "myregistry.domain.com/group/project:abcd",
				ExpectTag:           "abcd",
			}),

		Entry("named image",
			TestInfoGetter{
				ImageName:           "backend",
				Ref:                 "myregistry.domain.com/group/project:abcd",
				Opts:                InfoGetterOptions{},
				ExpectIsNameless:    false,
				ExpectWerfImageName: "backend",
				ExpectName:          "myregistry.domain.com/group/project:abcd",
				ExpectTag:           "abcd",
			}),

		Entry("named image with custom tag",
			TestInfoGetter{
				ImageName: "backend",
				Ref:       "myregistry.domain.com/group/project:abcd",
				Opts: InfoGetterOptions{
					CustomTagFunc: func(werfImageName, tag string) string {
						return fmt.Sprintf("%s-%s", werfImageName, tag)
					},
				},
				ExpectIsNameless:    false,
				ExpectWerfImageName: "backend",
				ExpectName:          "myregistry.domain.com/group/project:backend-abcd",
				ExpectTag:           "backend-abcd",
			}),
	)
})

type TestInfoGetter struct {
	ImageName string
	Ref       string
	Opts      InfoGetterOptions

	ExpectIsNameless    bool
	ExpectWerfImageName string
	ExpectName          string
	ExpectTag           string
}
