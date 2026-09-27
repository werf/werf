package image

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/config"
	"github.com/werf/werf/v3/test/pkg/utils"
)

var _ = ginkgo.Describe("Local Git preparation", func() {
	ginkgo.It("fetches a shallow origin once per graph calculation across images and platforms", func(ctx ginkgo.SpecContext) {
		data, err := os.ReadFile("testdata/preparation/werf.yaml")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		origin := newProjectRepo(ctx, map[string]string{"werf.yaml": string(data)})
		commitFiles(ctx, origin, map[string]string{"data": "second"})
		shallowOrigin := filepath.Join(ginkgo.GinkgoT().TempDir(), "origin")
		utils.RunSucceedCommand(ctx, origin, "git", "clone", "--depth=1", "file://"+origin, shallowOrigin)
		project := filepath.Join(ginkgo.GinkgoT().TempDir(), "project")
		utils.RunSucceedCommand(ctx, origin, "git", "clone", "--depth=1", "file://"+shallowOrigin, project)
		_, _, opts := stapelImageConfig(ctx, project)
		_, cfg, err := config.GetWerfConfig(ctx, "", "", "", opts.GiterminismManager, config.WerfConfigOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		opts.Conveyor = &preparationTestConveyor{}
		images, err := config.NewImagesToProcess(cfg, []string{}, false, false)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		tree := NewImagesTree(cfg, ImagesTreeOptions{CommonImageOptions: opts, ImagesToProcess: images})
		trace := filepath.Join(ginkgo.GinkgoT().TempDir(), "trace")
		ginkgo.GinkgoT().Setenv("GIT_TRACE", trace)
		for attempt := 1; attempt <= 2; attempt++ {
			gomega.Expect(tree.Calculate(ctx)).To(gomega.Succeed())
			gomega.Expect(tree.GetImages()).To(gomega.HaveLen(4))
			shallow, err := opts.GiterminismManager.LocalGitRepo().IsShallowClone(ctx)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(shallow).To(gomega.BeTrue())
			data, err := os.ReadFile(trace)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(strings.Count(string(data), "--unshallow")).To(gomega.Equal(attempt))
		}
		utils.RunSucceedCommand(ctx, project, "git", "remote", "set-url", "origin", filepath.Join(project, "missing-origin"))
		gomega.Expect(tree.Calculate(ctx)).To(gomega.MatchError(gomega.ContainSubstring("unable to fetch local git repo")))
		utils.RunSucceedCommand(ctx, project, "git", "remote", "set-url", "origin", "file://"+shallowOrigin)
		gomega.Expect(tree.Calculate(ctx)).To(gomega.Succeed())
	})
})
