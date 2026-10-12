package stage_manager

import (
	"context"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/storage/manager"
	"github.com/werf/werf/v2/pkg/werf"
)

func TestMetadataRouting(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Stage manager metadata routing")
}

var _ = ginkgo.Describe("Separate cleanup metadata storage", func() {
	ginkgo.It("uses meta image history instead of stale stages-repo history", func(ctx ginkgo.SpecContext) {
		retained := newTestStageDesc("retained", 1)
		stale := newTestStageDesc("stale", 2)
		expected := map[string]map[string][]string{"app": {retained.StageID.String(): {"retained-commit"}}}
		stages := &metadataRoutingStorage{images: map[string]map[string][]string{"old-app": {stale.StageID.String(): {"stale-commit"}}}}
		meta := &metadataRoutingStorage{images: expected}
		sm := &manager.StorageManager{StagesStorage: stages, MetaStorage: meta}
		m := NewManager()
		m.managedStageDescSet = newManagedStageDescSet(image.NewStageDescSet(retained, stale))
		gomega.Expect(m.InitImagesMetadata(ctx, sm, metadataRoutingGitRepo{}, "project", []string{"app", "old-app"})).To(gomega.Succeed())
		gomega.Expect(m.GetImageStageIDCommitListToCleanup()).To(gomega.Equal(expected))
	})

	ginkgo.It("reads custom-tag IDs and their stage mapping from meta storage", func() {
		ctx := context.Background()
		gomega.Expect(werf.Init(ginkgo.GinkgoT().TempDir(), ginkgo.GinkgoT().TempDir())).To(gomega.Succeed())
		stages := &metadataRoutingStorage{customTags: map[string]*storage.CustomTagMetadata{"old": {StageID: "stale-2", Tag: "old"}}}
		meta := &metadataRoutingStorage{customTags: map[string]*storage.CustomTagMetadata{"release": {StageID: "retained-1", Tag: "release"}}}
		sm := &manager.StorageManager{StagesStorage: stages, MetaStorage: meta}
		m := NewManager()
		gomega.Expect(m.InitCustomTagsMetadata(ctx, sm)).To(gomega.Succeed())
		gomega.Expect(m.GetCustomTagsMetadata()).To(gomega.Equal(map[string][]string{"retained-1": {"release"}}))
	})
})
