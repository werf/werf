package host_cleaning

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	chart "github.com/werf/common-go/pkg/lock"
	"github.com/werf/lockgate"
)

var _ = Describe("auto host locker", Ordered, func() {
	var locker *autoHostCleanupLocker
	var ctx context.Context
	BeforeEach(func() {
		locker = newAutoHostCleanupLocker()
		ctx = context.Background()
	})
	Describe("TryLockOrNothing", func() {
		It("should not call given callback if lock is not acquired", func() {
			lock := acquireLock(ctx, locker, BeTrue())

			spy := &spyHostLock{}
			ok, err := locker.TryLockOrNothing(ctx, spy.Handle2)
			Expect(err).To(Succeed())
			Expect(ok).To(BeFalse())
			Expect(spy.callsCount).To(Equal(0))

			releaseLock(ctx, lock)
		})
		It("should release acquired lock if callback returns err", func() {
			spy := &spyHostLock{err: errors.New("some err")}

			ok, err := locker.TryLockOrNothing(ctx, spy.Handle2)
			Expect(errors.Is(err, spy.err)).To(BeTrue())
			Expect(ok).To(BeFalse())
			Expect(spy.callsCount).To(Equal(1))

			lock := acquireLock(ctx, locker, BeTrue())
			releaseLock(ctx, lock)
		})
		It("should release acquired lock if callback returns false", func() {
			spy := &spyHostLock{ok: false}

			ok, err := locker.TryLockOrNothing(ctx, spy.Handle2)
			Expect(err).To(Succeed())
			Expect(ok).To(BeFalse())
			Expect(spy.callsCount).To(Equal(1))

			lock := acquireLock(ctx, locker, BeTrue())
			releaseLock(ctx, lock)
		})
		It("should call given callback and acquire lock", func() {
			spy := &spyHostLock{ok: true}

			ok, err := locker.TryLockOrNothing(ctx, spy.Handle2)
			Expect(err).To(Succeed())
			Expect(ok).To(BeTrue())
			Expect(spy.callsCount).To(Equal(1))

			acquireLock(ctx, locker, BeFalse())
			releaseLock(ctx, locker.lockHandle)
		})
	})
	Describe("Unlock", func() {
		// TODO (zaytsev):  cover it
	})
})

func acquireLock(ctx context.Context, locker *autoHostCleanupLocker, matcher types.GomegaMatcher) lockgate.LockHandle {
	acquired, lock, err := chart.AcquireHostLock(ctx, locker.lockName, locker.lockOptions)
	Expect(err).To(Succeed())
	Expect(acquired).To(matcher)
	return lock
}

func releaseLock(_ context.Context, lock lockgate.LockHandle) {
	err := chart.ReleaseHostLock(lock)
	Expect(err).To(Succeed())
}
