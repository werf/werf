package chart_extender_for_werf_helm

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/3p-helm-for-werf-helm/pkg/chart/loader"
	"github.com/werf/werf/v2/pkg/werf"
)

var _ = Describe("CopyChartDependenciesIntoCache", func() {
	It("should succeed for a chart with subcharts in charts/ but without Chart.lock", func(ctx context.Context) {
		tmpDir := GinkgoT().TempDir()
		Expect(werf.Init(tmpDir, filepath.Join(tmpDir, "home"))).To(Succeed())

		chartDir := filepath.Join(tmpDir, "chart")
		Expect(os.MkdirAll(filepath.Join(chartDir, "charts", "child"), os.ModePerm)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(chartDir, "Chart.yaml"), []byte("apiVersion: v2\nname: parent\nversion: 0.1.0\n"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(chartDir, "charts", "child", "Chart.yaml"), []byte("apiVersion: v2\nname: child\nversion: 0.1.0\n"), 0o644)).To(Succeed())

		ch, err := loader.LoadDir(chartDir)
		Expect(err).To(Succeed())
		Expect(ch.Lock).To(BeNil())

		Expect(CopyChartDependenciesIntoCache(ctx, chartDir, ch)).To(Succeed())
	})
})
