package publish

import (
	"bytes"
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	"github.com/werf/common-go/pkg/secrets_manager"
	"github.com/werf/nelm/v2/pkg/action"
	nelmchart "github.com/werf/nelm/v2/pkg/chart"
	nelmcommon "github.com/werf/nelm/v2/pkg/common"
	chartcommon "github.com/werf/nelm/v2/pkg/helm/pkg/chart/common"
	"github.com/werf/nelm/v2/pkg/helm/pkg/chart/loader"
	chart "github.com/werf/nelm/v2/pkg/helm/pkg/chart/v2"
	chartv2util "github.com/werf/nelm/v2/pkg/helm/pkg/chart/v2/util"
	"github.com/werf/werf/v3/pkg/config"
)

type convergeEquivalenceCase struct {
	chartYAMLDependencies string
	chartValues           string
	chartSecretValues     string
	subchartValues        string
	values                string
	secretValues          string
	set                   []string
	setJSON               []string
	setFile               string
	serviceValues         string

	applyValues                     string
	applySet                        []string
	applyDisableDefaultSecretValues bool

	expectedValues string
}

var _ = Describe("published bundle", func() {
	DescribeTable("renders like converge with the publish flags followed by the apply flags",
		func(c convergeEquivalenceCase) {
			ctx := context.Background()
			GinkgoT().Setenv("WERF_SECRET_KEY", string(generateSecretKey()))

			chartDir := writeEquivalenceChart(ctx, c)
			flagsDir := GinkgoT().TempDir()

			var valuesFiles, secretValuesFiles, applyValuesFiles, setFile []string
			if c.values != "" {
				valuesFiles = []string{writeFile(filepath.Join(flagsDir, "values.yaml"), c.values)}
			}
			if c.secretValues != "" {
				secretValuesFiles = []string{writeFile(filepath.Join(flagsDir, "secret-values.yaml"), encryptValues(ctx, c.secretValues))}
			}
			if c.applyValues != "" {
				applyValuesFiles = []string{writeFile(filepath.Join(flagsDir, "apply-values.yaml"), c.applyValues)}
			}
			if c.setFile != "" {
				setFile = []string{"file=" + writeFile(filepath.Join(flagsDir, "file.txt"), c.setFile)}
			}

			serviceValues := parseValues(c.serviceValues)

			converged := renderEquivalenceChart(ctx, chartDir, nelmcommon.LegacyChartTypeChart, nelmcommon.ValuesOptions{
				ValuesFiles:   append(append([]string{}, valuesFiles...), applyValuesFiles...),
				ValuesSet:     append(append([]string{}, c.set...), c.applySet...),
				ValuesSetJSON: c.setJSON,
				ValuesSetFile: setFile,
			}, secretValuesFiles, serviceValues, c.applyDisableDefaultSecretValues)

			bundleDir := publishEquivalenceBundle(ctx, chartDir, nelmcommon.ValuesOptions{
				ValuesFiles:   valuesFiles,
				ValuesSet:     c.set,
				ValuesSetJSON: c.setJSON,
				ValuesSetFile: setFile,
			}, secretValuesFiles, serviceValues)

			Expect(os.RemoveAll(flagsDir)).To(Succeed())
			if c.applyValues != "" {
				applyValuesFiles = []string{writeFile(filepath.Join(flagsDir, "apply-values.yaml"), c.applyValues)}
			}

			applied := renderEquivalenceChart(ctx, bundleDir, nelmcommon.LegacyChartTypeBundle, nelmcommon.ValuesOptions{
				ValuesFiles: applyValuesFiles,
				ValuesSet:   c.applySet,
			}, nil, nil, c.applyDisableDefaultSecretValues)

			Expect(renderedValues(applied)).To(Equal(renderedValues(converged)))
			Expect(renderedValues(converged)).To(Equal(parseValues(c.expectedValues)))
			Expect(applied.ReleaseConfig).To(Equal(converged.ReleaseConfig))
		},
		Entry("1: --set beats an export from the chart values.yaml", convergeEquivalenceCase{
			chartYAMLDependencies: "- name: sub\n  version: 1.0.0\n  export-values:\n  - parent: g\n    child: x",
			chartValues:           "g: chart",
			set:                   []string{"sub.x=set"},

			expectedValues: "g: chart\nsub:\n  global: {}\n  x: set",
		}),
		Entry("1: an import from --values beats the own key of the parent in the chart values.yaml", convergeEquivalenceCase{
			chartYAMLDependencies: "- name: sub\n  version: 1.0.0\n  import-values:\n  - child: fromchild\n    parent: imported",
			chartValues:           "imported:\n  k: chart",
			values:                "sub:\n  fromchild:\n    k: values",

			expectedValues: "imported:\n  k: values\nsub:\n  fromchild:\n    k: values\n  global: {}",
		}),
		Entry("2: an import from --set beats the chart secret-values.yaml", convergeEquivalenceCase{
			chartYAMLDependencies: "- name: sub\n  version: 1.0.0\n  import-values:\n  - child: fromchild\n    parent: imported",
			chartSecretValues:     "imported:\n  k: chart-secret",
			set:                   []string{"sub.fromchild.k=set"},

			expectedValues: "imported:\n  k: set\nsub:\n  fromchild:\n    k: set\n  global: {}",
		}),
		Entry("3: --secret-values beat an export from the chart secret-values.yaml", convergeEquivalenceCase{
			chartYAMLDependencies: "- name: sub\n  version: 1.0.0\n  export-values:\n  - parent: g\n    child: x",
			chartSecretValues:     "g: chart-secret",
			secretValues:          "sub:\n  x: custom-secret",

			expectedValues: "g: chart-secret\nsub:\n  global: {}\n  x: custom-secret",
		}),
		Entry("3: --values beat an export from the chart secret-values.yaml", convergeEquivalenceCase{
			chartYAMLDependencies: "- name: sub\n  version: 1.0.0\n  export-values:\n  - parent: g\n    child: x",
			chartSecretValues:     "g: chart-secret",
			values:                "sub:\n  x: values",

			expectedValues: "g: chart-secret\nsub:\n  global: {}\n  x: values",
		}),
		Entry("4: an export to a disabled dependency keeps the secret values", convergeEquivalenceCase{
			chartYAMLDependencies: "- name: sub\n  version: 1.0.0\n  condition: sub.enabled\n  export-values:\n  - parent: g\n    child: x",
			chartValues:           "sub:\n  enabled: false",
			chartSecretValues:     "sub:\n  x: chart-secret",
			set:                   []string{"g=set"},

			expectedValues: "g: set\nsub:\n  enabled: false\n  x: chart-secret",
		}),
		Entry("5: --set null deletes the values exported from --secret-values", convergeEquivalenceCase{
			chartYAMLDependencies: "- name: sub\n  version: 1.0.0\n  export-values:\n  - parent: g\n    child: x",
			chartValues:           "sub:\n  a: chart",
			subchartValues:        "own: sub",
			secretValues:          "g: custom-secret",
			set:                   []string{"sub=null"},

			expectedValues: "g: custom-secret\nsub:\n  global: {}\n  own: sub",
		}),
		Entry("6: --set-json replaces a map with the values exported from the secret values", convergeEquivalenceCase{
			chartYAMLDependencies: "- name: sub\n  version: 1.0.0\n  export-values:\n  - parent: g\n    child: x",
			chartSecretValues:     "g: chart-secret",
			secretValues:          "g: custom-secret",
			setJSON:               []string{`sub={"z":1}`},

			expectedValues: "g: custom-secret\nsub:\n  global: {}\n  x: chart-secret\n  z: 1",
		}),
		Entry("7: a --secret-values null replaces a map exported from --values", convergeEquivalenceCase{
			chartYAMLDependencies: "- name: sub\n  version: 1.0.0\n  export-values:\n  - parent: g\n    child: x",
			values:                "g:\n  k: values",
			secretValues:          "g: null",

			expectedValues: "g: null\nsub:\n  global: {}\n  x: null",
		}),
		Entry("8: --set patches a list from --secret-values", convergeEquivalenceCase{
			secretValues: "l:\n- custom-secret-0\n- custom-secret-1",
			set:          []string{"l[1]=set"},

			expectedValues: "l:\n- custom-secret-0\n- set",
		}),
		Entry("9: a null in --values for a key absent in the charts is kept", convergeEquivalenceCase{
			values: "unknown: null\nknown: values",

			expectedValues: "known: values\nunknown: null",
		}),
		Entry("10: --values replace a null in the chart secret-values.yaml", convergeEquivalenceCase{
			chartValues:       "a:\n  c: chart",
			chartSecretValues: "a: null",
			values:            "a:\n  b: values",

			expectedValues: "a:\n  b: values",
		}),
		Entry("11: --values overridden by --secret-values are kept without the chart secret-values.yaml", convergeEquivalenceCase{
			chartSecretValues:               "b: chart-secret",
			values:                          "a: values",
			secretValues:                    "a: custom-secret",
			applyDisableDefaultSecretValues: true,

			expectedValues: "a: custom-secret",
		}),
		Entry("12: --set null deletes a value of the dependency", convergeEquivalenceCase{
			chartYAMLDependencies: "- name: sub\n  version: 1.0.0",
			subchartValues:        "key: sub\nother: sub",
			set:                   []string{"sub.key=null"},

			expectedValues: "sub:\n  global: {}\n  other: sub",
		}),
		Entry("service values", convergeEquivalenceCase{
			chartYAMLDependencies: "- name: sub\n  version: 1.0.0\n  export-values:\n  - parent: g\n    child: x",
			chartValues:           "g: chart",
			set:                   []string{"sub.x=set"},
			serviceValues:         "werf:\n  repo: service\nsub:\n  x: service",

			expectedValues: "g: chart\nsub:\n  global: {}\n  x: service\nwerf:\n  repo: service",
		}),
		Entry("--set-file keeps the file contents after the file is gone", convergeEquivalenceCase{
			setFile: "from-file",

			expectedValues: "file: from-file",
		}),
		Entry("apply --set patches a list from publish --values", convergeEquivalenceCase{
			values:   "l:\n- values-0\n- values-1",
			applySet: []string{"l[1]=apply"},

			expectedValues: "l:\n- values-0\n- apply",
		}),
		Entry("apply --set beats publish --set", convergeEquivalenceCase{
			set:      []string{"a=publish", "b=publish"},
			applySet: []string{"a=apply"},

			expectedValues: "a: apply\nb: publish",
		}),
		Entry("apply --values beat publish --values", convergeEquivalenceCase{
			values:      "a: publish\nb: publish",
			applyValues: "a: apply",

			expectedValues: "a: apply\nb: publish",
		}),
	)
})

