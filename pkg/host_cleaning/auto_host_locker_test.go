package host_cleaning

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("auto host locker", Serial, func() {
	var locker *autoHostCleanupLocker
	var anotherProcessLocker *autoHostCleanupLocker
	var ctx context.Context
	BeforeEach(func() {
		var err error
		locker, err = newAutoHostCleanupLocker()
		Expect(err).To(Succeed())
		anotherProcessLocker, err = newAutoHostCleanupLocker()
		Expect(err).To(Succeed())
		ctx = context.Background()
	})
	Describe("isLockFree", func() {
		It("should return false if another process is already took the lock", func() {
			Expect(anotherProcessLocker.lock.Lock()).To(Succeed())

			ok, err := locker.isLockFree(ctx)
			Expect(err).To(Succeed())
			Expect(ok).To(BeFalse())

			Expect(locker.lock.Locked()).To(BeFalse())
			Expect(anotherProcessLocker.lock.Unlock()).To(Succeed())
		})
		It("should return true and release the acquired lock otherwise", func() {
			ok, err := locker.isLockFree(ctx)
			Expect(err).To(Succeed())
			Expect(ok).To(BeTrue())

			Expect(locker.lock.Locked()).To(BeFalse())
		})
	})
	Describe("WithLockOrNothing", func() {
		It("should do nothing if lock is not acquired", func() {
			Expect(anotherProcessLocker.lock.Lock()).To(Succeed())

			spy := spyHostLock{}

			err := locker.WithLockOrNothing(ctx, spy.Handle1)
			Expect(err).To(Succeed())
			Expect(spy.callsCount).To(Equal(0))

			Expect(anotherProcessLocker.lock.Unlock()).To(Succeed())
		})
		It("should execute callback and release lock after", func() {
			spy := spyHostLock{err: errors.New("some err")}

			err := locker.WithLockOrNothing(ctx, spy.Handle1)
			Expect(errors.Is(err, spy.err)).To(BeTrue())
			Expect(locker.lock.Locked()).To(BeFalse())
		})
	})
})
