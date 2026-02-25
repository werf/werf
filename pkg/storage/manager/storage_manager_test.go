package manager

import (
	"context"
	"errors"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestStorageManager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Storage Manager Suite")
}

var _ = Describe("RetryOnUnexpectedStagesStorageState", func() {
	It("should terminate after exhausting max retries", func() {
		ctx := context.Background()
		callCount := 0

		err := RetryOnUnexpectedStagesStorageState(ctx, nil, func() error {
			callCount++
			return ErrUnexpectedStagesStorageState
		})

		Expect(err).To(HaveOccurred())
		Expect(errors.Is(err, ErrUnexpectedStagesStorageState)).To(BeTrue())
		Expect(err.Error()).To(ContainSubstring("exhausted"))
		Expect(callCount).To(Equal(4), "expected 1 initial attempt + 3 retries = 4 total calls")
	})

	It("should succeed when function recovers within retry limit", func() {
		ctx := context.Background()
		callCount := 0

		err := RetryOnUnexpectedStagesStorageState(ctx, nil, func() error {
			callCount++
			if callCount < 3 {
				return ErrUnexpectedStagesStorageState
			}
			return nil
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(callCount).To(Equal(3))
	})

	It("should respect context cancellation", func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		callCount := 0

		err := RetryOnUnexpectedStagesStorageState(ctx, nil, func() error {
			callCount++
			return ErrUnexpectedStagesStorageState
		})

		Expect(errors.Is(err, context.Canceled)).To(BeTrue())
		Expect(callCount).To(Equal(1), "expected only 1 call before context cancellation is detected")
	})

	It("should return non-retry errors immediately without retrying", func() {
		ctx := context.Background()
		callCount := 0
		otherErr := errors.New("some other error")

		err := RetryOnUnexpectedStagesStorageState(ctx, nil, func() error {
			callCount++
			return otherErr
		})

		Expect(errors.Is(err, otherErr)).To(BeTrue())
		Expect(callCount).To(Equal(1))
	})

	It("should return nil when function succeeds on first attempt", func() {
		ctx := context.Background()

		err := RetryOnUnexpectedStagesStorageState(ctx, nil, func() error {
			return nil
		})

		Expect(err).NotTo(HaveOccurred())
	})
})