func writeEquivalenceChart(ctx context.Context, c convergeEquivalenceCase) string {
	chartDir := GinkgoT().TempDir()

	chartYAML := "apiVersion: v2\nname: root\nversion: 1.0.0\n"
	if c.chartYAMLDependencies != "" {
		chartYAML += "dependencies:\n" + c.chartYAMLDependencies + "\n"

		writeFile(filepath.Join(chartDir, "charts", "sub", "Chart.yaml"), "apiVersion: v2\nname: sub\nversion: 1.0.0\n")
		writeFile(filepath.Join(chartDir, "charts", "sub", "values.yaml"), c.subchartValues)
	}

	writeFile(filepath.Join(chartDir, "Chart.yaml"), chartYAML)
	writeFile(filepath.Join(chartDir, "values.yaml"), c.chartValues)

	if c.chartSecretValues != "" {
		writeFile(filepath.Join(chartDir, "secret-values.yaml"), encryptValues(ctx, c.chartSecretValues))
	}

	return chartDir
}

// publishEquivalenceBundle packs the chart like werf bundle publish and unpacks it like werf
// bundle apply pulls it.
func publishEquivalenceBundle(ctx context.Context, chartDir string, valuesOpts nelmcommon.ValuesOptions, secretValuesFiles []string, serviceValues map[string]interface{}) string {
	prevCommonCmdData := commonCmdData
	DeferCleanup(func() { commonCmdData = prevCommonCmdData })

	commonCmdData.ValuesOptions = valuesOpts
	commonCmdData.SecretValuesFiles = secretValuesFiles

	opts, err := newChartPackOptions(ctx, &config.WerfConfig{Meta: &config.Meta{Project: "proj"}}, serviceValues)
	Expect(err).To(Succeed())

	opts.Name = "bundle"
	opts.Version = "1.0.0"
	opts.Destination = GinkgoT().TempDir()

	result, err := action.ChartPack(ctx, chartDir, opts)
	Expect(err).To(Succeed())

	archive, err := os.ReadFile(result.ArchivePath)
	Expect(err).To(Succeed())

	loaded, err := loader.LoadArchive(nelmcommon.ContextWithHelmOptions(ctx, nelmcommon.HelmOptions{
		ChartLoadOpts: nelmcommon.ChartLoadOptions{NoSecrets: true},
	}), bytes.NewReader(archive))
	Expect(err).To(Succeed())

	bundleDir := filepath.Join(GinkgoT().TempDir(), "bundle")
	Expect(chartv2util.SaveIntoDir(loaded.(*chart.Chart), bundleDir)).To(Succeed())

	return bundleDir
}

