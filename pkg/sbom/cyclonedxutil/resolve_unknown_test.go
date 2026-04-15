package cyclonedxutil

import (
	cdx "github.com/CycloneDX/cyclonedx-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ResolveUnknownGoVersions", func() {
	type testCase struct {
		inputBOM         *cdx.BOM
		version          string
		mainModule       string
		replaceTargets   []string
		replacePaths     []string
		checkIdentity    bool
		expectedNames    []string
		expectedVersions []string
		expectedPURLs    []string
	}

	DescribeTable("version resolution",
		func(tc testCase) {
			result := ResolveUnknownGoVersions(tc.inputBOM, tc.version, tc.mainModule, tc.replaceTargets, tc.replacePaths)

			if tc.checkIdentity {
				Expect(result).To(BeIdenticalTo(tc.inputBOM))
			}

			if tc.inputBOM.Components != nil {
				for i, expectedName := range tc.expectedNames {
					Expect((*result.Components)[i].Name).To(Equal(expectedName))
				}
				for i, expectedVersion := range tc.expectedVersions {
					Expect((*result.Components)[i].Version).To(Equal(expectedVersion))
				}
				for i, expectedPURL := range tc.expectedPURLs {
					if expectedPURL != "" {
						Expect((*result.Components)[i].PackageURL).To(Equal(expectedPURL))
					}
				}
			} else {
				Expect(result.Components).To(BeNil())
			}
		},
		Entry("updates main module UNKNOWN version", testCase{
			inputBOM: &cdx.BOM{
				Components: &[]cdx.Component{{
					Name:    "example.com/module",
					Version: "UNKNOWN",
					Type:    cdx.ComponentTypeLibrary,
				}},
			},
			version:          "v1.2.3",
			mainModule:       "example.com/module",
			replaceTargets:   nil,
			checkIdentity:    true,
			expectedVersions: []string{"v1.2.3"},
		}),
		Entry("updates local replace target", testCase{
			inputBOM: &cdx.BOM{
				Components: &[]cdx.Component{{
					Name:    "example.com/replaced",
					Version: "UNKNOWN",
					Type:    cdx.ComponentTypeLibrary,
				}},
			},
			version:          "v1.0.0",
			mainModule:       "example.com/other",
			replaceTargets:   []string{"example.com/replaced"},
			expectedVersions: []string{"v1.0.0"},
		}),
		Entry("ignores non-matching module", testCase{
			inputBOM: &cdx.BOM{
				Components: &[]cdx.Component{{
					Name:    "example.com/other",
					Version: "UNKNOWN",
					Type:    cdx.ComponentTypeLibrary,
				}},
			},
			version:          "v2.0.0",
			mainModule:       "example.com/module",
			replaceTargets:   []string{"example.com/replaced"},
			expectedVersions: []string{"UNKNOWN"},
		}),
		Entry("does not modify components with real versions", testCase{
			inputBOM: &cdx.BOM{
				Components: &[]cdx.Component{{
					Name:    "example.com/module",
					Version: "v0.9.0",
					Type:    cdx.ComponentTypeLibrary,
				}},
			},
			version:          "v1.0.0",
			mainModule:       "example.com/module",
			replaceTargets:   nil,
			expectedVersions: []string{"v0.9.0"},
		}),
		Entry("updates command-line-arguments module", testCase{
			inputBOM: &cdx.BOM{
				Components: &[]cdx.Component{{
					Name:    "command-line-arguments",
					Version: "UNKNOWN",
					Type:    cdx.ComponentTypeLibrary,
				}},
			},
			version:          "v0.0.1",
			mainModule:       "example.com/module",
			replaceTargets:   nil,
			expectedVersions: []string{"v0.0.1"},
		}),
		Entry("updates PURL when UNKNOWN is present", testCase{
			inputBOM: &cdx.BOM{
				Components: &[]cdx.Component{{
					Name:       "example.com/module",
					Version:    "UNKNOWN",
					PackageURL: "pkg:golang/example.com/module@UNKNOWN",
					Type:       cdx.ComponentTypeLibrary,
				}},
			},
			version:          "v1.2.3",
			mainModule:       "example.com/module",
			replaceTargets:   nil,
			expectedVersions: []string{"v1.2.3"},
			expectedPURLs:    []string{"pkg:golang/example.com/module@v1.2.3"},
		}),
		Entry("adds version to PURL without version", testCase{
			inputBOM: &cdx.BOM{
				Components: &[]cdx.Component{{
					Name:       "example.com/module",
					Version:    "UNKNOWN",
					PackageURL: "pkg:golang/example.com/module",
					Type:       cdx.ComponentTypeLibrary,
				}},
			},
			version:          "v1.2.3",
			mainModule:       "example.com/module",
			replaceTargets:   nil,
			expectedVersions: []string{"v1.2.3"},
			expectedPURLs:    []string{"pkg:golang/example.com/module@v1.2.3"},
		}),
		Entry("handles nil components", testCase{
			inputBOM:       &cdx.BOM{},
			version:        "v1.0.0",
			mainModule:     "example.com/module",
			replaceTargets: nil,
			checkIdentity:  true,
		}),
		Entry("skips updates on empty version", testCase{
			inputBOM: &cdx.BOM{
				Components: &[]cdx.Component{{
					Name:    "example.com/module",
					Version: "UNKNOWN",
					Type:    cdx.ComponentTypeLibrary,
				}},
			},
			version:          "",
			mainModule:       "example.com/module",
			replaceTargets:   nil,
			expectedVersions: []string{"UNKNOWN"},
		}),
		Entry("resolves (devel) version same as UNKNOWN", testCase{
			inputBOM: &cdx.BOM{
				Components: &[]cdx.Component{{
					Name:    "example.com/module",
					Version: "(devel)",
					Type:    cdx.ComponentTypeLibrary,
				}},
			},
			version:          "v1.0.0",
			mainModule:       "example.com/module",
			expectedVersions: []string{"v1.0.0"},
		}),
		Entry("resolves component matched by local replace path and fixes name/purl", testCase{
			inputBOM: &cdx.BOM{
				Components: &[]cdx.Component{{
					Name:       "./mylib",
					Version:    "(devel)",
					PackageURL: "pkg:golang/./mylib@%28devel%29",
					Type:       cdx.ComponentTypeLibrary,
				}},
			},
			version:          "v2.0.0",
			mainModule:       "example.com/app",
			replaceTargets:   []string{"example.com/mylib"},
			replacePaths:     []string{"./mylib"},
			expectedNames:    []string{"example.com/mylib"},
			expectedVersions: []string{"v2.0.0"},
			expectedPURLs:    []string{"pkg:golang/example.com/mylib@v2.0.0"},
		}),
		Entry("resolves URL-encoded (devel) in PURL", testCase{
			inputBOM: &cdx.BOM{
				Components: &[]cdx.Component{{
					Name:       "example.com/module",
					Version:    "(devel)",
					PackageURL: "pkg:golang/example.com/module@%28devel%29",
					Type:       cdx.ComponentTypeLibrary,
				}},
			},
			version:          "v1.5.0",
			mainModule:       "example.com/module",
			expectedVersions: []string{"v1.5.0"},
			expectedPURLs:    []string{"pkg:golang/example.com/module@v1.5.0"},
		}),
		Entry("updates only UNKNOWN library components", testCase{
			inputBOM: &cdx.BOM{
				Components: &[]cdx.Component{
					{
						Name:    "example.com/module",
						Version: "UNKNOWN",
						Type:    cdx.ComponentTypeLibrary,
					},
					{
						Name:    "example.com/module",
						Version: "UNKNOWN",
						Type:    cdx.ComponentTypeApplication,
					},
					{
						Name:    "example.com/module",
						Version: "v1.0.0",
						Type:    cdx.ComponentTypeLibrary,
					},
				},
			},
			version:          "v2.0.0",
			mainModule:       "example.com/module",
			replaceTargets:   nil,
			expectedVersions: []string{"v2.0.0", "UNKNOWN", "v1.0.0"},
		}),
	)
})
