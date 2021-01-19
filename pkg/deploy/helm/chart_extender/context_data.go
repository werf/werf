package chart_extender

import "context"

type ChartExtenderContextData struct {
	chartExtenderContext context.Context
}

func NewChartExtenderContextData(ctx context.Context) *ChartExtenderContextData {
	return &ChartExtenderContextData{chartExtenderContext: ctx}
}