func renderEquivalenceChart(ctx context.Context, chartDir string, chartType nelmcommon.LegacyChartType, valuesOpts nelmcommon.ValuesOptions, secretValuesFiles []string, serviceValues map[string]interface{}, defaultSecretValuesDisable bool) *nelmchart.RenderChartResult {
	result, err := nelmchart.RenderChart(ctx, chartDir, "release", "namespace", 1, nelmcommon.DeployTypeInstall, nil, nil, nelmchart.RenderChartOptions{
		ValuesOptions: valuesOpts,
		HelmOptions: nelmcommon.HelmOptions{
			ChartLoadOpts: nelmcommon.ChartLoadOptions{
				ChartType:                  chartType,
				DefaultChartAPIVersion:     chart.APIVersionV2,
				DefaultSecretValuesDisable: defaultSecretValuesDisable,
				ExtraValues:                serviceValues,
				SecretValuesFiles:          secretValuesFiles,
				SecretWorkDir:              chartDir,
			},
		},
		LocalKubeVersion: "1.30.0",
	})
	Expect(err).To(Succeed())

	return result
}

func renderedValues(result *nelmchart.RenderChartResult) map[string]interface{} {
	vals, ok := result.Values["Values"].(chartcommon.Values)
	Expect(ok).To(BeTrue())

	return vals.AsMap()
}

func parseValues(data string) map[string]interface{} {
	vals := map[string]interface{}{}
	Expect(yaml.Unmarshal([]byte(data), &vals)).To(Succeed())

	return vals
}

func encryptValues(ctx context.Context, plain string) string {
	encoder, err := secrets_manager.Manager.GetYamlEncoder(ctx, "", false)
	Expect(err).To(Succeed())

	encrypted, err := encoder.EncryptYamlData([]byte(plain))
	Expect(err).To(Succeed())

	return string(encrypted)
}

func generateSecretKey() []byte {
	key, err := secrets_manager.GenerateSecretKey()
	Expect(err).To(Succeed())

	return key
}

func writeFile(filePath, data string) string {
	Expect(os.MkdirAll(filepath.Dir(filePath), 0o755)).To(Succeed())
	Expect(os.WriteFile(filePath, []byte(data), 0o644)).To(Succeed())

	return filePath
}
