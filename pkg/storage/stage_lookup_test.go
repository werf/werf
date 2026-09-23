package storage

import (
	"errors"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/image"
)

var _ = ginkgo.Describe("stage lookup", func() {
	ginkgo.DescribeTable("returns an unavailable error instead of a nil descriptor",
		func(ctx ginkgo.SpecContext, present, rejected bool, expected error) {
			registry := &stageLookupRegistry{markerRegistry: newMarkerRegistry()}
			storage := &RepoStagesStorage{RepoAddress: "registry.example/project", DockerRegistry: registry}
			stageID := image.NewStageID("digest", 1)
			if present {
				registry.put(storage.ConstructStageImageName("project", stageID.Digest, stageID.CreationTs), nil)
			}
			if rejected {
				registry.put(makeRepoRejectedStageImageRecord(storage.RepoAddress, stageID.Digest, stageID.CreationTs), nil)
			}

			desc, err := storage.GetStageDesc(ctx, "project", *stageID)
			if expected != nil {
				gomega.Expect(err).To(gomega.MatchError(expected))
				gomega.Expect(IsErrStageUnavailable(err)).To(gomega.BeTrue())
				gomega.Expect(desc).To(gomega.BeNil())
			} else {
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(desc).NotTo(gomega.BeNil())
				gomega.Expect(desc.StageID).To(gomega.Equal(stageID))
			}
		},
		ginkgo.Entry("missing", false, false, ErrStageNotFound),
		ginkgo.Entry("rejected", true, true, ErrStageRejected),
		ginkgo.Entry("available", true, false, nil),
	)

	ginkgo.It("returns an unavailable error for a missing local image", func(ctx ginkgo.SpecContext) {
		storage := NewLocalStagesStorage(&stageLookupBackend{})
		desc, err := storage.GetStageDesc(ctx, "project", *image.NewStageID("digest", 1))
		gomega.Expect(err).To(gomega.MatchError(ErrStageNotFound))
		gomega.Expect(desc).To(gomega.BeNil())
	})

	ginkgo.DescribeTable("copies a missing destination without hiding other failures",
		func(ctx ginkgo.SpecContext, existing, rejected bool, copyErr, expected error) {
			registry := &stageLookupRegistry{markerRegistry: newMarkerRegistry()}
			registry.copyErr = copyErr
			source := &RepoStagesStorage{RepoAddress: "registry.example/source", DockerRegistry: registry}
			destination := &RepoStagesStorage{RepoAddress: "registry.example/destination", DockerRegistry: registry}
			stageID := image.NewStageID("digest", 1)
			sourceRef := source.ConstructStageImageName("project", stageID.Digest, stageID.CreationTs)
			destinationRef := destination.ConstructStageImageName("project", stageID.Digest, stageID.CreationTs)
			registry.put(sourceRef, nil)
			if existing {
				registry.put(destinationRef, nil)
			}
			if rejected {
				registry.put(makeRepoRejectedStageImageRecord(destination.RepoAddress, stageID.Digest, stageID.CreationTs), nil)
			}

			desc, err := destination.CopyFromStorage(ctx, source, "project", *stageID, CopyFromStorageOptions{})
			if expected != nil {
				gomega.Expect(errors.Is(err, expected)).To(gomega.BeTrue(), "error: %v", err)
				gomega.Expect(desc).To(gomega.BeNil())
			} else {
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(desc).NotTo(gomega.BeNil())
				gomega.Expect(desc.Info.Name).To(gomega.Equal(destinationRef))
			}
		},
		ginkgo.Entry("copy on miss", false, false, nil, nil),
		ginkgo.Entry("reuse an existing image without copying", true, false, ErrBrokenImage, nil),
		ginkgo.Entry("preserve a rejection", true, true, nil, ErrStageRejected),
		ginkgo.Entry("propagate a copy error", false, false, ErrBrokenImage, ErrBrokenImage),
	)
})
