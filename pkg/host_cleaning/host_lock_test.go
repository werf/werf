package host_cleaning

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("withHostLockOrNothing", func() {
	var ctx context.Context
	var spy *spyHostLockCallback
	BeforeEach(func() {
		ctx = context.Background()
		spy = &spyHostLockCallback{}
	})
	It("should call callback function if lock is acquired", func() {
		lockName := "test"
		err := withHostLockOrNothing(ctx, lockName, spy.Method)
		Expect(err).To(Succeed())
		Expect(spy.callsCount).To(Equal(1))
	})
	It("should not call callback function if lock isn't acquired", func() {
		lockName := "test"
		err := withHostLockOrNothing(ctx, lockName, func() error {
			return withHostLockOrNothing(ctx, lockName, spy.Method) // lock is already acquired in parent function
		})
		Expect(err).To(Succeed())
		Expect(spy.callsCount).To(Equal(0))
	})
})

type spyHostLockCallback struct {
	callsCount int
}

func (s *spyHostLockCallback) Method() error {
	s.callsCount++
	return nil
}
