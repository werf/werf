package common

import (
	"bytes"
	"context"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/samber/lo"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/level"
	"github.com/werf/werf/v3/pkg/build"
	"github.com/werf/werf/v3/pkg/config"
	"github.com/werf/werf/v3/pkg/giterminism_manager"
	"github.com/werf/werf/v3/pkg/storage/manager"
)

var _ = ginkgo.Describe("conveyor retries under a command-root collector", func() {
	// The retry wrapper counts restarts into the collector of the context it was called with,
	// so a command that enters the wrapper without InitOperationsStatistics loses the count.
	ginkgo.It("counts a restart into the collector installed at the command root", func() {
		var out bytes.Buffer
		logger := logboek.NewLogger(&out, &out)
		logger.SetAcceptedLevel(level.Debug)

		ctx, finish := InitOperationsStatistics(
			logboek.NewContext(context.Background(), logger),
			&CmdData{BuildReportOperations: lo.ToPtr(false)},
		)

		wrapper := build.NewConveyorWithRetryWrapper(
			&config.WerfConfig{Meta: &config.Meta{Project: "test"}},
			(*giterminism_manager.Manager)(nil),
			ginkgo.GinkgoT().TempDir(), ginkgo.GinkgoT().TempDir(),
			nil, nil, nil,
			build.ConveyorOptions{},
		)

		var attempts int
		gomega.Expect(wrapper.WithRetryBlock(ctx, func(_ *build.Conveyor) error {
			attempts++
			if attempts == 1 {
				return manager.ErrUnexpectedStagesStorageState
			}
			return nil
		})).To(gomega.Succeed())
		gomega.Expect(attempts).To(gomega.Equal(2))

		finish()
		gomega.Expect(out.String()).To(gomega.ContainSubstring("Recovery: 0 broken stage detections, 1 conveyor restarts"))
	})
})
