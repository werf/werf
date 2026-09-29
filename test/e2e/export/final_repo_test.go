package e2e_export_test

import (
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/image"
	"github.com/werf/werf/v3/test/pkg/contback"
	"github.com/werf/werf/v3/test/pkg/suite_init"
	"github.com/werf/werf/v3/test/pkg/utils"
	"github.com/werf/werf/v3/test/pkg/utils/docker"
	"github.com/werf/werf/v3/test/pkg/werf"
)

var _ = ginkgo.Describe("Export with a final repo", ginkgo.Label("e2e", "export", "final-repo"), func() {
	ginkgo.DescribeTable("exports the same image before and after content-anchor reuse",
		func(ctx ginkgo.SpecContext, backendMode string, local bool) {
			contback.SkipIfUnavailable(backendMode)
			setupEnv()
			SuiteData.Stubs.SetEnv("WERF_BUILDAH_MODE", backendMode)
			if local {
				SuiteData.Stubs.SetEnv("WERF_REPO", ":local")
			}

			finalRepo := suite_init.TestRepo(fmt.Sprintf("werf-export-final-%s", utils.GetRandomString(10)))
			SuiteData.Stubs.SetEnv("WERF_FINAL_REPO", finalRepo)
			SuiteData.InitTestRepo(ctx, "repo", "simple")
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath("repo"))
			exportRepo := suite_init.TestRepo(fmt.Sprintf("werf-export-%s", utils.GetRandomString(10)))

			coldOut := werfProject.Export(ctx, &werf.ExportOptions{
				CommonOptions: werf.CommonOptions{ExtraArgs: getExportArgs(exportRepo+":cold", commonTestOptions{})},
			})
			gomega.Expect(coldOut).To(gomega.ContainSubstring("Copy stage"))
			gomega.Expect(coldOut).To(gomega.ContainSubstring("Exporting image"))

			repository, err := name.NewRepository(finalRepo)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			tags, err := remote.List(repository, remote.WithContext(ctx))
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(tags).To(gomega.HaveLen(1))
			finalRef := repository.Tag(tags[0])
			_, err = remote.Head(finalRef, remote.WithContext(ctx))
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			if local && backendMode == "docker" {
				ginkgo.By("removing only the local final-repo alias, leaving the primary image cached")
				gomega.Expect(docker.CliRmi(ctx, finalRef.Name())).To(gomega.Succeed())
			}

			warmOut := werfProject.Export(ctx, &werf.ExportOptions{
				CommonOptions: werf.CommonOptions{ExtraArgs: getExportArgs(exportRepo+":warm", commonTestOptions{})},
			})
			gomega.Expect(warmOut).To(gomega.ContainSubstring("by content-based tag"))
			gomega.Expect(warmOut).To(gomega.ContainSubstring("Use previously built final image"))
			gomega.Expect(warmOut).To(gomega.ContainSubstring("Exporting image"))

			var coldDigest string
			for _, tag := range []string{"cold", "warm"} {
				reference, err := name.ParseReference(exportRepo + ":" + tag)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				exported, err := remote.Image(reference, remote.WithContext(ctx))
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				digest, err := exported.Digest()
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				if tag == "cold" {
					coldDigest = digest.String()
				} else {
					gomega.Expect(digest.String()).To(gomega.Equal(coldDigest))
				}
				config, err := exported.ConfigFile()
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				for label := range config.Config.Labels {
					gomega.Expect(strings.HasPrefix(label, image.WerfLabelPrefix)).To(gomega.BeFalse(), "service label %s must be stripped", label)
				}
			}
		},
		ginkgo.Entry("remote primary with Docker", "docker", false),
		ginkgo.Entry("local primary with Docker", "docker", true),
		ginkgo.Entry("local primary with Native Buildah", ginkgo.Label(suite_init.LabelNeedsBuildah), "native-rootless", true),
	)
})
