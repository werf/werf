package manager

import (
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/lockgate"
)

var _ = Describe("StorageManager shared host image locks", func() {
	It("appends under the lock", func(ctx SpecContext) {
		m := &StorageManager{}
		m.sharedHostImagesLocksMu.Lock()

		appended := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			m.addSharedHostImageLock(lockgate.LockHandle{})
			close(appended)
		}()

		Consistently(appended, "100ms").ShouldNot(BeClosed())
		m.sharedHostImagesLocksMu.Unlock()

		Eventually(appended).Should(BeClosed())
		Expect(m.SharedHostImagesLocks).To(HaveLen(1))
	})

	It("keeps every lock taken by concurrent image builds", func() {
		const goroutines = 50

		m := &StorageManager{}

		var wg sync.WaitGroup
		wg.Add(goroutines)
		for i := 0; i < goroutines; i++ {
			go func() {
				defer GinkgoRecover()
				defer wg.Done()
				m.addSharedHostImageLock(lockgate.LockHandle{})
			}()
		}
		wg.Wait()

		Expect(m.SharedHostImagesLocks).To(HaveLen(goroutines))
	})
})
