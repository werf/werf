package gost

import (
	cdx "github.com/CycloneDX/cyclonedx-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

var _ = Describe("Gost SBOM setter", func() {
	DescribeTable("Set",
		func(bom *cdx.BOM, config Config, expectedComponents []cdx.Component, expectedErrMatcher OmegaMatcher) {
			err := Upsert(bom, config)
			Expect(err).To(expectedErrMatcher)
			if err != nil {
				return
			}
			if bom.Components != nil {
				Expect(lo.FromPtr(bom.Components)).To(Equal(expectedComponents))
			}
		},
		Entry("should fail if BOM is nil",
			nil, Config{}, nil, MatchError("BOM is required")),
		Entry("should add GOST properties if missing",
			&cdx.BOM{
				Components: &[]cdx.Component{{Name: "test"}},
			},
			Config{AttackSurface: GostValueYes, SecurityFunction: GostValueNo},
			[]cdx.Component{
				{
					Name: "test",
					Properties: &[]cdx.Property{
						{Name: PropertyAttackSurface, Value: "yes"},
						{Name: PropertySecurityFunction, Value: "no"},
					},
				},
			},
			Succeed()),
		Entry("should update existing GOST properties",
			&cdx.BOM{
				Components: &[]cdx.Component{
					{
						Name: "test",
						Properties: &[]cdx.Property{
							{Name: PropertyAttackSurface, Value: "no"},
							{Name: PropertySecurityFunction, Value: "no"},
						},
					},
				},
			},
			Config{AttackSurface: GostValueYes, SecurityFunction: GostValueYes},
			[]cdx.Component{
				{
					Name: "test",
					Properties: &[]cdx.Property{
						{Name: PropertyAttackSurface, Value: "yes"},
						{Name: PropertySecurityFunction, Value: "yes"},
					},
				},
			},
			Succeed()),
		Entry("should ignore undefined values in config during update",
			&cdx.BOM{
				Components: &[]cdx.Component{
					{
						Name: "test",
						Properties: &[]cdx.Property{
							{Name: PropertyAttackSurface, Value: "no"},
						},
					},
				},
			},
			Config{AttackSurface: GostValueUndefined, SecurityFunction: GostValueYes},
			[]cdx.Component{
				{
					Name: "test",
					Properties: &[]cdx.Property{
						{Name: PropertyAttackSurface, Value: "no"},
						{Name: PropertySecurityFunction, Value: "yes"},
					},
				},
			},
			Succeed()),
		Entry("should inject 'indirect' value",
			&cdx.BOM{
				Components: &[]cdx.Component{{Name: "test"}},
			},
			Config{AttackSurface: GostValueIndirect, SecurityFunction: GostValueIndirect},
			[]cdx.Component{
				{
					Name: "test",
					Properties: &[]cdx.Property{
						{Name: PropertyAttackSurface, Value: "indirect"},
						{Name: PropertySecurityFunction, Value: "indirect"},
					},
				},
			},
			Succeed()),
	)
})
