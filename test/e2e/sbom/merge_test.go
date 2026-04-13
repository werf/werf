package e2e_build_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/build"
	"github.com/werf/werf/v2/test/pkg/report"
	"github.com/werf/werf/v2/test/pkg/werf"
)

var _ = Describe("Sbom merge", Label("e2e", "sbom", "merge", "simple"), func() {
	Describe("happy path", Label("simple"), func() {
		DescribeTable("should produce valid product SBOM",
			func(ctx SpecContext, testOpts simpleTestOptions, isprasFormat string) {
				By("initializing")
				setupEnv(testOpts.setupEnvOptions)

				repoDirname := "repo1"
				buildReportPath := filepath.Join(SuiteData.TmpDir, "merge-build-report.json")
				mappingPath := filepath.Join(SuiteData.TmpDir, "merge-mapping-"+isprasFormat+".json")
				outputPath := filepath.Join(SuiteData.TmpDir, "merge-output-"+isprasFormat+".json")

				By("preparing test repo")
				SuiteData.InitTestRepo(ctx, repoDirname, "state1")

				By("building images")
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
				reportProject := report.NewProjectWithReport(werfProject)
				_, buildReport := reportProject.BuildWithReport(ctx, buildReportPath, nil)

				By("creating merge input mapping")
				mapping := mergeInputMappingFromReport(buildReport)
				Expect(mapping).To(HaveLen(2), "fixture must have exactly 2 images: frontend and backend")
				writeMergeJSON(mappingPath, mapping)

				By("running sbom merge --ispras-format=" + isprasFormat)
				werfProject.SbomMerge(ctx, &werf.SbomMergeOptions{
					CommonOptions: werf.CommonOptions{
						ExtraArgs: []string{
							"--input", mappingPath,
							"--repo", os.Getenv("WERF_REPO"),
							"--ispras-format", isprasFormat,
							"--app-name", "test-app",
							"--app-version", "1.2.3",
							"--manufacturer", "TestCorp",
							"--output", outputPath,
						},
					},
				})

				By("parsing output SBOM")
				merged := readMergeBOM(outputPath)

				By("asserting metadata invariants")
				assertMergeMetadataInvariants(merged, "test-app", "1.2.3", "TestCorp")

				switch isprasFormat {
				case "oss":
					By("asserting oss-specific invariants")
					assertMergeOSSInvariants(merged)
				case "container":
					By("asserting container-specific invariants")
					assertMergeContainerInvariants(merged, []string{"frontend", "backend"})
				}
			},
			Entry("oss format with Vanilla Docker",
				simpleTestOptions{setupEnvOptions{
					ContainerBackendMode:        "vanilla-docker",
					WithLocalRepo:               true,
					WithStagedDockerfileBuilder: false,
				}},
				"oss",
			),
			Entry("container format with Vanilla Docker",
				simpleTestOptions{setupEnvOptions{
					ContainerBackendMode:        "vanilla-docker",
					WithLocalRepo:               true,
					WithStagedDockerfileBuilder: false,
				}},
				"container",
			),
		)
	})

	Describe("negative cases", Label("negative"), func() {
		DescribeTable("should fail with invalid input",
			func(ctx SpecContext, extraArgs []string) {
				setupEnv(setupEnvOptions{
					ContainerBackendMode: "vanilla-docker",
					WithLocalRepo:        true,
				})
				repoDirname := "repo1-neg"
				SuiteData.InitTestRepo(ctx, repoDirname, "state1")
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

				werfProject.SbomMerge(ctx, &werf.SbomMergeOptions{
					CommonOptions: werf.CommonOptions{
						ShouldFail: true,
						ExtraArgs:  extraArgs,
					},
				})
			},
			Entry("no flags at all", []string{}),
			Entry("invalid ispras-format",
				[]string{
					"--input", "/dev/null",
					"--repo", "localhost:5000/test",
					"--ispras-format", "invalid-format",
					"--app-name", "app",
					"--app-version", "1.0.0",
					"--manufacturer", "Corp",
				},
			),
			Entry("non-existent mapping file",
				[]string{
					"--input", "/does/not/exist/mapping.json",
					"--repo", "localhost:5000/test",
					"--ispras-format", "oss",
					"--app-name", "app",
					"--app-version", "1.0.0",
					"--manufacturer", "Corp",
				},
			),
		)

		It("fails on empty mapping JSON", func(ctx SpecContext) {
			setupEnv(setupEnvOptions{ContainerBackendMode: "vanilla-docker", WithLocalRepo: true})
			repoDirname := "repo1-neg-empty"
			SuiteData.InitTestRepo(ctx, repoDirname, "state1")
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

			mappingPath := filepath.Join(GinkgoT().TempDir(), "empty-mapping.json")
			writeMergeJSON(mappingPath, map[string]string{})

			werfProject.SbomMerge(ctx, &werf.SbomMergeOptions{
				CommonOptions: werf.CommonOptions{
					ShouldFail: true,
					ExtraArgs: []string{
						"--input", mappingPath,
						"--repo", os.Getenv("WERF_REPO"),
						"--ispras-format", "oss",
						"--app-name", "app",
						"--app-version", "1.0.0",
						"--manufacturer", "Corp",
					},
				},
			})
		})

		It("fails on invalid digest in mapping", func(ctx SpecContext) {
			setupEnv(setupEnvOptions{ContainerBackendMode: "vanilla-docker", WithLocalRepo: true})
			repoDirname := "repo1-neg-digest"
			SuiteData.InitTestRepo(ctx, repoDirname, "state1")
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

			mappingPath := filepath.Join(GinkgoT().TempDir(), "bad-digest-mapping.json")
			writeMergeJSON(mappingPath, map[string]string{"frontend": "not-a-valid-digest"})

			werfProject.SbomMerge(ctx, &werf.SbomMergeOptions{
				CommonOptions: werf.CommonOptions{
					ShouldFail: true,
					ExtraArgs: []string{
						"--input", mappingPath,
						"--repo", os.Getenv("WERF_REPO"),
						"--ispras-format", "oss",
						"--app-name", "app",
						"--app-version", "1.0.0",
						"--manufacturer", "Corp",
					},
				},
			})
		})
	})
})

