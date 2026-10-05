package build

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/storage/manager"
)

var _ = ginkgo.Describe("Check mode import metadata", func() {
	ginkgo.DescribeTable("reads secondary metadata without copying only in check mode", func(ctx ginkgo.SpecContext, checkMode bool, expectedWrites int) {
		metadata := &storage.ImportMetadata{ImportSourceID: "import", Checksum: "checksum"}
		primary := &importMetadataStorageStub{err: storage.ErrImportMetadataNotFound}
		secondary := &importMetadataStorageStub{metadata: metadata}
		c := &Conveyor{StorageManager: &manager.StorageManager{
			StagesStorage:              primary,
			SecondaryStagesStorageList: []storage.StagesStorage{secondary},
		}}
		NewBuildPhase(c, BuildPhaseOptions{ShouldBeBuiltMode: checkMode})

		actual, err := c.FetchImportMetadata(ctx, "project", "import")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(actual).To(gomega.Equal(metadata))
		gomega.Expect(primary.writes).To(gomega.Equal(expectedWrites))
		gomega.Expect(secondary.writes).To(gomega.BeZero())
		gomega.Expect(c.IsImagesReadOnly(ctx)).To(gomega.Equal(checkMode))
	},
		ginkgo.Entry("check mode", true, 0),
		ginkgo.Entry("ordinary build", false, 1),
	)

	ginkgo.It("restores ordinary import publication for the next build phase", func(ctx ginkgo.SpecContext) {
		primary := &importMetadataStorageStub{err: storage.ErrImportMetadataNotFound}
		secondary := &importMetadataStorageStub{metadata: &storage.ImportMetadata{Checksum: "checksum"}}
		c := &Conveyor{StorageManager: &manager.StorageManager{
			StagesStorage:              primary,
			SecondaryStagesStorageList: []storage.StagesStorage{secondary},
		}}
		NewBuildPhase(c, BuildPhaseOptions{ShouldBeBuiltMode: true})
		NewBuildPhase(c, BuildPhaseOptions{})

		_, err := c.FetchImportMetadata(ctx, "project", "import")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(primary.writes).To(gomega.Equal(1))
		gomega.Expect(c.IsImagesReadOnly(ctx)).To(gomega.BeFalse())
	})
})
