package image

import (
	"bytes"
	"net/http"
	"os"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/level"
	"github.com/werf/werf/v3/pkg/config"
)

var _ = ginkgo.DescribeTable("Git mapping service output", func(ctx ginkgo.SpecContext, remote bool, expectedCommits int) {
	var tree *ImagesTree
	if remote {
		tree = newRemotePreparationTree(ctx, func(backend http.Handler) http.Handler { return backend })
		for _, img := range tree.werfConfig.GetImagesForProcessing(tree.ImagesToProcess) {
			for _, mapping := range img.(config.StapelImageInterface).ImageBaseConfig().Git.Remote {
				mapping.Name = "shared"
			}
		}
		app := tree.werfConfig.GetImage("app").(config.StapelImageInterface).ImageBaseConfig()
		otherRef := *app.Git.Remote[0]
		otherExport := *otherRef.GitRemoteExport
		otherRef.GitRemoteExport = &otherExport
		otherRef.Commit = ""
		otherRef.Branch = "main"
		otherRef.RepoCacheKey += "-branch"
		app.Git.Remote = append(app.Git.Remote, &otherRef)
	} else {
		data, err := os.ReadFile("testdata/preparation/werf.yaml")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		project := newProjectRepo(ctx, map[string]string{"werf.yaml": string(data)})
		_, _, opts := stapelImageConfig(ctx, project)
		_, cfg, err := config.GetWerfConfig(ctx, "", "", "", opts.GiterminismManager, config.WerfConfigOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		opts.Conveyor = &preparationTestConveyor{}
		images, err := config.NewImagesToProcess(cfg, []string{}, false, false)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		tree = NewImagesTree(cfg, ImagesTreeOptions{CommonImageOptions: opts, ImagesToProcess: images})
	}
	var output bytes.Buffer
	logger := logboek.NewLogger(&output, &output)
	logger.SetAcceptedLevel(level.Debug)
	logger.Streams().DisableStyle()
	logger.Streams().DisableLineWrapping()
	runCtx := logboek.NewContext(ctx, logger)
	for range 2 {
		output.Reset()
		gomega.Expect(tree.Calculate(runCtx)).To(gomega.Succeed())
		gomega.Expect(tree.GetImages()).To(gomega.HaveLen(4))
		gomega.Expect(strings.Count(output.String(), " will be used for ")).To(gomega.Equal(expectedCommits))
		gomega.Expect(output.String()).To(gomega.ContainSubstring("git mapping from"))
		gomega.Expect(output.String()).NotTo(gomega.ContainSubstring("Initializing git mappings"))
	}
}, ginkgo.Entry("local commit once per calculation", false, 1),
	ginkgo.Entry("distinct repositories and distinct commits of one repository", true, 4))
