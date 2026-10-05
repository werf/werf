package stages

import (
	"errors"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/ref"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/storage/manager"
)

func TestStages(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Stages Suite")
}

var _ = ginkgo.Describe("Copied stage import metadata", func() {
	ginkgo.DescribeTable("selects only metadata matching copied import checksums", func(ctx ginkgo.SpecContext, index bool) {
		info := &image.Info{Labels: map[string]string{
			image.WerfImportChecksumLabelPrefix + "import":      "used",
			image.WerfImportSourceStageIDLabelPrefix + "import": "old-source",
			"unrelated": "unused",
		}}
		if index {
			info = &image.Info{Index: []*image.Info{info}}
		}
		source := &RemoteStorage{
			RegistryAddress: &ref.RegistryAddress{Reference: &ref.Reference{Repo: "source"}},
			RegistryClient:  &importMetadataRegistryStub{info: info},
			StorageManager: &manager.StorageManager{StagesStorage: &importMetadataStorageStub{
				ids: []string{"old", "new", "unrelated"},
				metadata: map[string]*storage.ImportMetadata{
					"old":       {Checksum: "used", SourceStageID: "old-source"},
					"new":       {Checksum: "used", SourceStageID: "new-source"},
					"unrelated": {Checksum: "unused", SourceStageID: "old-source"},
				},
			}},
		}
		refs, err := source.getImportMetadataImageRefs(ctx, "project", []string{"source:stage", "source:same-import"})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(refs).To(gomega.ConsistOf("source:import-metadata-old", "source:import-metadata-new"))
	},
		ginkgo.Entry("image", false),
		ginkgo.Entry("image index", true),
	)
	ginkgo.It("does not list metadata when copied stages have no import checksums", func(ctx ginkgo.SpecContext) {
		metadataStorage := &importMetadataStorageStub{}
		source := &RemoteStorage{RegistryClient: &importMetadataRegistryStub{info: &image.Info{}}, StorageManager: &manager.StorageManager{StagesStorage: metadataStorage}}
		refs, err := source.getImportMetadataImageRefs(ctx, "project", []string{"source:stage"})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(refs).To(gomega.BeEmpty())
		gomega.Expect(metadataStorage.listCalls).To(gomega.BeZero())
	})
	ginkgo.DescribeTable("handles unavailable metadata", func(ctx ginkgo.SpecContext, metadataErr error, expectedError bool) {
		source := &RemoteStorage{
			RegistryAddress: &ref.RegistryAddress{Reference: &ref.Reference{Repo: "source"}},
			RegistryClient:  &importMetadataRegistryStub{info: &image.Info{Labels: map[string]string{image.WerfImportChecksumLabelPrefix + "import": "used"}}},
			StorageManager: &manager.StorageManager{StagesStorage: &importMetadataStorageStub{
				ids:      []string{"stale", "valid"},
				errors:   map[string]error{"stale": metadataErr},
				metadata: map[string]*storage.ImportMetadata{"valid": {Checksum: "used"}},
			}},
		}
		refs, err := source.getImportMetadataImageRefs(ctx, "project", []string{"source:stage"})
		if expectedError {
			gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("access denied")))
			return
		}
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(refs).To(gomega.Equal([]string{"source:import-metadata-valid"}))
	},
		ginkgo.Entry("deleted metadata", storage.ErrImportMetadataNotFound, false),
		ginkgo.Entry("broken metadata", storage.ErrBrokenImage, false),
		ginkgo.Entry("registry error", errors.New("access denied"), true),
	)
})