func assertMergeMetadataInvariants(bom *cdx.BOM, appName, appVersion, manufacturer string) {
	Expect(bom.BOMFormat).To(Equal("CycloneDX"))
	Expect(bom.SpecVersion).To(Equal(cdx.SpecVersion1_6))
	Expect(bom.Metadata).NotTo(BeNil())
	Expect(bom.Metadata.Manufacturer).To(BeNil(),
		"manufacturer must not appear at metadata root level — only inside metadata.component")

	comp := bom.Metadata.Component
	Expect(comp).NotTo(BeNil())
	Expect(comp.BOMRef).To(BeEmpty(), "metadata.component must not have bom-ref")
	Expect(comp.Type).To(Equal(cdx.ComponentTypeApplication))
	Expect(comp.Name).To(Equal(appName))
	Expect(comp.Version).To(Equal(appVersion))
	Expect(comp.Manufacturer).NotTo(BeNil())
	Expect(comp.Manufacturer.Name).To(Equal(manufacturer))
	Expect(mergeGOSTPropertiesOf(comp)).To(BeEmpty(),
		"metadata.component must not have GOST properties")
}

func assertMergeOSSInvariants(bom *cdx.BOM) {
	Expect(bom.Components).NotTo(BeNil())
	components := *bom.Components
	Expect(components).NotTo(BeEmpty())

	for _, c := range components {
		Expect(c.Type).NotTo(Equal(cdx.ComponentTypeContainer),
			"oss format must not contain container-type components at top level")
	}

	lodashCount := 0
	for _, c := range components {
		if c.Name == "lodash" {
			lodashCount++
		}
	}
	Expect(lodashCount).To(Equal(1),
		"lodash appears in both images but must be deduplicated to 1 entry in oss format")

	Expect(bom.Dependencies).NotTo(BeNil(), "dependencies must be preserved after merge")
	Expect(*bom.Dependencies).NotTo(BeEmpty())
}

func assertMergeContainerInvariants(bom *cdx.BOM, expectedImages []string) {
	Expect(bom.Components).NotTo(BeNil())
	containers := *bom.Components
	Expect(containers).To(HaveLen(len(expectedImages)),
		"container format must have exactly one top-level component per image")

	names := make([]string, 0, len(containers))
	for _, c := range containers {
		Expect(c.Type).To(Equal(cdx.ComponentTypeContainer),
			"all top-level components in container format must be type=container")
		Expect(c.Components).NotTo(BeNil(),
			"each container must have nested components")
		Expect(*c.Components).NotTo(BeEmpty())
		names = append(names, c.Name)
	}

	for _, img := range expectedImages {
		Expect(names).To(ContainElement(img),
			"container component for image %q must be present", img)
	}

	Expect(bom.Dependencies).NotTo(BeNil(), "dependencies must be preserved after merge")
	Expect(*bom.Dependencies).NotTo(BeEmpty())
}

func mergeInputMappingFromReport(buildReport build.ImagesReport) map[string]string {
	mapping := make(map[string]string, len(buildReport.Images))
	for imageName, record := range buildReport.Images {
		digest := record.DockerImageDigest
		if !strings.HasPrefix(digest, "sha256:") {
			digest = "sha256:" + digest
		}
		mapping[imageName] = digest
	}
	return mapping
}

func writeMergeJSON(path string, v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	Expect(err).NotTo(HaveOccurred())
	Expect(os.WriteFile(path, data, 0o644)).To(Succeed())
}

func readMergeBOM(path string) *cdx.BOM {
	data, err := os.ReadFile(path)
	Expect(err).NotTo(HaveOccurred())
	var bom cdx.BOM
	Expect(json.Unmarshal(data, &bom)).To(Succeed())
	return &bom
}

func mergeGOSTPropertiesOf(comp *cdx.Component) []cdx.Property {
	if comp.Properties == nil {
		return nil
	}
	var result []cdx.Property
	for _, p := range *comp.Properties {
		if strings.HasPrefix(p.Name, "GOST:") {
			result = append(result, p)
		}
	}
	return result
}
