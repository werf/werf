package e2e_build_test

import (
	"fmt"
	"sort"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/container_backend/thirdparty/platformutil"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/test/pkg/contback"
	"github.com/werf/werf/test/pkg/werf"
)

type multiarchTestOptions struct {
	setupEnvOptions

	Platforms                   []string
	EnableStapelImage           bool
	EnableStagedDockerfileImage bool
	EnableDockerfileImage       bool

	ExpectedStapelImageInfo           expectedImageInfo
	ExpectedStagedDockerfileImageInfo expectedImageInfo
	ExpectedDockerfileImageInfo       expectedImageInfo
}

type expectedImageInfo struct {
	ImageName        string
	DigestByPlatform map[string]string
}

var _ = Describe("Multiarch build", Label("e2e", "build", "multiarch", "simple"), func() {
	DescribeTable("should build images for multiple architectures and publish multiarch manifests",
		func(testOpts multiarchTestOptions) {
			setupEnv(testOpts.setupEnvOptions)
			Expect(SuiteData.WerfRepo).NotTo(BeEmpty())

			contBack, err := contback.NewContainerBackend(testOpts.ContainerBackendMode)
			if err == contback.ErrRuntimeUnavailable {
				Skip(err.Error())
			} else if err != nil {
				Fail(err.Error())
			}

			repoDirname := "repo0"
			fixtureRelPath := "multiarch/state0"
			buildReportName := "report0.json"

			SuiteData.InitTestRepo(repoDirname, fixtureRelPath)

			var expects []*expectedImageInfo

			if testOpts.EnableStapelImage {
				SuiteData.Stubs.SetEnv("ENABLE_STAPEL_IMAGE", "1")
				expects = append(expects, &testOpts.ExpectedStapelImageInfo)
			} else {
				SuiteData.Stubs.SetEnv("ENABLE_STAPEL_IMAGE", "0")
			}

			if testOpts.EnableStagedDockerfileImage {
				SuiteData.Stubs.SetEnv("ENABLE_STAGED_DOCKERFILE_IMAGE", "1")
				expects = append(expects, &testOpts.ExpectedStagedDockerfileImageInfo)
			} else {
				SuiteData.Stubs.SetEnv("ENABLE_STAGED_DOCKERFILE_IMAGE", "0")
			}

			if testOpts.EnableDockerfileImage {
				SuiteData.Stubs.SetEnv("ENABLE_DOCKERFILE_IMAGE", "1")
				expects = append(expects, &testOpts.ExpectedDockerfileImageInfo)
			} else {
				SuiteData.Stubs.SetEnv("ENABLE_DOCKERFILE_IMAGE", "0")
			}

			if testOpts.Platforms != nil {
				SuiteData.Stubs.SetEnv("WERF_PLATFORM", strings.Join(testOpts.Platforms, ","))
			} else {
				SuiteData.Stubs.SetEnv("WERF_PLATFORM", "")
			}

			SuiteData.Stubs.SetEnv("WERF_ENABLE_REPORT_BY_PLATFORM", "1")

			By("building images")
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
			_, buildReport := werfProject.BuildWithReport(SuiteData.GetBuildReportPath(buildReportName), nil)

			By("check digests by platform and multiarch digest")
			for _, expect := range expects {
				byPlatform := buildReport.ImagesByPlatform[expect.ImageName]
				Expect(len(byPlatform)).To(Equal(len(testOpts.Platforms)))

				for _, platform := range testOpts.Platforms {
					Expect(strings.HasPrefix(byPlatform[platform].DockerTag, expect.DigestByPlatform[platform]+"-")).To(BeTrue())

					platformSpec, err := platformutil.ParsePlatform(platform)
					Expect(err).To(Succeed())

					ref := fmt.Sprintf("%s:%s", SuiteData.WerfRepo, byPlatform[platform].DockerTag)
					inspect := contBack.GetImageInspect(ref)

					fmt.Printf("Check image %q inspect:\n%#v\n---\n", ref, inspect)

					Expect(inspect.Os).To(Equal(platformSpec.OS))
					Expect(inspect.Architecture).To(Equal(platformSpec.Architecture))
					Expect(inspect.Variant).To(Equal(platformSpec.Variant))
				}

				// Meta digest only used for multiplatform builds
				if len(expect.DigestByPlatform) > 1 {
					platforms := util.MapKeys(expect.DigestByPlatform)
					sort.Strings(platforms)

					metaDeps := util.MapFuncToSlice(platforms, func(platform string) string {
						return buildReport.ImagesByPlatform[expect.ImageName][platform].DockerTag
					})
					expectedMetaDigest := util.Sha3_224Hash(metaDeps...)

					fmt.Printf("metaDeps %v -> %q\n", metaDeps, expectedMetaDigest)
					Expect(buildReport.Images[expect.ImageName].DockerTag).To(Equal(expectedMetaDigest))
				} else {
					expectedDigest := util.MapValues(expect.DigestByPlatform)[0]
					Expect(strings.HasPrefix(buildReport.Images[expect.ImageName].DockerTag, expectedDigest+"-")).To(BeTrue())
				}
			}
		},

		Entry("Buildah backend, build arbitrary platforms, all builders available", multiarchTestOptions{
			setupEnvOptions: setupEnvOptions{
				WithLocalRepo:        true,
				ContainerBackendMode: "native-chroot",
			},
			Platforms:                   []string{"linux/arm64", "linux/amd64"},
			EnableStapelImage:           true,
			EnableStagedDockerfileImage: true,
			EnableDockerfileImage:       true,

			ExpectedStapelImageInfo: expectedImageInfo{
				ImageName: "orange",
				DigestByPlatform: map[string]string{
					"linux/arm64": "0bfe25871a0014046617e5c47e399183886b23ab394ee8422cc8b10c",
					"linux/amd64": "f401fc847eba504268377fe9a6e192bd90cd9cd4e9333c3564846264",
				},
			},
			ExpectedStagedDockerfileImageInfo: expectedImageInfo{
				ImageName: "apple",
				DigestByPlatform: map[string]string{
					"linux/arm64": "e09a53cbff65cd32668dd9749845923d46be8a70c1e1f12b1ae3318b",
					"linux/amd64": "8f89f4cdc76a994e38f4146ef4e3edd83e070266cc15c135696bddd2",
				},
			},
			ExpectedDockerfileImageInfo: expectedImageInfo{
				ImageName: "potato",
				DigestByPlatform: map[string]string{
					"linux/arm64": "cd8956d94612821a843cfbbb3b44d9ed837f8001f2a2361a4815acce",
					"linux/amd64": "30bd7a840fcb35eb283d9940f58d1dd02a576a6967e416d9c13deaf2",
				},
			},
		}),

		Entry("Docker backend, docker can build stapel image only for linux/amd64 platform", multiarchTestOptions{
			setupEnvOptions: setupEnvOptions{
				WithLocalRepo:        true,
				ContainerBackendMode: "vanilla-docker",
			},
			Platforms:         []string{"linux/amd64"},
			EnableStapelImage: true,

			ExpectedStapelImageInfo: expectedImageInfo{
				ImageName: "orange",
				DigestByPlatform: map[string]string{
					"linux/amd64": "f401fc847eba504268377fe9a6e192bd90cd9cd4e9333c3564846264",
				},
			},
		}),

		Entry("Docker backend, build arbitrary platforms, only non-staged dockerfile builder available", multiarchTestOptions{
			setupEnvOptions: setupEnvOptions{
				WithLocalRepo:        true,
				ContainerBackendMode: "vanilla-docker",
			},
			Platforms:             []string{"linux/arm64", "linux/amd64"},
			EnableDockerfileImage: true,

			ExpectedDockerfileImageInfo: expectedImageInfo{
				ImageName: "potato",
				DigestByPlatform: map[string]string{
					"linux/arm64": "cd8956d94612821a843cfbbb3b44d9ed837f8001f2a2361a4815acce",
					"linux/amd64": "30bd7a840fcb35eb283d9940f58d1dd02a576a6967e416d9c13deaf2",
				},
			},
		}),
	)
})
