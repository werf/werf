package host_cleaning

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("auto cleanup notice", func() {
	It("should not be saved when nothing was cleaned up and the storage limit is not exceeded", func(ctx context.Context) {
		dir := GinkgoT().TempDir()

		Expect(writeAutoCleanupNotice(dir, "docker", RunGCReport{
			UsedBytes:    50,
			TotalBytes:   100,
			AllowedBytes: 70,
		})).To(Succeed())

		message, needsAttention, err := PopAutoCleanupNotice(ctx, dir)
		Expect(err).To(Succeed())
		Expect(message).To(BeEmpty())
		Expect(needsAttention).To(BeFalse())
	})

	It("should be reported once when images were removed and the storage limit is not exceeded", func(ctx context.Context) {
		dir := GinkgoT().TempDir()

		Expect(writeAutoCleanupNotice(dir, "docker", RunGCReport{
			ImagesDeleted:  3,
			SpaceReclaimed: 30,
			UsedBytes:      50,
			TotalBytes:     100,
			AllowedBytes:   70,
			StoragePath:    "/var/lib/docker",
		})).To(Succeed())

		message, needsAttention, err := PopAutoCleanupNotice(ctx, dir)
		Expect(err).To(Succeed())
		Expect(needsAttention).To(BeFalse())
		Expect(message).To(ContainSubstring("3 local docker images removed"))
		Expect(message).To(ContainSubstring("/var/lib/docker"))
		Expect(message).NotTo(ContainSubstring("still above"))

		message, _, err = PopAutoCleanupNotice(ctx, dir)
		Expect(err).To(Succeed())
		Expect(message).To(BeEmpty())
	})

	It("should not claim whether images were removed when only the total is known", func(ctx context.Context) {
		dir := GinkgoT().TempDir()

		Expect(writeAutoCleanupNotice(dir, "docker", RunGCReport{
			SpaceReclaimed: 30,
			UsedBytes:      50,
			TotalBytes:     100,
			AllowedBytes:   70,
			StoragePath:    "/var/lib/docker",
		})).To(Succeed())

		message, _, err := PopAutoCleanupNotice(ctx, dir)
		Expect(err).To(Succeed())
		Expect(message).To(ContainSubstring("freed 30 B in total"))
		Expect(message).NotTo(ContainSubstring("images removed"))
		Expect(message).NotTo(ContainSubstring("might be rebuilt"))
	})

	It("should require attention when the storage limit is still exceeded after cleanup", func(ctx context.Context) {
		dir := GinkgoT().TempDir()

		Expect(writeAutoCleanupNotice(dir, "buildah", RunGCReport{
			ImagesDeleted:  1,
			SpaceReclaimed: 5,
			UsedBytes:      95,
			TotalBytes:     100,
			AllowedBytes:   70,
			StoragePath:    "/var/lib/containers",
		})).To(Succeed())

		message, needsAttention, err := PopAutoCleanupNotice(ctx, dir)
		Expect(err).To(Succeed())
		Expect(needsAttention).To(BeTrue())
		Expect(message).To(ContainSubstring("still above the allowed level after the cleanup"))
		Expect(message).To(ContainSubstring("--allowed-backend-storage-volume-usage"))
	})
})
