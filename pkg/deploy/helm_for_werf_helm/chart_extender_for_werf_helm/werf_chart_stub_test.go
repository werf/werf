package chart_extender_for_werf_helm

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	chart "github.com/werf/3p-helm-for-werf-helm/pkg/chart"
	"github.com/werf/werf/v2/pkg/logging"
)

var _ = DescribeTable("WerfChartStub refuses charts with .werf-data",
	func(ctx context.Context, files []*chart.ChartExtenderBufferedFile, wantErr string) {
		ctx = logging.WithLogger(ctx)

		stub := NewWerfChartStub(ctx, false)
		Expect(stub.ChartCreated(&chart.Chart{})).To(Succeed())

		err := stub.ChartLoaded(files)
		if wantErr == "" {
			Expect(err).To(Succeed())

			return
		}

		Expect(err).To(MatchError(ContainSubstring(wantErr)))
	},
	Entry("chart without .werf-data",
		[]*chart.ChartExtenderBufferedFile{{Name: "values.yaml", Data: []byte("key: value\n")}},
		"",
	),
	Entry("chart with .werf-data apiVersion v2",
		[]*chart.ChartExtenderBufferedFile{{Name: ".werf-data/metadata.yaml", Data: []byte("apiVersion: v2\n")}},
		`this chart has .werf-data/metadata.yaml with apiVersion "v2": it requires werf v3 or Nelm v2, or newer`,
	),
	Entry(".werf-data metadata without apiVersion",
		[]*chart.ChartExtenderBufferedFile{{Name: ".werf-data/metadata.yaml", Data: []byte("values: {}\n")}},
		"",
	),
)
