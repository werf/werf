package host_cleaning

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("host lock", func() {
	var ctx context.Context
	BeforeEach(func() {
		ctx = context.Background()
	})
	Describe("withHostLockOrNothing", func() {
		It("should call callback function if lock is acquired", func() {
			spy := &spyHostLock{}
			lockName := "test"
			err := withHostLockOrNothing(ctx, lockName, spy.Handle1)
			Expect(err).To(Succeed())
			Expect(spy.callsCount).To(Equal(1))
		})
		It("should not call callback function if lock isn't acquired", func() {
			spy := &spyHostLock{}
			lockName := "test"
			err := withHostLockOrNothing(ctx, lockName, func() error {
				return withHostLockOrNothing(ctx, lockName, spy.Handle1) // lock is already acquired in parent function
			})
			Expect(err).To(Succeed())
			Expect(spy.callsCount).To(Equal(0))
		})
	})
})

type spyHostLock struct {
	callsCount int
	err        error
}

func (s *spyHostLock) Handle1() error {
	s.callsCount++
	return s.err
}
