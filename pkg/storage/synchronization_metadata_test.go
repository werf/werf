package storage

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Synchronization metadata isolation", func() {
	ginkgo.It("retains synchronization records outside movable project metadata", func(ctx ginkgo.SpecContext) {
		registry := newMarkerRegistry()
		primary := newRepoStorage(registry, stagesRepo)
		client := &ClientIDRecord{ClientID: "shared-client", TimestampMillisec: 10}
		server := &SyncServerRecord{Server: "https://sync.example.test", TimestampMillisec: 20}
		gomega.Expect(primary.PostClientIDRecord(ctx, proj, client)).To(gomega.Succeed())
		gomega.Expect(primary.PostSyncServerRecord(ctx, proj, server)).To(gomega.Succeed())
		gomega.Expect(primary.PutImageMetadata(ctx, proj, "app", "commit", "stage")).To(gomega.Succeed())

		records, err := primary.collectMetadataRecords(ctx)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(records).To(gomega.HaveLen(1))
		gomega.Expect(records[0].tag).To(gomega.HavePrefix(RepoImageMetadataByCommitRecord_ImageTagPrefix))
		clients, err := primary.GetClientIDRecords(ctx, proj)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(clients).To(gomega.Equal([]*ClientIDRecord{client}))
		servers, err := primary.GetSyncServerRecords(ctx, proj)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(servers).To(gomega.Equal([]*SyncServerRecord{server}))
	})
})
