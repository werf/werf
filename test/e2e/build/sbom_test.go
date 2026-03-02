package e2e_build_test

import (
	"archive/tar"
	"fmt"
	"io"

	cdx "github.com/CycloneDX/cyclonedx-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	imagePkg "github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/sbom/cyclonedxutil"
	"github.com/werf/werf/v2/pkg/sbom/extract"
	sbomImage "github.com/werf/werf/v2/pkg/sbom/image"
	"github.com/werf/werf/v2/test/pkg/contback"
	"github.com/werf/werf/v2/test/pkg/report"
	"github.com/werf/werf/v2/test/pkg/utils"
	"github.com/werf/werf/v2/test/pkg/werf"
)

var _ = Describe("Simple build", Label("e2e", "build", "sbom", "simple"), func() {
	DescribeTable("should generate and store SBOM as an image",
		func(ctx SpecContext, testOpts simpleTestOptions) {
			By("initializing")
			setupEnv(testOpts.setupEnvOptions)

			contRuntime, err := contback.NewContainerBackend(testOpts.ContainerBackendMode)
			if err == contback.ErrRuntimeUnavailable {
				Skip(err.Error())
			} else if err != nil {
				Fail(err.Error())
			}

			By("state0: case", func() {
				repoDirname := "repo0"
				fixtureRelPath := "sbom/state0"
				buildReportName := "report0.json"

				By("state0: preparing test repo")
				SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)

				By("state0: building images")
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
				reportProject := report.NewProjectWithReport(werfProject)
				buildOut, buildReport := reportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath(buildReportName), nil)
				Expect(buildOut).To(ContainSubstring("Building stage"))

				By("state0: SBOM logging output")
				Expect(buildOut).To(ContainSubstring("SBOM"))
				Expect(buildOut).To(ContainSubstring("Scan image"))
				Expect(buildOut).To(ContainSubstring("Build destination image"))

				for builtImgName, reportRecord := range buildReport.Images {
					By(fmt.Sprintf("state0: validate result for %q", builtImgName))
					{
						By("state0: SBOM image metadata")
						imgInspect := contRuntime.GetImageInspect(ctx, reportRecord.DockerImageName)
						sbomImgInspect := contRuntime.GetImageInspect(ctx, sbomImage.ImageName(reportRecord.DockerImageName))

						// shared labels
						Expect(sbomImgInspect.Config.Labels[imagePkg.WerfLabel]).To(Equal(imgInspect.Config.Labels[imagePkg.WerfLabel]))
						Expect(sbomImgInspect.Config.Labels[imagePkg.WerfVersionLabel]).To(Equal(imgInspect.Config.Labels[imagePkg.WerfVersionLabel]))
						Expect(sbomImgInspect.Config.Labels[imagePkg.WerfProjectRepoCommitLabel]).To(Equal(imgInspect.Config.Labels[imagePkg.WerfProjectRepoCommitLabel]))
						Expect(sbomImgInspect.Config.Labels[imagePkg.WerfStageContentDigestLabel]).To(Equal(imgInspect.Config.Labels[imagePkg.WerfStageContentDigestLabel]))
						// sbom labels
						Expect(sbomImgInspect.Config.Labels[imagePkg.WerfSbomLabel]).To(HavePrefix("0c15bc4e5bd8541138b5b6b7065eb8f641284b4913878d953be46419f50e8ebc"))

						By("state0: SBOM image file system layout")
						opener := func() (io.ReadCloser, error) {
							return contRuntime.SaveImageToStream(ctx, sbomImage.ImageName(reportRecord.DockerImageName)), nil
						}

						flattenedFsStreamReaderCloser, err := extract.FromImageStream(opener)
						Expect(err).To(Succeed(), "should extract SBOM image from the stream")

						var actualFilePaths []string
						err = utils.ForEachInTarball(tar.NewReader(flattenedFsStreamReaderCloser), func(header *tar.Header) error {
							actualFilePaths = append(actualFilePaths, header.Name)
							return nil
						})
						Expect(err).To(Succeed(), "should iterate over the tarball entries")
						Expect(flattenedFsStreamReaderCloser.Close()).To(Succeed(), "should close the stream reader")

						expectedFilePaths := []string{
							"sbom",
							"sbom/cyclonedx@1.6",
							"sbom/cyclonedx@1.6/70ee6b0600f471718988bc123475a625ecd4a5763059c62802ae6280e65f5623.json",
						}
						Expect(actualFilePaths).To(Equal(expectedFilePaths))

						By("state0: extracting and validating SBOM content")
						bom := extractBOMFromSbomImage(ctx, contRuntime, reportRecord.DockerImageName)

						By("state0: verifying SBOM structure")
						Expect(bom.BOMFormat).To(Equal("CycloneDX"))
						Expect(bom.SpecVersion).To(Equal(cdx.SpecVersion1_6))
						Expect(bom.Version).To(Equal(1))
						Expect(bom.SerialNumber).To(HavePrefix("urn:uuid:"))
						Expect(bom.Components).NotTo(BeNil())

						By("state0: verifying curl component with strict field validation")
						components := *bom.Components
						curlComponent := findComponentByName(components, "curl")
						assertComponentEquals(curlComponent, expectedComponent{
							Name:     "curl",
							Type:     cdx.ComponentTypeApplication,
							Version:  "8.12.1",
							PURL:     "pkg:generic/curl@8.12.1",
							Licenses: []string{"MIT"},
							Hashes: map[cdx.HashAlgorithm]string{
								cdx.HashAlgoSHA256: "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
							},
						})

						By("state0: verifying components count")
						Expect(len(components)).To(Equal(1), "SBOM should contain exactly 1 component (curl from fragment)")
					}
				}
			})
		},
		Entry("without repo using Vanilla Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "vanilla-docker",
			WithLocalRepo:               false,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("with local repo using Vanilla Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "vanilla-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("without repo using BuildKit Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "buildkit-docker",
			WithLocalRepo:               false,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("with local repo using BuildKit Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "buildkit-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		// TODO (zaytsev): it does not work currently
		// https://github.com/werf/werf/actions/runs/15076648086/job/42385521980?pr=6860#step:11:150
		XEntry("with local repo using Native Buildah with rootless isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-rootless",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		// TODO: "werf purge --project-name=..." is not implemented for Buildah. So we have potential risk to fail the test.
		XEntry("with local repo using Native Buildah with chroot isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-chroot",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}, FlakeAttempts(5)),
	)

	DescribeTable("should succeed when base image SBOM is scratch",
		func(ctx SpecContext, testOpts simpleTestOptions) {
			By("initializing")
			setupEnv(testOpts.setupEnvOptions)

			By("preparing test repo")
			repoDirname := "repo_base_sbom"
			fixtureRelPath := "sbom/state1"
			SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)

			By("building images")
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
			reportProject := report.NewProjectWithReport(werfProject)
			buildOut, _ := reportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath("report_base_sbom.json"), nil)

			Expect(buildOut).To(ContainSubstring("SBOM processing"))
			Expect(buildOut).To(ContainSubstring("image stapel"))
		},
		Entry("with local repo using Vanilla Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "vanilla-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("with local repo using BuildKit Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "buildkit-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		XEntry("with local repo using Native Buildah with chroot isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-chroot",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		XEntry("with local repo using Native Buildah with rootless isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-rootless",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
	)

	DescribeTable("should fail when base image SBOM is not found in registry",
		func(ctx SpecContext, testOpts simpleTestOptions) {
			By("initializing")
			setupEnv(testOpts.setupEnvOptions)

			By("preparing test repo")
			repoDirname := "repo_base_sbom"
			fixtureRelPath := "sbom/state2"
			SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)

			By("building images (expecting failure)")
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
			out, err := werfProject.BuildWithErr(ctx, nil)

			Expect(err).To(HaveOccurred(), "build should fail when base image SBOM is not found")
			Expect(out).To(ContainSubstring("unable to get base image"))
		},
		Entry("with local repo using Vanilla Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "vanilla-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("with local repo using BuildKit Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "buildkit-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		XEntry("with local repo using Native Buildah with chroot isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-chroot",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		XEntry("with local repo using Native Buildah with rootless isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-rootless",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
	)

	DescribeTable("should succeed and store SBOMs",
		func(ctx SpecContext, testOpts simpleTestOptions) {
			By("initializing")
			setupEnv(testOpts.setupEnvOptions)

			By("preparing test repo")
			repoDirName := "repo_import_sbom"
			fixtureRelPath := "sbom/import_stapel/state0"
			SuiteData.InitTestRepo(ctx, repoDirName, fixtureRelPath)

			By("building images")
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirName))
			reportProject := report.NewProjectWithReport(werfProject)
			buildOut, _ := reportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath("report_import_sbom.json"), nil)

			Expect(buildOut).To(ContainSubstring("SBOM processing"))
			Expect(buildOut).To(ContainSubstring("image stapel-scratch-based"))
			Expect(buildOut).To(ContainSubstring("image stapel"))
		},
		Entry("with local repo using Vanilla Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "vanilla-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("with local repo using BuildKit Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "buildkit-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		XEntry("with local repo using Native Buildah with chroot isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-chroot",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		XEntry("with local repo using Native Buildah with rootless isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-rootless",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
	)

	DescribeTable("should fail when import image not found",
		func(ctx SpecContext, testOpts simpleTestOptions) {
			By("initializing")
			setupEnv(testOpts.setupEnvOptions)

			By("preparing test repo")
			repoDirName := "repo_import_sbom"
			fixtureRelPath := "sbom/import_stapel/state1"
			SuiteData.InitTestRepo(ctx, repoDirName, fixtureRelPath)

			By("building images")
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirName))
			out, err := werfProject.BuildWithErr(ctx, nil)

			Expect(err).To(HaveOccurred(), "build should fail when base image SBOM is not found")
			Expect(out).To(ContainSubstring("unable to get import image sbom for"))
		},
		Entry("with local repo using Vanilla Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "vanilla-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("with local repo using BuildKit Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "buildkit-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		XEntry("with local repo using Native Buildah with chroot isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-chroot",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		XEntry("with local repo using Native Buildah with rootless isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-rootless",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
	)
})

var _ = Describe("SBOM merge", Label("e2e", "build", "sbom", "merge", "simple"), func() {
	DescribeTable("should merge base image SBOM with fragment",
		func(ctx SpecContext, testOpts simpleTestOptions) {
			By("initializing")
			setupEnv(testOpts.setupEnvOptions)

			contRuntime, err := contback.NewContainerBackend(testOpts.ContainerBackendMode)
			if err == contback.ErrRuntimeUnavailable {
				Skip(err.Error())
			} else if err != nil {
				Fail(err.Error())
			}

			By("preparing test repo")
			repoDirname := "repo_merge_base_fragment"
			fixtureRelPath := "sbom/merge_base_fragment"
			SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)

			By("building images")
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
			reportProject := report.NewProjectWithReport(werfProject)
			buildOut, buildReport := reportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath("report_merge_base_fragment.json"), nil)

			Expect(buildOut).To(ContainSubstring("SBOM processing"))

			By("extracting and verifying merged SBOM")
			for imgName, reportRecord := range buildReport.Images {
				if imgName != "app" {
					continue
				}

				bom := extractBOMFromSbomImage(ctx, contRuntime, reportRecord.DockerImageName)

				By("verifying SBOM structure")
				Expect(bom.BOMFormat).To(Equal("CycloneDX"))
				Expect(bom.SpecVersion).To(Equal(cdx.SpecVersion1_6))
				Expect(bom.Version).To(Equal(1))
				Expect(bom.SerialNumber).To(HavePrefix("urn:uuid:"))
				Expect(bom.Components).NotTo(BeNil())

				By("verifying custom-component with strict field validation")
				components := *bom.Components
				customComponent := findComponentByName(components, "custom-component")
				assertComponentEquals(customComponent, expectedComponent{
					Name:    "custom-component",
					Type:    cdx.ComponentTypeApplication,
					Version: "1.0.0",
					PURL:    "pkg:generic/custom-component@1.0.0",
				})

				By("verifying components count (fragment only, scratch has no components)")
				Expect(len(components)).To(Equal(3),
					"merged SBOM should contain exactly 3 components from fragment (custom-component + dep-a + dep-b)")

				By("verifying Services with strict field validation")
				Expect(bom.Services).NotTo(BeNil(), "SBOM should have services")
				services := *bom.Services
				customService := findServiceByName(services, "custom-api-service")
				assertServiceEquals(customService, expectedService{
					Name:          "custom-api-service",
					Endpoints:     []string{"https://api.example.com/v1"},
					Authenticated: boolPtr(true),
				})
				Expect(len(services)).To(Equal(1), "should have exactly 1 service")

				By("verifying ExternalReferences with strict field validation")
				Expect(bom.ExternalReferences).NotTo(BeNil(), "SBOM should have external references")
				refs := *bom.ExternalReferences
				websiteRef := findExternalReferenceByURL(refs, "https://example.com")
				assertExternalReferenceEquals(websiteRef, expectedExternalReference{
					Type: cdx.ERTypeWebsite,
					URL:  "https://example.com",
				})
				docsRef := findExternalReferenceByURL(refs, "https://docs.example.com")
				assertExternalReferenceEquals(docsRef, expectedExternalReference{
					Type: cdx.ERTypeDocumentation,
					URL:  "https://docs.example.com",
				})
				Expect(len(refs)).To(Equal(2), "should have exactly 2 external references")

				By("verifying Properties with strict field validation")
				Expect(bom.Properties).NotTo(BeNil(), "SBOM should have properties")
				props := *bom.Properties
				buildEnvProp := findPropertyByName(props, "build-environment")
				assertPropertyEquals(buildEnvProp, expectedProperty{
					Name:  "build-environment",
					Value: "production",
				})
				customProp := findPropertyByName(props, "custom-property")
				assertPropertyEquals(customProp, expectedProperty{
					Name:  "custom-property",
					Value: "custom-value",
				})
				Expect(len(props)).To(Equal(2), "should have exactly 2 properties")

				By("verifying Dependencies with strict field validation")
				Expect(bom.Dependencies).NotTo(BeNil(), "SBOM should have dependencies")
				deps := *bom.Dependencies
				customDep := findDependencyByRef(deps, "custom-component")
				assertDependencyEquals(customDep, expectedDependency{
					Ref:       "custom-component",
					DependsOn: []string{"dep-a", "dep-b"},
				})
				Expect(len(deps)).To(Equal(1), "should have exactly 1 dependency")

				By("verifying Annotations are present")
				Expect(bom.Annotations).NotTo(BeNil(), "SBOM should have annotations")
				annotations := *bom.Annotations
				Expect(findAnnotationByText(annotations, "This is a test annotation for merge verification")).NotTo(BeNil(),
					"fragment annotation should be present")
				Expect(len(annotations)).To(Equal(1), "should have exactly 1 annotation")
			}
		},
		Entry("with local repo using Vanilla Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "vanilla-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("with local repo using BuildKit Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "buildkit-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		XEntry("with local repo using Native Buildah with chroot isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-chroot",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		XEntry("with local repo using Native Buildah with rootless isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-rootless",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
	)

	DescribeTable("should merge derived image SBOM with base",
		func(ctx SpecContext, testOpts simpleTestOptions) {
			By("initializing")
			setupEnv(testOpts.setupEnvOptions)

			contRuntime, err := contback.NewContainerBackend(testOpts.ContainerBackendMode)
			if err == contback.ErrRuntimeUnavailable {
				Skip(err.Error())
			} else if err != nil {
				Fail(err.Error())
			}

			By("preparing test repo")
			repoDirname := "repo_merge_derived_with_base"
			fixtureRelPath := "sbom/merge_derived_with_base_fragment"
			SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)

			By("building images")
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
			reportProject := report.NewProjectWithReport(werfProject)
			buildOut, buildReport := reportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath("report_merge_derived_with_base.json"), nil)

			Expect(buildOut).To(ContainSubstring("SBOM processing"))

			expectedCurl := expectedComponent{
				Name:     "curl",
				Type:     cdx.ComponentTypeApplication,
				Version:  "8.12.1",
				PURL:     "pkg:generic/curl@8.12.1",
				Licenses: []string{"Apache-2.0"},
				Hashes: map[cdx.HashAlgorithm]string{
					cdx.HashAlgoSHA256: "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
				},
			}

			By("extracting and verifying SBOM for base-level-0")
			baseReportRecord, ok := buildReport.Images["base-level-0"]
			Expect(ok).To(BeTrue(), "base-level-0 should be in build report")

			baseBom := extractBOMFromSbomImage(ctx, contRuntime, baseReportRecord.DockerImageName)

			By("verifying base-level-0 SBOM structure")
			Expect(baseBom.BOMFormat).To(Equal("CycloneDX"))
			Expect(baseBom.SpecVersion).To(Equal(cdx.SpecVersion1_6))
			Expect(baseBom.Version).To(Equal(1))
			Expect(baseBom.SerialNumber).To(HavePrefix("urn:uuid:"))
			Expect(baseBom.Components).NotTo(BeNil())

			By("verifying base-level-0 curl component with strict field validation")
			baseComponents := *baseBom.Components
			baseCurlComponent := findComponentByName(baseComponents, "curl")
			assertComponentEquals(baseCurlComponent, expectedCurl)
			Expect(len(baseComponents)).To(Equal(2), "base-level-0 should have exactly 2 components (curl + libcurl)")

			By("verifying base-level-0 dependencies")
			Expect(baseBom.Dependencies).NotTo(BeNil(), "base-level-0 SBOM should have dependencies")
			baseDeps := *baseBom.Dependencies
			baseCurlDep := findDependencyByRef(baseDeps, "curl")
			assertDependencyEquals(baseCurlDep, expectedDependency{
				Ref:       "curl",
				DependsOn: []string{"libcurl"},
			})
			Expect(len(baseDeps)).To(Equal(1), "base-level-0 should have exactly 1 dependency")

			By("extracting and verifying SBOM for derived-level-1")
			derivedReportRecord, ok := buildReport.Images["derived-level-1"]
			Expect(ok).To(BeTrue(), "derived-level-1 should be in build report")

			derivedBom := extractBOMFromSbomImage(ctx, contRuntime, derivedReportRecord.DockerImageName)

			By("verifying derived-level-1 SBOM structure")
			Expect(derivedBom.BOMFormat).To(Equal("CycloneDX"))
			Expect(derivedBom.SpecVersion).To(Equal(cdx.SpecVersion1_6))
			Expect(derivedBom.Version).To(Equal(1))
			Expect(derivedBom.SerialNumber).To(HavePrefix("urn:uuid:"))
			Expect(derivedBom.Components).NotTo(BeNil())

			By("verifying derived-level-1 has curl component inherited from base-level-0 with same values")
			derivedComponents := *derivedBom.Components
			derivedCurlComponent := findComponentByName(derivedComponents, "curl")
			assertComponentEquals(derivedCurlComponent, expectedCurl)

			By("verifying derived-level-1 components count (curl + libcurl from base, empty fragment)")
			Expect(len(derivedComponents)).To(Equal(2),
				"derived-level-1 SBOM should contain exactly 2 components (curl + libcurl inherited from base)")

			By("verifying derived-level-1 dependencies inherited from base")
			Expect(derivedBom.Dependencies).NotTo(BeNil(), "derived-level-1 SBOM should have dependencies inherited from base")
			derivedDeps := *derivedBom.Dependencies
			derivedCurlDep := findDependencyByRef(derivedDeps, "curl")
			assertDependencyEquals(derivedCurlDep, expectedDependency{
				Ref:       "curl",
				DependsOn: []string{"libcurl"},
			})
			Expect(len(derivedDeps)).To(Equal(1), "derived-level-1 should have exactly 1 dependency inherited from base")
		},
		Entry("with local repo using Vanilla Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "vanilla-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("with local repo using BuildKit Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "buildkit-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		XEntry("with local repo using Native Buildah with chroot isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-chroot",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		XEntry("with local repo using Native Buildah with rootless isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-rootless",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
	)

	DescribeTable("should merge all SBOM sources (base + import + fragment)",
		func(ctx SpecContext, testOpts simpleTestOptions) {
			By("initializing")
			setupEnv(testOpts.setupEnvOptions)

			contRuntime, err := contback.NewContainerBackend(testOpts.ContainerBackendMode)
			if err == contback.ErrRuntimeUnavailable {
				Skip(err.Error())
			} else if err != nil {
				Fail(err.Error())
			}

			By("preparing test repo")
			repoDirname := "repo_merge_full"
			fixtureRelPath := "sbom/merge_full"
			SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)

			By("building images")
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
			reportProject := report.NewProjectWithReport(werfProject)
			buildOut, buildReport := reportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath("report_merge_full.json"), nil)

			Expect(buildOut).To(ContainSubstring("SBOM processing"))

			By("extracting and verifying merged SBOM for app image")
			for imgName, reportRecord := range buildReport.Images {
				if imgName != "app" {
					continue
				}

				bom := extractBOMFromSbomImage(ctx, contRuntime, reportRecord.DockerImageName)

				By("verifying SBOM structure")
				Expect(bom.BOMFormat).To(Equal("CycloneDX"))
				Expect(bom.SpecVersion).To(Equal(cdx.SpecVersion1_6))
				Expect(bom.Version).To(Equal(1))
				Expect(bom.SerialNumber).To(HavePrefix("urn:uuid:"))
				Expect(bom.Components).NotTo(BeNil())

				components := *bom.Components

				By("verifying app-custom component from fragment with strict validation")
				appCustomComponent := findComponentByName(components, "app-custom")
				assertComponentEquals(appCustomComponent, expectedComponent{
					Name:    "app-custom",
					Type:    cdx.ComponentTypeApplication,
					Version: "2.0.0",
					PURL:    "pkg:generic/app-custom@2.0.0",
				})

				By("verifying builder-custom component from import with strict validation")
				builderCustomComponent := findComponentByName(components, "builder-custom")
				assertComponentEquals(builderCustomComponent, expectedComponent{
					Name:    "builder-custom",
					Type:    cdx.ComponentTypeLibrary,
					Version: "1.0.0",
					PURL:    "pkg:generic/builder-custom@1.0.0",
				})

				By("verifying components count (import + fragment)")
				Expect(len(components)).To(Equal(5),
					"merged SBOM should contain exactly 5 components (builder-custom + builder-dep from import + app-custom + app-dep-a + app-dep-b from fragment)")

				By("verifying metadata is present")
				Expect(bom.Metadata).NotTo(BeNil())

				By("verifying services from both builder and app with strict validation")
				Expect(bom.Services).NotTo(BeNil(), "SBOM should have services")
				services := *bom.Services

				builderService := findServiceByName(services, "builder-service")
				assertServiceEquals(builderService, expectedService{
					Name:      "builder-service",
					Endpoints: []string{"https://builder.example.com"},
				})

				appService := findServiceByName(services, "app-api-service")
				assertServiceEquals(appService, expectedService{
					Name:          "app-api-service",
					Endpoints:     []string{"https://app.example.com/api"},
					Authenticated: boolPtr(true),
				})

				Expect(len(services)).To(Equal(2), "should have exactly 2 services")

				By("verifying Properties from both sources with strict validation")
				Expect(bom.Properties).NotTo(BeNil(), "SBOM should have properties")
				props := *bom.Properties

				builderProp := findPropertyByName(props, "builder-property")
				assertPropertyEquals(builderProp, expectedProperty{
					Name:  "builder-property",
					Value: "builder-value",
				})

				builderStageProp := findPropertyByName(props, "builder-stage")
				assertPropertyEquals(builderStageProp, expectedProperty{
					Name:  "builder-stage",
					Value: "build",
				})

				appProp := findPropertyByName(props, "app-property")
				assertPropertyEquals(appProp, expectedProperty{
					Name:  "app-property",
					Value: "app-value",
				})

				envProp := findPropertyByName(props, "environment")
				assertPropertyEquals(envProp, expectedProperty{
					Name:  "environment",
					Value: "production",
				})

				Expect(len(props)).To(Equal(4), "should have exactly 4 properties")

				By("verifying ExternalReferences with strict validation")
				Expect(bom.ExternalReferences).NotTo(BeNil(), "SBOM should have external references")
				refs := *bom.ExternalReferences

				vcsRef := findExternalReferenceByURL(refs, "https://github.com/example/app")
				assertExternalReferenceEquals(vcsRef, expectedExternalReference{
					Type: cdx.ERTypeVCS,
					URL:  "https://github.com/example/app",
				})

				issueTrackerRef := findExternalReferenceByURL(refs, "https://github.com/example/app/issues")
				assertExternalReferenceEquals(issueTrackerRef, expectedExternalReference{
					Type: cdx.ERTypeIssueTracker,
					URL:  "https://github.com/example/app/issues",
				})

				Expect(len(refs)).To(Equal(2), "should have exactly 2 external references")

				By("verifying Dependencies from both builder and app with strict validation")
				Expect(bom.Dependencies).NotTo(BeNil(), "SBOM should have dependencies")
				deps := *bom.Dependencies

				builderDep := findDependencyByRef(deps, "builder-custom")
				assertDependencyEquals(builderDep, expectedDependency{
					Ref:       "builder-custom",
					DependsOn: []string{"builder-dep"},
				})

				appDep := findDependencyByRef(deps, "app-custom")
				assertDependencyEquals(appDep, expectedDependency{
					Ref:       "app-custom",
					DependsOn: []string{"app-dep-a", "app-dep-b"},
				})

				Expect(len(deps)).To(Equal(2), "should have exactly 2 dependencies (builder + app)")

				By("verifying Annotations are present")
				Expect(bom.Annotations).NotTo(BeNil(), "SBOM should have annotations")
				annotations := *bom.Annotations
				Expect(findAnnotationByText(annotations, "Application level annotation")).NotTo(BeNil(),
					"app annotation should be present")
				Expect(len(annotations)).To(Equal(1), "should have exactly 1 annotation")
			}
		},
		Entry("with local repo using Vanilla Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "vanilla-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("with local repo using BuildKit Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "buildkit-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		XEntry("with local repo using Native Buildah with chroot isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-chroot",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		XEntry("with local repo using Native Buildah with rootless isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-rootless",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
	)
})

var _ = Describe("SBOM cross-project merge", Label("e2e", "build", "sbom", "merge", "simple"), func() {
	DescribeTable("should merge SBOM from external werf image (cross-project)",
		func(ctx SpecContext, testOpts simpleTestOptions) {
			By("initializing")
			setupEnv(testOpts.setupEnvOptions)

			contRuntime, err := contback.NewContainerBackend(testOpts.ContainerBackendMode)
			if err == contback.ErrRuntimeUnavailable {
				Skip(err.Error())
			} else if err != nil {
				Fail(err.Error())
			}

			By("Step 1: building base project")
			baseRepoDirname := "repo_cross_project_base"
			baseFixtureRelPath := "sbom/cross_project/base"
			SuiteData.InitTestRepo(ctx, baseRepoDirname, baseFixtureRelPath)

			baseWerfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(baseRepoDirname))
			baseReportProject := report.NewProjectWithReport(baseWerfProject)
			baseBuildOut, baseBuildReport := baseReportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath("report_cross_project_base.json"), nil)

			Expect(baseBuildOut).To(ContainSubstring("SBOM processing"))

			baseReportRecord, ok := baseBuildReport.Images["base-level-0"]
			Expect(ok).To(BeTrue(), "base-level-0 should be in build report")

			By("Step 2: verifying base image SBOM")
			baseBom := extractBOMFromSbomImage(ctx, contRuntime, baseReportRecord.DockerImageName)
			Expect(baseBom.Components).NotTo(BeNil())
			baseComponents := *baseBom.Components
			baseCurlComponent := findComponentByName(baseComponents, "curl")
			Expect(baseCurlComponent).NotTo(BeNil(), "base image should have curl component")

			By("verifying base image dependencies")
			Expect(baseBom.Dependencies).NotTo(BeNil(), "base image should have dependencies")
			baseDeps := *baseBom.Dependencies
			baseCurlDep := findDependencyByRef(baseDeps, "curl")
			assertDependencyEquals(baseCurlDep, expectedDependency{
				Ref:       "curl",
				DependsOn: []string{"libcurl"},
			})

			By("Step 3: building derived project with BASE_IMAGE_REF env")
			derivedRepoDirname := "repo_cross_project_derived"
			derivedFixtureRelPath := "sbom/cross_project/derived"
			SuiteData.InitTestRepo(ctx, derivedRepoDirname, derivedFixtureRelPath)

			derivedWerfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(derivedRepoDirname))
			derivedReportProject := report.NewProjectWithReport(derivedWerfProject)

			derivedBuildOpts := &werf.WithReportOptions{
				CommonOptions: werf.CommonOptions{
					Envs: []string{
						fmt.Sprintf("BASE_IMAGE_REF=%s", baseReportRecord.DockerImageName),
					},
				},
			}
			derivedBuildOut, derivedBuildReport := derivedReportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath("report_cross_project_derived.json"), derivedBuildOpts)

			Expect(derivedBuildOut).To(ContainSubstring("SBOM processing"))

			derivedReportRecord, ok := derivedBuildReport.Images["derived-level-1"]
			Expect(ok).To(BeTrue(), "derived-level-1 should be in build report")

			By("Step 4: verifying derived image SBOM contains components from base")
			derivedBom := extractBOMFromSbomImage(ctx, contRuntime, derivedReportRecord.DockerImageName)

			Expect(derivedBom.BOMFormat).To(Equal("CycloneDX"))
			Expect(derivedBom.SpecVersion).To(Equal(cdx.SpecVersion1_6))
			Expect(derivedBom.Components).NotTo(BeNil())

			derivedComponents := *derivedBom.Components

			By("verifying curl component inherited from base with strict validation")
			expectedCurl := expectedComponent{
				Name:     "curl",
				Type:     cdx.ComponentTypeApplication,
				Version:  "8.12.1",
				PURL:     "pkg:generic/curl@8.12.1",
				Licenses: []string{"Apache-2.0"},
				Hashes: map[cdx.HashAlgorithm]string{
					cdx.HashAlgoSHA256: "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
				},
			}
			derivedCurlComponent := findComponentByName(derivedComponents, "curl")
			assertComponentEquals(derivedCurlComponent, expectedCurl)

			By("verifying derived-component from fragment with strict validation")
			expectedDerived := expectedComponent{
				Name:    "derived-component",
				Type:    cdx.ComponentTypeApplication,
				Version: "1.0.0",
				PURL:    "pkg:generic/derived-component@1.0.0",
			}
			derivedOwnComponent := findComponentByName(derivedComponents, "derived-component")
			assertComponentEquals(derivedOwnComponent, expectedDerived)

			By("verifying total components count")
			Expect(len(derivedComponents)).To(Equal(4),
				"derived SBOM should contain exactly 4 components (curl + libcurl from base + derived-component + derived-dep from fragment)")

			By("verifying derived dependencies from both base and fragment")
			Expect(derivedBom.Dependencies).NotTo(BeNil(), "derived SBOM should have dependencies")
			derivedDeps := *derivedBom.Dependencies

			derivedCurlDep := findDependencyByRef(derivedDeps, "curl")
			assertDependencyEquals(derivedCurlDep, expectedDependency{
				Ref:       "curl",
				DependsOn: []string{"libcurl"},
			})

			derivedOwnDep := findDependencyByRef(derivedDeps, "derived-component")
			assertDependencyEquals(derivedOwnDep, expectedDependency{
				Ref:       "derived-component",
				DependsOn: []string{"derived-dep"},
			})

			Expect(len(derivedDeps)).To(Equal(2),
				"derived SBOM should have exactly 2 dependencies (curl from base + derived-component from fragment)")
		},
		Entry("with local repo using Vanilla Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "vanilla-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("with local repo using BuildKit Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "buildkit-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		XEntry("with local repo using Native Buildah with chroot isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-chroot",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		XEntry("with local repo using Native Buildah with rootless isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-rootless",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
	)
})

func extractBOMFromSbomImage(ctx SpecContext, contRuntime contback.ContainerBackend, dockerImageName string) *cdx.BOM {
	sbomImageName := sbomImage.ImageName(dockerImageName)

	opener := func() (io.ReadCloser, error) {
		return contRuntime.SaveImageToStream(ctx, sbomImageName), nil
	}

	artifactContent, err := extract.FromImageBytes(opener)
	Expect(err).NotTo(HaveOccurred(), "failed to find SBOM artifact")

	bom, err := cyclonedxutil.BuildCycloneDX16BOMFromJSON(artifactContent)
	Expect(err).NotTo(HaveOccurred(), "failed to parse SBOM artifact")

	return bom
}

func findComponentByName(components []cdx.Component, name string) *cdx.Component {
	for i := range components {
		if components[i].Name == name {
			return &components[i]
		}
	}
	return nil
}

func findServiceByName(services []cdx.Service, name string) *cdx.Service {
	for i := range services {
		if services[i].Name == name {
			return &services[i]
		}
	}
	return nil
}

func findPropertyByName(properties []cdx.Property, name string) *cdx.Property {
	for i := range properties {
		if properties[i].Name == name {
			return &properties[i]
		}
	}
	return nil
}

func findExternalReferenceByURL(refs []cdx.ExternalReference, url string) *cdx.ExternalReference {
	for i := range refs {
		if refs[i].URL == url {
			return &refs[i]
		}
	}
	return nil
}

func findAnnotationByText(annotations []cdx.Annotation, text string) *cdx.Annotation {
	for i := range annotations {
		if annotations[i].Text == text {
			return &annotations[i]
		}
	}
	return nil
}

type expectedComponent struct {
	Name     string
	Type     cdx.ComponentType
	Version  string
	PURL     string
	Licenses []string
	Hashes   map[cdx.HashAlgorithm]string
}

type expectedService struct {
	Name          string
	Endpoints     []string
	Authenticated *bool
}

type expectedProperty struct {
	Name  string
	Value string
}

type expectedExternalReference struct {
	Type cdx.ExternalReferenceType
	URL  string
}

type expectedDependency struct {
	Ref       string
	DependsOn []string
}

func assertComponentEquals(actual *cdx.Component, expected expectedComponent) {
	Expect(actual).NotTo(BeNil(), "component %q should exist", expected.Name)
	Expect(actual.Name).To(Equal(expected.Name), "component name mismatch")
	Expect(actual.Type).To(Equal(expected.Type), "component %q type mismatch", expected.Name)
	Expect(actual.Version).To(Equal(expected.Version), "component %q version mismatch", expected.Name)
	Expect(actual.PackageURL).To(Equal(expected.PURL), "component %q PURL mismatch", expected.Name)

	if len(expected.Licenses) > 0 {
		Expect(actual.Licenses).NotTo(BeNil(), "component %q should have licenses", expected.Name)
		actualLicenseIDs := extractLicenseIDs(*actual.Licenses)
		Expect(actualLicenseIDs).To(ConsistOf(expected.Licenses), "component %q licenses mismatch", expected.Name)
	}

	if len(expected.Hashes) > 0 {
		Expect(actual.Hashes).NotTo(BeNil(), "component %q should have hashes", expected.Name)
		for alg, expectedHash := range expected.Hashes {
			foundHash := findHashByAlgorithm(*actual.Hashes, alg)
			Expect(foundHash).NotTo(BeNil(), "component %q should have hash with algorithm %s", expected.Name, alg)
			Expect(foundHash.Value).To(Equal(expectedHash), "component %q hash value mismatch for algorithm %s", expected.Name, alg)
		}
	}
}

func assertServiceEquals(actual *cdx.Service, expected expectedService) {
	Expect(actual).NotTo(BeNil(), "service %q should exist", expected.Name)
	Expect(actual.Name).To(Equal(expected.Name), "service name mismatch")

	if len(expected.Endpoints) > 0 {
		Expect(actual.Endpoints).NotTo(BeNil(), "service %q should have endpoints", expected.Name)
		Expect(*actual.Endpoints).To(ConsistOf(expected.Endpoints), "service %q endpoints mismatch", expected.Name)
	}

	if expected.Authenticated != nil {
		Expect(actual.Authenticated).NotTo(BeNil(), "service %q should have authenticated field", expected.Name)
		Expect(*actual.Authenticated).To(Equal(*expected.Authenticated), "service %q authenticated mismatch", expected.Name)
	}
}

func assertPropertyEquals(actual *cdx.Property, expected expectedProperty) {
	Expect(actual).NotTo(BeNil(), "property %q should exist", expected.Name)
	Expect(actual.Name).To(Equal(expected.Name), "property name mismatch")
	Expect(actual.Value).To(Equal(expected.Value), "property %q value mismatch", expected.Name)
}

func assertExternalReferenceEquals(actual *cdx.ExternalReference, expected expectedExternalReference) {
	Expect(actual).NotTo(BeNil(), "external reference with URL %q should exist", expected.URL)
	Expect(actual.Type).To(Equal(expected.Type), "external reference type mismatch for URL %q", expected.URL)
	Expect(actual.URL).To(Equal(expected.URL), "external reference URL mismatch")
}

func extractLicenseIDs(licenses cdx.Licenses) []string {
	var ids []string
	for _, lic := range licenses {
		if lic.License != nil && lic.License.ID != "" {
			ids = append(ids, lic.License.ID)
		}
	}
	return ids
}

func findHashByAlgorithm(hashes []cdx.Hash, alg cdx.HashAlgorithm) *cdx.Hash {
	for i := range hashes {
		if hashes[i].Algorithm == alg {
			return &hashes[i]
		}
	}
	return nil
}

func findDependencyByRef(deps []cdx.Dependency, ref string) *cdx.Dependency {
	for i := range deps {
		if deps[i].Ref == ref {
			return &deps[i]
		}
	}
	return nil
}

func assertDependencyEquals(actual *cdx.Dependency, expected expectedDependency) {
	Expect(actual).NotTo(BeNil(), "dependency with ref %q should exist", expected.Ref)
	Expect(actual.Ref).To(Equal(expected.Ref), "dependency ref mismatch")

	if len(expected.DependsOn) > 0 {
		Expect(actual.Dependencies).NotTo(BeNil(), "dependency %q should have dependsOn", expected.Ref)
		Expect(*actual.Dependencies).To(ConsistOf(expected.DependsOn), "dependency %q dependsOn mismatch", expected.Ref)
	} else if actual.Dependencies != nil {
		Expect(*actual.Dependencies).To(BeEmpty(), "dependency %q should have no dependsOn", expected.Ref)
	}
}

func boolPtr(b bool) *bool {
	return &b
}
