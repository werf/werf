package graceful

import (
	"sync"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = Describe("Graceful", func() {
	BeforeEach(func() {
		terminationErrChan = make(chan *terminationError, 1)
	})

	Describe("Terminate", func() {
		It("should panic", func() {
			expected := newTerminationError("some err", 1)
			gomega.Expect(func() {
				Terminate(expected.error.Error(), expected.code)
			}).To(gomega.PanicWith(expected))
		})
	})
	Describe("TerminateGoroutine", func() {
		It("should send error to terminationErrChan", func() {
			expected := newTerminationError("some err", 1)

			runAndWaitGoroutine(func() {
				TerminateGoroutine(expected.error.Error(), expected.code)
			})

			actual := <-terminationErrChan
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
			expected := newTerminationError("some err", 1)

			defer func() {
				gomega.Expect(spy.callsCount).To(gomega.Equal(1))
				gomega.Expect(spy.message).To(gomega.Equal(expected.error.Error()))
				gomega.Expect(spy.code).To(gomega.Equal(expected.code))
			}()
			defer Shutdown()

			Terminate(expected.error.Error(), expected.code)
		})
		It("should handle panic from child goroutine", func() {
			expected := newTerminationError("some err", 1)

			defer func() {
				gomega.Expect(spy.callsCount).To(gomega.Equal(1))
				gomega.Expect(spy.message).To(gomega.Equal(expected.error.Error()))
				gomega.Expect(spy.code).To(gomega.Equal(expected.code))
			}()
			defer Shutdown()

			runAndWaitGoroutine(func() {
				TerminateGoroutine(expected.error.Error(), expected.code)
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
