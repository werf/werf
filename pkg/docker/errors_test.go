package docker

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	errorsPkg "github.com/pkg/errors"
)

var _ = Describe("docker errors", func() {
	Describe("IsErrContainerPaused", func() {
		It("should return true if err is \"container is paused\" error", func() {
			causeErr := errors.New("cannot remove container XXXX: container is paused and must be unpaused first")
			err0 := errorsPkg.WithMessage(causeErr, "some text")
			Expect(IsErrContainerPaused(err0)).To(BeTrue())
		})
		It("should return false otherwise", func() {
			err0 := errors.New("some err")
			Expect(IsErrContainerPaused(err0)).To(BeFalse())
		})
	})
	Describe("IsErrContainerRunning", func() {
		It("should return true if err is \"container is paused\" error", func() {
			causeErr := errors.New("cannot remove container XXXX: container is running: stop the container before removing or force remove")
			err0 := errorsPkg.WithMessage(causeErr, "some text")
			Expect(IsErrContainerRunning(err0)).To(BeTrue())
		})
		It("should return false otherwise", func() {
			err0 := errors.New("some err")
			Expect(IsErrContainerRunning(err0)).To(BeFalse())
		})
	})
	Describe("IsErrPruneRunning", func() {
		It("should return true if err is \"prune running\" error", func() {
			causeErr := errors.New("a prune operation is already running")
			err0 := errorsPkg.WithMessage(causeErr, "some text")
			Expect(IsErrPruneRunning(err0)).To(BeTrue())
		})
		It("should return false otherwise", func() {
			err0 := errors.New("some err")
			Expect(IsErrPruneRunning(err0)).To(BeFalse())
		})
	})

	Describe("IsErrContentNotFound", func() {
		DescribeTable("detection",
			func(msg string, expected bool) {
				err := errors.New(msg)
				Expect(IsErrContentNotFound(err)).To(Equal(expected))
			},
			Entry("matches full Docker daemon error",
				"failed to push image: content digest sha256:abc123def456: not found", true),
			Entry("matches minimal form",
				"content digest sha256:0000: not found", true),
			Entry("does not match missing 'content digest'",
				"sha256:abc123 not found", false),
			Entry("does not match missing 'not found'",
				"content digest sha256:abc123 exists", false),
			Entry("does not match unrelated error",
				"connection timeout", false),
			Entry("returns false for nil", "", false),
		)

		It("works through pkg/errors wrapping", func() {
			cause := errors.New("content digest sha256:deadbeef: not found")
			wrapped := errorsPkg.WithMessage(cause, "push failed")
			Expect(IsErrContentNotFound(wrapped)).To(BeTrue())
		})
	})

	Describe("ContentNotFoundDigest", func() {
		DescribeTable("extraction",
			func(msg, expectedDigest string) {
				err := errors.New(msg)
				Expect(ContentNotFoundDigest(err)).To(Equal(expectedDigest))
			},
			Entry("extracts full sha256 digest",
				"content digest sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2: not found",
				"sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"),
			Entry("extracts short digest",
				"content digest sha256:abcdef0123456789: not found",
				"sha256:abcdef0123456789"),
			Entry("returns empty when regex doesn't match",
				"content digest unknown: not found",
				""),
			Entry("returns empty for nil", "", ""),
		)

		It("works through pkg/errors wrapping", func() {
			cause := errors.New("content digest sha256:deadbeef: not found")
			wrapped := errorsPkg.WithMessage(cause, "push failed")
			Expect(ContentNotFoundDigest(wrapped)).To(Equal("sha256:deadbeef"))
		})
	})
})
