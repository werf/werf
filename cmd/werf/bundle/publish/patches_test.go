package publish

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/nelm/v2/pkg/action"
	nelmcommon "github.com/werf/nelm/v2/pkg/common"
	chart "github.com/werf/nelm/v2/pkg/helm/pkg/chart/v2"
)

var _ = Describe("published bundle", func() {
	It("renders with the publish-time --patches like converge with them", func(ctx context.Context) {
		chartDir := GinkgoT().TempDir()
		writeFile(filepath.Join(chartDir, "Chart.yaml"), "apiVersion: v2\nname: root\nversion: 1.0.0\n")
		writeFile(filepath.Join(chartDir, "templates", "cm.yaml"), "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: cm\ndata:\n  key: value\n")
		writeFile(filepath.Join(chartDir, "patches.yaml"), "renderPatches:\n- patch: .metadata.labels.chart = \"chart\"\n")

		patchesFile := writeFile(filepath.Join(GinkgoT().TempDir(), "patches.yaml"), "renderPatches:\n- patch: .metadata.labels.publish = \"publish\"\n")

		// converge adds the same annotations as werf bundle publish records.
		converged := renderWithPatches(ctx, chartDir, nelmcommon.LegacyChartTypeChart, []string{patchesFile}, map[string]string{
			"project.werf.io/name": "proj",
			"project.werf.io/env":  "",
		})

		prevCommonCmdData := commonCmdData
		DeferCleanup(func() { commonCmdData = prevCommonCmdData })
		commonCmdData.PatchesFiles = []string{patchesFile}

		bundleDir := publishEquivalenceBundle(ctx, chartDir, nelmcommon.ValuesOptions{}, nil, nil)
		Expect(os.Remove(patchesFile)).To(Succeed())

		applied := renderWithPatches(ctx, bundleDir, nelmcommon.LegacyChartTypeBundle, nil, nil)

		Expect(applied).To(HaveLen(1))
		Expect(applied[0]["metadata"]).To(HaveKeyWithValue("labels", map[string]interface{}{"chart": "chart", "publish": "publish"}))
		Expect(applied).To(Equal(converged))
	})
})

func renderWithPatches(ctx context.Context, chartDir string, chartType nelmcommon.LegacyChartType, patchesFiles []string, extraRuntimeAnnotations map[string]string) []map[string]interface{} {
	result, err := action.ChartRender(ctx, action.ChartRenderOptions{
		Chart:                   chartDir,
		DefaultChartAPIVersion:  chart.APIVersionV2,
		ExtraRuntimeAnnotations: extraRuntimeAnnotations,
		LegacyChartType:         chartType,
		LocalKubeVersion:        "1.30.0",
		OutputNoPrint:           true,
		PatchesFiles:            patchesFiles,
		ReleaseName:             "release",
		ReleaseNamespace:        "namespace",
		TempDirPath:             GinkgoT().TempDir(),
	})
	Expect(err).To(Succeed())

	var objects []map[string]interface{}
	for _, resource := range result.Resources {
		objects = append(objects, resource.Unstruct.Object)
	}

	return objects
}
