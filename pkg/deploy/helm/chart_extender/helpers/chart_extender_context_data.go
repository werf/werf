package helpers

import "context"

type ChartExtenderContextData struct {
	ChartExtenderContext context.Context
}

func NewChartExtenderContextData(ctx context.Context) *ChartExtenderContextData {
	return &ChartExtenderContextData{ChartExtenderContext: ctx}
}
