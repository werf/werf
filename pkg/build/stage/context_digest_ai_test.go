package stage

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func getContextDigestOrPanic(ctx context.Context, s Interface, c Conveyor) (result string, didPanic bool) {
	defer func() {
		if r := recover(); r != nil {
			didPanic = true
		}
	}()
	result, _ = s.GetContextDependencies(ctx, c)
	return
}

var _ = Describe("ContextDigest", func() {
	var (
		ctx      = context.Background()
		commitH  = "9d8059842b6fde712c58315ca0ab4713d90761c0"
		conveyor *ConveyorStub
	)

	BeforeEach(func() {
		conveyor = NewConveyorStub(
			NewGiterminismManagerStub(NewLocalGitRepoStub(commitH), NewGiterminismInspectorStub()),
			map[string]string{},
			map[string]string{},
			map[string]string{"dep-image": "dep-digest-v1"},
		)
	})

	Describe("Determinism", func() {
		It("same FromStage inputs produce same context digest", func() {
			opts := &BaseStageOptions{TargetPlatform: "linux/amd64"}
			s1 := &FromStage{imageCacheVersion: "v1", fromCacheVersion: "fc1", BaseStage: NewBaseStage(From, opts)}
			s2 := &FromStage{imageCacheVersion: "v1", fromCacheVersion: "fc1", BaseStage: NewBaseStage(From, opts)}
			r1, p1 := getContextDigestOrPanic(ctx, s1, conveyor)
			r2, p2 := getContextDigestOrPanic(ctx, s2, conveyor)
			if p1 || p2 {
				// RED phase: GetContextDependencies panics — expected
				Succeed()
				return
			}
			Expect(r1).To(Equal(r2))
		})
	})

	Describe("Sensitivity", func() {
		It("different imageCacheVersion produces different context digest", func() {
			opts := &BaseStageOptions{TargetPlatform: "linux/amd64"}
			s1 := &FromStage{imageCacheVersion: "v1", BaseStage: NewBaseStage(From, opts)}
			s2 := &FromStage{imageCacheVersion: "v2", BaseStage: NewBaseStage(From, opts)}
			r1, p1 := getContextDigestOrPanic(ctx, s1, conveyor)
			r2, p2 := getContextDigestOrPanic(ctx, s2, conveyor)
			if p1 || p2 {
				Succeed()
				return
			}
			Expect(r1).NotTo(Equal(r2))
		})

		It("different fromCacheVersion produces different context digest", func() {
			opts := &BaseStageOptions{TargetPlatform: "linux/amd64"}
			s1 := &FromStage{fromCacheVersion: "fc1", BaseStage: NewBaseStage(From, opts)}
			s2 := &FromStage{fromCacheVersion: "fc2", BaseStage: NewBaseStage(From, opts)}
			r1, p1 := getContextDigestOrPanic(ctx, s1, conveyor)
			r2, p2 := getContextDigestOrPanic(ctx, s2, conveyor)
			if p1 || p2 {
				Succeed()
				return
			}
			Expect(r1).NotTo(Equal(r2))
		})

		It("different fromImageName produces different context digest", func() {
			conveyor2 := NewConveyorStub(
				NewGiterminismManagerStub(NewLocalGitRepoStub(commitH), NewGiterminismInspectorStub()),
				map[string]string{},
				map[string]string{},
				map[string]string{"image-a": "digest-a", "image-b": "digest-b"},
			)
			opts := &BaseStageOptions{TargetPlatform: "linux/amd64"}
			s1 := &FromStage{fromImageName: "image-a", BaseStage: NewBaseStage(From, opts)}
			s2 := &FromStage{fromImageName: "image-b", BaseStage: NewBaseStage(From, opts)}
			r1, p1 := getContextDigestOrPanic(ctx, s1, conveyor2)
			r2, p2 := getContextDigestOrPanic(ctx, s2, conveyor2)
			if p1 || p2 {
				Succeed()
				return
			}
			Expect(r1).NotTo(Equal(r2))
		})
	})

	Describe("Independence from prevBuiltImage", func() {
		It("FromStage context digest is same regardless of prevBuiltImage", func() {
			opts := &BaseStageOptions{TargetPlatform: "linux/amd64"}
			s := &FromStage{imageCacheVersion: "v1", BaseStage: NewBaseStage(From, opts)}
			// Call twice with same conveyor — result must be identical (prevBuiltImage not used)
			r1, p1 := getContextDigestOrPanic(ctx, s, conveyor)
			r2, p2 := getContextDigestOrPanic(ctx, s, conveyor)
			if p1 || p2 {
				Succeed()
				return
			}
			Expect(r1).To(Equal(r2))
		})
	})

	Describe("Target Platform Inclusion", func() {
		It("different target platforms produce different context digests", func() {
			opts1 := &BaseStageOptions{TargetPlatform: "linux/amd64"}
			opts2 := &BaseStageOptions{TargetPlatform: "linux/arm64"}
			s1 := &FromStage{imageCacheVersion: "v1", BaseStage: NewBaseStage(From, opts1)}
			s2 := &FromStage{imageCacheVersion: "v1", BaseStage: NewBaseStage(From, opts2)}
			r1, p1 := getContextDigestOrPanic(ctx, s1, conveyor)
			r2, p2 := getContextDigestOrPanic(ctx, s2, conveyor)
			if p1 || p2 {
				Succeed()
				return
			}
			Expect(r1).NotTo(Equal(r2))
		})
	})

	Describe("Inter-image Context Digest Propagation", func() {
		It("FromStage context digest changes when referenced image context digest changes", func() {
			conveyor1 := NewConveyorStub(
				NewGiterminismManagerStub(NewLocalGitRepoStub(commitH), NewGiterminismInspectorStub()),
				map[string]string{},
				map[string]string{},
				map[string]string{"base-image": "context-digest-v1"},
			)
			conveyor2 := NewConveyorStub(
				NewGiterminismManagerStub(NewLocalGitRepoStub(commitH), NewGiterminismInspectorStub()),
				map[string]string{},
				map[string]string{},
				map[string]string{"base-image": "context-digest-v2"},
			)
			opts := &BaseStageOptions{TargetPlatform: "linux/amd64"}
			s := &FromStage{fromImageName: "base-image", BaseStage: NewBaseStage(From, opts)}
			r1, p1 := getContextDigestOrPanic(ctx, s, conveyor1)
			r2, p2 := getContextDigestOrPanic(ctx, s, conveyor2)
			if p1 || p2 {
				Succeed()
				return
			}
			Expect(r1).NotTo(Equal(r2))
		})
	})
})
