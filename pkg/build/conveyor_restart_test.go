package build

import (
	"context"
	"errors"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/logging"
	"github.com/werf/werf/v2/pkg/opstats"
	"github.com/werf/werf/v2/pkg/storage/manager"
)

var _ = ginkgo.Describe("Conveyor restart counting", func() {
	ginkgo.DescribeTable("counts every conveyor attempt after the first one",
		func(specCtx ginkgo.SpecContext, failures, expectedAttempts, expectedRestarts int) {
			collector := opstats.NewCollector()
			ctx := opstats.NewContext(logging.WithLogger(specCtx), collector)

			var attempts int
			err := retryWithRestartCount(ctx, nil, func() error {
				attempts++
				if attempts <= failures {
					return manager.ErrUnexpectedStagesStorageState
				}
				return nil
			})

			if failures < expectedAttempts {
				gomega.Expect(err).To(gomega.Succeed())
			} else {
				gomega.Expect(err).To(gomega.MatchError(manager.ErrUnexpectedStagesStorageState))
			}
			gomega.Expect(attempts).To(gomega.Equal(expectedAttempts))
			gomega.Expect(eventCounts(collector)[opstats.EventConveyorRestart]).To(gomega.Equal(expectedRestarts))
		},
		ginkgo.Entry("success on the first attempt", 0, 1, 0),
		ginkgo.Entry("one restart after one unexpected state", 1, 2, 1),
		ginkgo.Entry("exhausted retries", 4, 4, 3),
	)

	ginkgo.It("counts nothing for an attempt that a permanent error ends", func(specCtx ginkgo.SpecContext) {
		collector := opstats.NewCollector()
		ctx := opstats.NewContext(logging.WithLogger(specCtx), collector)

		gomega.Expect(retryWithRestartCount(ctx, nil, func() error {
			return errors.New("permanent")
		})).To(gomega.MatchError("permanent"))

		gomega.Expect(eventCounts(collector)).To(gomega.BeEmpty())
	})

	ginkgo.It("counts no restart when the context is canceled during the backoff", func(ctx ginkgo.SpecContext) {
		collector := opstats.NewCollector()
		cancelCtx, cancel := context.WithCancel(opstats.NewContext(logging.WithLogger(ctx), collector))

		var attempts int
		err := retryWithRestartCount(cancelCtx, nil, func() error {
			attempts++
			go func() {
				time.Sleep(100 * time.Millisecond)
				cancel()
			}()
			return manager.ErrUnexpectedStagesStorageState
		})

		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(attempts).To(gomega.Equal(1))
		gomega.Expect(eventCounts(collector)).To(gomega.BeEmpty())
	})
})
