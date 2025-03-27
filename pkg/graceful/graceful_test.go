package graceful

import (
	"sync"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = Describe("Graceful", func() {
	BeforeEach(func() {
		panicChan = make(chan *panicError, 1)
	})

	Describe("Panic", func() {
		It("should panic", func() {
			expected := newPanicError("some err", 1)
			gomega.Expect(func() {
				Panic(expected.error.Error(), expected.code)
			}).To(gomega.PanicWith(expected))
		})
	})
	Describe("PanicGoroutine", func() {
		It("should send error to panicChan", func() {
			expected := newPanicError("some err", 1)

			runAndWaitGoroutine(func() {
				PanicGoroutine(expected.error.Error(), expected.code)
			})

			actual := <-panicChan
			gomega.Expect(actual).To(gomega.Equal(expected))
		})
	})
	Describe("Shutdown", func() {
		var spy *spyFatal
		BeforeEach(func() {
			spy = &spyFatal{}
			fatal = spy.Method
		})
		It("should handle panic from main goroutine", func() {
			expected := newPanicError("some err", 1)

			defer func() {
				gomega.Expect(spy.callsCount).To(gomega.Equal(1))
				gomega.Expect(spy.message).To(gomega.Equal(expected.error.Error()))
				gomega.Expect(spy.code).To(gomega.Equal(expected.code))
			}()
			defer Shutdown()

			Panic(expected.error.Error(), expected.code)
		})
		It("should handle panic from child goroutine", func() {
			expected := newPanicError("some err", 1)

			defer func() {
				gomega.Expect(spy.callsCount).To(gomega.Equal(1))
				gomega.Expect(spy.message).To(gomega.Equal(expected.error.Error()))
				gomega.Expect(spy.code).To(gomega.Equal(expected.code))
			}()
			defer Shutdown()

			runAndWaitGoroutine(func() {
				PanicGoroutine(expected.error.Error(), expected.code)
			})
		})
	})
})

func runAndWaitGoroutine(callback func()) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		callback()
	}()
	wg.Wait()
}

type spyFatal struct {
	callsCount int
	message    string
	code       int
}

func (s *spyFatal) Method(message string, code int) {
	s.callsCount++
	s.message = message
	s.code = code
}
