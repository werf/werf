package stage

import (
	"bytes"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/storage"
)

var _ = ginkgo.Describe("Check mode import checksum", func() {
	ginkgo.DescribeTable("rejects unavailable metadata without fetching or regenerating", func(ctx ginkgo.SpecContext, metadata *storage.ImportMetadata, metadataErr error) {
		c := &importChecksumConveyorStub{readOnly: true, metadata: metadata, err: metadataErr}
		s := newDependenciesStage(nil, nil, DependenciesAfterSetup, &BaseStageOptions{ProjectName: "project"})
		var output bytes.Buffer
		logCtx := logboek.NewContext(ctx, logboek.NewLogger(&output, &output))

		_, err := s.getImportSourceChecksum(logCtx, c, NewContainerBackendStub(), &config.Import{ArtifactExport: &config.ArtifactExport{ExportBase: &config.ExportBase{Add: "/app"}}, ImageName: "source"})
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("run a regular build to regenerate it")))
		gomega.Expect(c.fetches).To(gomega.BeZero())
		gomega.Expect(output.String()).NotTo(gomega.ContainSubstring("will regenerate"))
	},
		ginkgo.Entry("missing metadata", nil, storage.ErrImportMetadataNotFound),
		ginkgo.Entry("broken metadata", nil, storage.ErrBrokenImage),
		ginkgo.Entry("empty checksum", &storage.ImportMetadata{}, nil),
	)

	ginkgo.It("uses an existing checksum without fetching its source", func(ctx ginkgo.SpecContext) {
		c := &importChecksumConveyorStub{readOnly: true, metadata: &storage.ImportMetadata{Checksum: "checksum"}}
		s := newDependenciesStage(nil, nil, DependenciesAfterSetup, &BaseStageOptions{ProjectName: "project"})
		checksum, err := s.getImportSourceChecksum(ctx, c, NewContainerBackendStub(), &config.Import{ArtifactExport: &config.ArtifactExport{ExportBase: &config.ExportBase{Add: "/app"}}, ImageName: "source"})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(checksum).To(gomega.Equal("checksum"))
		gomega.Expect(c.fetches).To(gomega.BeZero())
	})

	ginkgo.It("still attempts checksum generation during an ordinary build", func(ctx ginkgo.SpecContext) {
		c := &importChecksumConveyorStub{err: storage.ErrImportMetadataNotFound}
		s := newDependenciesStage(nil, nil, DependenciesAfterSetup, &BaseStageOptions{ProjectName: "project"})
		_, err := s.getImportSourceChecksum(ctx, c, NewContainerBackendStub(), &config.Import{ArtifactExport: &config.ArtifactExport{ExportBase: &config.ExportBase{Add: "/app"}}, ImageName: "source"})
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("checksum source fetch requested")))
		gomega.Expect(c.fetches).To(gomega.Equal(1))
	})
})
