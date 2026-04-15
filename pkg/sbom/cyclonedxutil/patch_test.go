package cyclonedxutil

import (
	cdx "github.com/CycloneDX/cyclonedx-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PatchComponents", func() {
	DescribeTable("applies patch to matching components",
		func(bom *cdx.BOM, match func(*cdx.Component) bool, expectedVersions []string) {
			PatchComponents(bom, match, func(c *cdx.Component) {
				c.Version = "patched"
			})

			if bom == nil || bom.Components == nil {
				return
			}
			for i, expected := range expectedVersions {
				Expect((*bom.Components)[i].Version).To(Equal(expected))
			}
		},
		Entry("patches matching components only", &cdx.BOM{
			Components: &[]cdx.Component{
				{Name: "a", Version: "old", Type: cdx.ComponentTypeLibrary},
				{Name: "b", Version: "old", Type: cdx.ComponentTypeApplication},
				{Name: "c", Version: "old", Type: cdx.ComponentTypeLibrary},
			},
		}, func(c *cdx.Component) bool {
			return c.Type == cdx.ComponentTypeLibrary
		}, []string{"patched", "old", "patched"}),

		Entry("no-op on nil BOM", (*cdx.BOM)(nil),
			func(c *cdx.Component) bool { return true },
			[]string(nil)),

		Entry("no-op on nil components", &cdx.BOM{},
			func(c *cdx.Component) bool { return true },
			[]string(nil)),

		Entry("no-op when nothing matches", &cdx.BOM{
			Components: &[]cdx.Component{
				{Name: "a", Version: "keep"},
			},
		}, func(c *cdx.Component) bool {
			return false
		}, []string{"keep"}),
	)
})
