package build

import (
	"context"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/level"
	"github.com/werf/werf/v2/pkg/opstats"
)

var _ = Describe("Conveyor operations collector gate", func() {
	newCtx := func(acceptedLevel level.Level) context.Context {
		logger := logboek.NewLogger(io.Discard, io.Discard)
		logger.SetAcceptedLevel(acceptedLevel)
		return logboek.NewContext(context.Background(), logger)
	}

	DescribeTable("newOperationsCollector",
		func(acceptedLevel level.Level, forceEnabled, expectCollector bool) {
			c := &Conveyor{}
			ctx, collector, _ := c.newOperationsCollector(newCtx(acceptedLevel), forceEnabled)

			if !expectCollector {
				Expect(collector).To(BeNil())
				Expect(opstats.FromContext(ctx)).To(BeNil())
				return
			}

			Expect(collector).NotTo(BeNil())
			Expect(opstats.FromContext(ctx)).To(BeIdenticalTo(collector))
		},
		Entry("disabled without debug logging and without the flag", level.Default, false, false),
		Entry("enabled by the flag without debug logging", level.Default, true, true),
		Entry("enabled by debug logging without the flag", level.Debug, false, true),
		Entry("enabled by both", level.Debug, true, true),
	)
})
