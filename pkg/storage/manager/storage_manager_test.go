package manager

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"github.com/werf/werf/v2/pkg/logging"
)

var _ = Describe("RetryOnUnexpectedStorageState", func() {
	DescribeTable("retry behavior",
		func(ctx context.Context, fn func() error, expectedErrMatcher types.GomegaMatcher, expectedCallCount int) {
			ctx = logging.WithLogger(ctx)

			callCount := 0
			err := RetryOnUnexpectedStorageState(ctx, nil, func() error {
				callCount++
				return fn()
			})

			Expect(err).To(expectedErrMatcher)
			Expect(callCount).To(Equal(expectedCallCount))
		},
		Entry("should terminate after exhausting max retries",
			func() error { return ErrUnexpectedStorageState },
			And(MatchError(ErrUnexpectedStorageState), MatchError(ContainSubstring("exhausted"))),
			4,
		),
		Entry("should succeed on first attempt",
			func() error { return nil },
			Succeed(),
			1,
		),
		Entry("should return non-retry errors immediately",
			func() error { return errors.New("some other error") },
			MatchError("some other error"),
			1,
		),
	)

	It("should succeed when function recovers within retry limit", func(ctx context.Context) {
		ctx = logging.WithLogger(ctx)

		callCount := 0
		err := RetryOnUnexpectedStorageState(ctx, nil, func() error {
			callCount++
			if callCount < 3 {
				return ErrUnexpectedStorageState
			}
			return nil
		})

		Expect(err).To(Succeed())
		Expect(callCount).To(Equal(3))
	})

	It("should respect context cancellation", func(ctx context.Context) {
		ctx = logging.WithLogger(ctx)
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()

		callCount := 0
		err := RetryOnUnexpectedStorageState(cancelCtx, nil, func() error {
			callCount++
			return ErrUnexpectedStorageState
		})

		Expect(err).To(MatchError(context.Canceled))
		Expect(callCount).To(Equal(1))
	})
})
