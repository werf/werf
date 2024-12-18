package helpers_for_werf_helm

import "context"

type ChartExtenderContextData struct {
	ChartExtenderContext context.Context
}

func NewChartExtenderContextData(ctx context.Context) *ChartExtenderContextData {
	return &ChartExtenderContextData{ChartExtenderContext: ctx}
}
