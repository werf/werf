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
})
