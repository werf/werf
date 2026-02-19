package cyclonedxutil

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

var _ = Describe("CycloneDX Schema Validation", func() {
	Describe("ValidateCycloneDX16Schema", func() {
		DescribeTable("validates JSON against schema",
			func(jsonStr string, matcher types.GomegaMatcher) {
				err := ValidateCycloneDX16Schema([]byte(jsonStr))
				Expect(err).To(matcher)
			},

			Entry("empty JSON",
				`   `,
				HaveOccurred()),

			Entry("invalid JSON syntax",
				`{"bomFormat":`,
				HaveOccurred()),

			Entry("valid minimal BOM",
				`{"bomFormat":"CycloneDX","specVersion":"1.6","version":1}`,
				Succeed()),

			Entry("invalid bomFormat",
				`{"bomFormat":"SPDX","specVersion":"1.6","version":1}`,
				MatchError(ContainSubstring("bomFormat"))),

			Entry("invalid specVersion",
				`{"bomFormat":"CycloneDX","specVersion":"1.5","version":1}`,
				MatchError(ContainSubstring("specVersion"))),

			Entry("missing specVersion (zero value is not 1.6)",
				`{"bomFormat":"CycloneDX","version":1}`,
				MatchError(ContainSubstring("specVersion"))),

			Entry("version less than 1",
				`{"bomFormat":"CycloneDX","specVersion":"1.6","version":0}`,
				MatchError(ContainSubstring("version"))),

			Entry("invalid serialNumber",
				`{"bomFormat":"CycloneDX","specVersion":"1.6","version":1,"serialNumber":"not-a-uuid"}`,
				MatchError(ContainSubstring("serialNumber"))),

			Entry("invalid license format (flat)",
				`{
					"bomFormat": "CycloneDX",
					"specVersion": "1.6",
					"version": 1,
					"components": [
						{
							"type": "library",
							"name": "test",
							"licenses": [
								{ "id": "MIT" }
							]
						}
					]
				}`,
				MatchError(ContainSubstring("Must validate one and only one schema"))),
		)
	})
})
