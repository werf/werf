package publish

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/logboek"
	chartcommon "github.com/werf/nelm/v2/pkg/helm/pkg/chart/common"
	chartcommonutil "github.com/werf/nelm/v2/pkg/helm/pkg/chart/common/util"
	chart "github.com/werf/nelm/v2/pkg/helm/pkg/chart/v2"
	"github.com/werf/nelm/v2/pkg/helm/pkg/cli/values"
	"github.com/werf/nelm/v2/pkg/helm/pkg/getter"
)

type bundleValuesCase struct {
	chartValues      string
	chartSecretVals  string
	customSecretVals string
	valuesFile       string
	set              []string
	setJSON          []string
	serviceVals      string

	expectedSecretVals string
	expectedVals       string
}

var _ = Describe("makeBundleValuesAndSecretValues", func() {
	DescribeTable("removes the secret values overridden at publish time and applies like converge",
		func(c bundleValuesCase) {
			ctx := context.Background()

			valsOpts := &values.Options{Values: c.set, JSONValues: c.setJSON}
			if c.valuesFile != "" {
				valuesFilePath := filepath.Join(GinkgoT().TempDir(), "values.yaml")
				Expect(os.WriteFile(valuesFilePath, []byte(c.valuesFile), 0o644)).To(Succeed())
				valsOpts.ValueFiles = []string{valuesFilePath}
			}

			publishChart := &chart.Chart{
				Metadata: &chart.Metadata{Name: "root", APIVersion: chart.APIVersionV2, Version: "1.0.0"},
				Values:   parseValues(c.chartValues),
			}

			bundleVals, bundleSecretVals, err := makeBundleValuesAndSecretValues(ctx, publishChart, valsOpts, parseValues(c.serviceVals), parseValues(c.chartSecretVals), parseValues(c.customSecretVals))
			Expect(err).To(Succeed())
			Expect(bundleSecretVals).To(Equal(parseValues(c.expectedSecretVals)))

			bundleChart := &chart.Chart{
				Metadata:           &chart.Metadata{Name: "root", APIVersion: chart.APIVersionV2, Version: "1.0.0"},
				Values:             bundleVals,
				SecretsRuntimeData: &fakeSecretsRuntimeData{defaultSecretValues: bundleSecretVals},
			}

			appliedVals, err := chartcommonutil.CoalesceValues(bundleChart, parseValues(c.serviceVals))
			Expect(err).To(Succeed())
			Expect(map[string]interface{}(appliedVals)).To(Equal(parseValues(c.expectedVals)))

			convergedVals := convergeValues(ctx, c)
			Expect(map[string]interface{}(appliedVals)).To(Equal(convergedVals))
		},
		Entry("--set overrides the chart secret values and --secret-values", bundleValuesCase{
			chartValues:      "a: chart",
			chartSecretVals:  "a: chart-secret\nkept: chart-secret",
			customSecretVals: "a: custom-secret",
			set:              []string{"a=set"},

			expectedSecretVals: "kept: chart-secret",
			expectedVals:       "a: set\nkept: chart-secret",
		}),
		Entry("--values override the chart secret values only", bundleValuesCase{
			chartSecretVals:  "a: chart-secret\nb: chart-secret",
			customSecretVals: "a: custom-secret",
			valuesFile:       "a: values\nb: values",

			expectedSecretVals: "a: custom-secret",
			expectedVals:       "a: custom-secret\nb: values",
		}),
		Entry("a scalar replaces a whole secret map", bundleValuesCase{
			chartSecretVals:  "a:\n  b: chart-secret\n  c: chart-secret",
			customSecretVals: "s:\n  b: custom-secret",
			set:              []string{"a=set", "s=set"},

			expectedSecretVals: "{}",
			expectedVals:       "a: set\ns: set",
		}),
		Entry("a map replaces a secret scalar", bundleValuesCase{
			chartSecretVals: "a: chart-secret",
			valuesFile:      "a:\n  b: values",

			expectedSecretVals: "{}",
			expectedVals:       "a:\n  b: values",
		}),
		Entry("only the overridden keys of a secret map are removed", bundleValuesCase{
			chartSecretVals:  "a:\n  b: chart-secret\n  c: chart-secret",
			customSecretVals: "a:\n  d: custom-secret\n  e: custom-secret",
			set:              []string{"a.b=set", "a.d=set"},

			expectedSecretVals: "a:\n  c: chart-secret\n  e: custom-secret",
			expectedVals:       "a:\n  b: set\n  c: chart-secret\n  d: set\n  e: custom-secret",
		}),
		Entry("a secret map without keys left is removed", bundleValuesCase{
			chartSecretVals: "a:\n  b: chart-secret",
			set:             []string{"a.b=set"},

			expectedSecretVals: "{}",
			expectedVals:       "a:\n  b: set",
		}),
		Entry("a list from --set replaces a secret list", bundleValuesCase{
			chartSecretVals: "l:\n- chart-secret",
			set:             []string{"l[0]=set"},

			expectedSecretVals: "{}",
			expectedVals:       "l:\n- set",
		}),
		Entry("--set null deletes the secret values and the chart values", bundleValuesCase{
			chartValues:      "a: chart",
			chartSecretVals:  "a: chart-secret",
			customSecretVals: "a: custom-secret",
			set:              []string{"a=null"},

			expectedSecretVals: "{}",
			expectedVals:       "{}",
		}),
		Entry("--values null deletes the chart secret values and the chart values", bundleValuesCase{
			chartValues:     "a: chart",
			chartSecretVals: "a: chart-secret",
			valuesFile:      "a: null",

			expectedSecretVals: "{}",
			expectedVals:       "{}",
		}),
		Entry("--secret-values null still deletes the lower values", bundleValuesCase{
			chartValues:      "a: chart\nb: chart",
			chartSecretVals:  "a: chart-secret",
			customSecretVals: "a: null\nb: null",
			valuesFile:       "b: values",

			expectedSecretVals: "a: null\nb: null",
			expectedVals:       "{}",
		}),
		Entry("a chart secret null is kept unless overridden", bundleValuesCase{
			chartValues:     "a: chart\nb: chart",
			chartSecretVals: "a: null\nb: null",
			set:             []string{"b=set"},

			expectedSecretVals: "a: null",
			expectedVals:       "b: set",
		}),
		Entry("service values override the chart secret values and --secret-values", bundleValuesCase{
			chartSecretVals:  "werf:\n  env: chart-secret\nkept: chart-secret",
			customSecretVals: "global:\n  werf:\n    env: custom-secret",
			serviceVals:      "werf:\n  env: service\nglobal:\n  werf:\n    env: service",

			expectedSecretVals: "kept: chart-secret",
			expectedVals:       "werf:\n  env: service\nglobal:\n  werf:\n    env: service\nkept: chart-secret",
		}),
		Entry("--set-json replaces a --secret-values map", bundleValuesCase{
			chartValues:      "m:\n  d: chart",
			chartSecretVals:  "m:\n  c: chart-secret\nkept: chart-secret",
			customSecretVals: "m:\n  k: custom-secret\n  j: custom-secret",
			valuesFile:       "m:\n  c: values",
			setJSON:          []string{`m={"j":"set-json"}`},

			expectedSecretVals: "m:\n  c: chart-secret\nkept: chart-secret",
			expectedVals:       "m:\n  c: chart-secret\n  d: chart\n  j: set-json\nkept: chart-secret",
		}),
		Entry("--set-json replaces a nested --secret-values map", bundleValuesCase{
			customSecretVals: "m:\n  inner:\n    k: custom-secret\n  o: custom-secret",
			setJSON:          []string{`m.inner={"j":"set-json"}`},

			expectedSecretVals: "m:\n  o: custom-secret",
			expectedVals:       "m:\n  inner:\n    j: set-json\n  o: custom-secret",
		}),
		Entry("--set-json merges into a secret map", bundleValuesCase{
			customSecretVals: "m:\n  k: custom-secret\n  j: custom-secret",
			setJSON:          []string{`m.j="set-json"`},

			expectedSecretVals: "m:\n  k: custom-secret",
			expectedVals:       "m:\n  j: set-json\n  k: custom-secret",
		}),
		Entry("secret values override the chart values when nothing overrides them", bundleValuesCase{
			chartValues:      "a: chart\nb: chart\nempty: {}",
			chartSecretVals:  "a: chart-secret\nempty: {}",
			customSecretVals: "b: custom-secret",
			set:              []string{"empty.x=set"},

			expectedSecretVals: "a: chart-secret\nb: custom-secret\nempty: {}",
			expectedVals:       "a: chart-secret\nb: custom-secret\nempty:\n  x: set",
		}),
		Entry("--values overridden by a --secret-values map do not remove the chart secret values", bundleValuesCase{
			chartValues:      "a:\n  d: chart",
			chartSecretVals:  "a:\n  x: chart-secret",
			customSecretVals: "a:\n  y: custom-secret",
			valuesFile:       "a: values",

			expectedSecretVals: "a:\n  x: chart-secret\n  y: custom-secret",
			expectedVals:       "a:\n  d: chart\n  x: chart-secret\n  y: custom-secret",
		}),
		Entry("--values null overridden by a --secret-values map keeps the lower values", bundleValuesCase{
			chartValues:      "a:\n  d: chart",
			chartSecretVals:  "a:\n  b: chart-secret\n  c: chart-secret",
			customSecretVals: "a:\n  b: custom-secret",
			valuesFile:       "a: null",

			expectedSecretVals: "a:\n  b: custom-secret\n  c: chart-secret",
			expectedVals:       "a:\n  b: custom-secret\n  c: chart-secret\n  d: chart",
		}),
		Entry("--set over a map that only --secret-values have", bundleValuesCase{
			customSecretVals: "a:\n  k: custom-secret",
			valuesFile:       "a: values",
			set:              []string{"a.b=set"},

			expectedSecretVals: "a:\n  k: custom-secret",
			expectedVals:       "a:\n  b: set\n  k: custom-secret",
		}),
		Entry("--values next to service values do not override --secret-values", bundleValuesCase{
			customSecretVals: "global:\n  db:\n    password: custom-secret",
			valuesFile:       "global:\n  db:\n    password: values\n    user: values",
			serviceVals:      "global:\n  werf:\n    env: service",

			expectedSecretVals: "global:\n  db:\n    password: custom-secret",
			expectedVals:       "global:\n  db:\n    password: custom-secret\n    user: values\n  werf:\n    env: service",
		}),
	)

	DescribeTable("fails like converge when --set* cannot be applied over --secret-values",
		func(valuesFile, customSecretVals string) {
			valsOpts := &values.Options{Values: []string{"a.y=set"}}
			if valuesFile != "" {
				valuesFilePath := filepath.Join(GinkgoT().TempDir(), "values.yaml")
				Expect(os.WriteFile(valuesFilePath, []byte(valuesFile), 0o644)).To(Succeed())
				valsOpts.ValueFiles = []string{valuesFilePath}
			}

			publishChart := &chart.Chart{Metadata: &chart.Metadata{Name: "root", APIVersion: chart.APIVersionV2, Version: "1.0.0"}}

			_, _, err := makeBundleValuesAndSecretValues(context.Background(), publishChart, valsOpts, nil, nil, parseValues(customSecretVals))
			Expect(err).To(MatchError(ContainSubstring("failed parsing --set data")))
		},
		Entry("a --secret-values scalar", "", "a: custom-secret"),
		Entry("a --secret-values scalar over a --values map", "a:\n  x: values", "a: custom-secret"),
	)

	DescribeTable("removes the secret values overridden by exported values",
		func(c bundleValuesCase, alias string) {
			valsOpts := &values.Options{Values: c.set, JSONValues: c.setJSON}
			if c.valuesFile != "" {
				valuesFilePath := filepath.Join(GinkgoT().TempDir(), "values.yaml")
				Expect(os.WriteFile(valuesFilePath, []byte(c.valuesFile), 0o644)).To(Succeed())
				valsOpts.ValueFiles = []string{valuesFilePath}
			}

			publishChart := &chart.Chart{
				Metadata: &chart.Metadata{
					Name:       "root",
					APIVersion: chart.APIVersionV2,
					Version:    "1.0.0",
					Dependencies: []*chart.Dependency{{
						Name:         "sub",
						Alias:        alias,
						Version:      "1.0.0",
						ExportValues: []interface{}{map[string]interface{}{"parent": "g.repo", "child": "image.repository"}},
					}},
				},
			}
			publishChart.AddDependency(&chart.Chart{
				Metadata: &chart.Metadata{Name: "sub", APIVersion: chart.APIVersionV2, Version: "1.0.0"},
			})

			_, bundleSecretVals, err := makeBundleValuesAndSecretValues(context.Background(), publishChart, valsOpts, parseValues(c.serviceVals), parseValues(c.chartSecretVals), parseValues(c.customSecretVals))
			Expect(err).To(Succeed())
			Expect(bundleSecretVals).To(Equal(parseValues(c.expectedSecretVals)))
		},
		Entry("exported from --set", bundleValuesCase{
			chartSecretVals:  "sub:\n  image:\n    repository: chart-secret\n    tag: chart-secret",
			customSecretVals: "sub:\n  image:\n    repository: custom-secret",
			set:              []string{"g.repo=set"},

			expectedSecretVals: "sub:\n  image:\n    tag: chart-secret",
		}, ""),
		Entry("exported from --values", bundleValuesCase{
			chartSecretVals:  "sub:\n  image:\n    repository: chart-secret",
			customSecretVals: "sub:\n  image:\n    tag: custom-secret",
			valuesFile:       "g:\n  repo: values",

			expectedSecretVals: "sub:\n  image:\n    tag: custom-secret",
		}, ""),
		Entry("exported from service values", bundleValuesCase{
			chartSecretVals:  "sub:\n  image:\n    repository: chart-secret",
			customSecretVals: "sub:\n  image:\n    repository: custom-secret",
			serviceVals:      "g:\n  repo: service",

			expectedSecretVals: "{}",
		}, ""),
		Entry("exported to an aliased dependency", bundleValuesCase{
			chartSecretVals: "app:\n  image:\n    repository: chart-secret\nsub:\n  image:\n    repository: chart-secret",
			set:             []string{"g.repo=set"},

			expectedSecretVals: "sub:\n  image:\n    repository: chart-secret",
		}, "app"),
		Entry("not exported", bundleValuesCase{
			chartSecretVals: "sub:\n  image:\n    repository: chart-secret",
			set:             []string{"other=set"},

			expectedSecretVals: "sub:\n  image:\n    repository: chart-secret",
		}, ""),
		Entry("exported from --values next to service values", bundleValuesCase{
			customSecretVals: "sub:\n  image:\n    repository: custom-secret",
			valuesFile:       "g:\n  repo: values",
			serviceVals:      "g:\n  werf: service",

			expectedSecretVals: "sub:\n  image:\n    repository: custom-secret",
		}, ""),
		Entry("exported from --values replaced by --set-json", bundleValuesCase{
			chartSecretVals: "sub:\n  image:\n    repository: chart-secret",
			valuesFile:      "g:\n  repo: values",
			setJSON:         []string{`g={}`},

			expectedSecretVals: "sub:\n  image:\n    repository: chart-secret",
		}, ""),
	)

	DescribeTable("warns once about the export-values skipped because of the publish-time values",
		func(c bundleValuesCase, expectedWarnings int) {
			valsOpts := &values.Options{Values: c.set, JSONValues: c.setJSON}
			if c.valuesFile != "" {
				valuesFilePath := filepath.Join(GinkgoT().TempDir(), "values.yaml")
				Expect(os.WriteFile(valuesFilePath, []byte(c.valuesFile), 0o644)).To(Succeed())
				valsOpts.ValueFiles = []string{valuesFilePath}
			}

			publishChart := &chart.Chart{
				Metadata: &chart.Metadata{
					Name:       "root",
					APIVersion: chart.APIVersionV2,
					Version:    "1.0.0",
					Dependencies: []*chart.Dependency{{
						Name:         "sub",
						Version:      "1.0.0",
						ExportValues: []interface{}{map[string]interface{}{"parent": "g", "child": "."}},
					}},
				},
			}
			publishChart.AddDependency(&chart.Chart{
				Metadata: &chart.Metadata{Name: "sub", APIVersion: chart.APIVersionV2, Version: "1.0.0"},
			})

			var output bytes.Buffer
			ctx := logboek.NewContext(context.Background(), logboek.NewLogger(&output, &output))

			bundleVals, bundleSecretVals, err := makeBundleValuesAndSecretValues(ctx, publishChart, valsOpts, parseValues(c.serviceVals), nil, parseValues(c.customSecretVals))
			Expect(err).To(Succeed())
			Expect(bundleVals).To(Equal(parseValues(c.expectedVals)))
			Expect(bundleSecretVals).To(Equal(parseValues(c.expectedSecretVals)))
			Expect(strings.Count(output.String(), `Skipping export-values of "g"`)).To(Equal(expectedWarnings), output.String())
		},
		Entry("a scalar from --set", bundleValuesCase{
			set: []string{"g=scalar"},

			expectedSecretVals: "{}",
			expectedVals:       "g: scalar",
		}, 1),
		Entry("a scalar from --values", bundleValuesCase{
			valuesFile: "g: scalar",

			expectedSecretVals: "{}",
			expectedVals:       "g: scalar",
		}, 1),
		Entry("a scalar from --secret-values", bundleValuesCase{
			customSecretVals: "g: scalar",

			expectedSecretVals: "g: scalar",
			expectedVals:       "{}",
		}, 1),
		Entry("a scalar from service values", bundleValuesCase{
			serviceVals: "g: scalar",

			expectedSecretVals: "{}",
			expectedVals:       "g: scalar",
		}, 1),
		Entry("a scalar from every publish-time layer", bundleValuesCase{
			customSecretVals: "g: scalar",
			valuesFile:       "g: scalar",
			set:              []string{"g=scalar"},
			serviceVals:      "g: scalar",

			expectedSecretVals: "{}",
			expectedVals:       "g: scalar",
		}, 1),
		Entry("a map", bundleValuesCase{
			set: []string{"g.key=value"},

			expectedSecretVals: "{}",
			expectedVals:       "g:\n  key: value",
		}, 0),
		Entry("a scalar from --values replaced by --set-json", bundleValuesCase{
			valuesFile: "g: scalar",
			setJSON:    []string{`g={"key":"value"}`},

			expectedSecretVals: "{}",
			expectedVals:       "g:\n  key: value",
		}, 0),
		Entry("a scalar from --secret-values replaced by --set-json", bundleValuesCase{
			customSecretVals: "g: scalar",
			setJSON:          []string{`g={"key":"value"}`},

			expectedSecretVals: "{}",
			expectedVals:       "g:\n  key: value",
		}, 0),
	)

	It("accepts --set* that are valid only over the not exported --values", func() {
		publishChart := &chart.Chart{
			Metadata: &chart.Metadata{
				Name:       "root",
				APIVersion: chart.APIVersionV2,
				Version:    "1.0.0",
				Dependencies: []*chart.Dependency{{
					Name:         "sub",
					Version:      "1.0.0",
					ExportValues: []interface{}{map[string]interface{}{"parent": "g.config", "child": "config"}},
				}},
			},
		}
		publishChart.AddDependency(&chart.Chart{
			Metadata: &chart.Metadata{Name: "sub", APIVersion: chart.APIVersionV2, Version: "1.0.0"},
		})

		valuesFilePath := filepath.Join(GinkgoT().TempDir(), "values.yaml")
		Expect(os.WriteFile(valuesFilePath, []byte("g:\n  config: values\nsub:\n  config:\n    other: values"), 0o644)).To(Succeed())

		bundleVals, _, err := makeBundleValuesAndSecretValues(context.Background(), publishChart, &values.Options{ValueFiles: []string{valuesFilePath}, Values: []string{"sub.config.key=set"}}, nil, nil, nil)
		Expect(err).To(Succeed())
		Expect(bundleVals).To(Equal(parseValues("g:\n  config: values\nsub:\n  config:\n    key: set\n    other: values")))
	})

	It("merges --set* that are valid only over the exported --values over --values", func() {
		publishChart := &chart.Chart{
			Metadata: &chart.Metadata{
				Name:       "root",
				APIVersion: chart.APIVersionV2,
				Version:    "1.0.0",
				Dependencies: []*chart.Dependency{{
					Name:         "sub",
					Version:      "1.0.0",
					ExportValues: []interface{}{map[string]interface{}{"parent": "g.config", "child": "config"}},
				}},
			},
		}
		publishChart.AddDependency(&chart.Chart{
			Metadata: &chart.Metadata{Name: "sub", APIVersion: chart.APIVersionV2, Version: "1.0.0"},
		})

		valuesFilePath := filepath.Join(GinkgoT().TempDir(), "values.yaml")
		Expect(os.WriteFile(valuesFilePath, []byte("g:\n  config:\n    x: values\nsub:\n  config: values"), 0o644)).To(Succeed())

		bundleVals, _, err := makeBundleValuesAndSecretValues(context.Background(), publishChart, &values.Options{ValueFiles: []string{valuesFilePath}, Values: []string{"sub.config.key=set"}}, nil, nil, nil)
		Expect(err).To(Succeed())
		Expect(bundleVals).To(Equal(parseValues("g:\n  config:\n    x: values\nsub:\n  config:\n    key: set")))
	})

	It("accepts --set* that are valid only over the exported --secret-values", func() {
		publishChart := &chart.Chart{
			Metadata: &chart.Metadata{
				Name:       "root",
				APIVersion: chart.APIVersionV2,
				Version:    "1.0.0",
				Dependencies: []*chart.Dependency{{
					Name:         "sub",
					Version:      "1.0.0",
					ExportValues: []interface{}{map[string]interface{}{"parent": "g.config", "child": "config"}},
				}},
			},
		}
		publishChart.AddDependency(&chart.Chart{
			Metadata: &chart.Metadata{Name: "sub", APIVersion: chart.APIVersionV2, Version: "1.0.0"},
		})

		bundleVals, bundleSecretVals, err := makeBundleValuesAndSecretValues(context.Background(), publishChart, &values.Options{Values: []string{"sub.config.key=set"}}, nil, nil, parseValues("g:\n  config:\n    x: custom-secret\nsub:\n  config: custom-secret"))
		Expect(err).To(Succeed())
		Expect(bundleVals).To(Equal(parseValues("sub:\n  config:\n    key: set")))
		Expect(bundleSecretVals).To(Equal(parseValues("g:\n  config:\n    x: custom-secret")))
	})

	It("leaves the dependency values in the dependency charts", func() {
		publishChart := &chart.Chart{
			Metadata: &chart.Metadata{Name: "root", APIVersion: chart.APIVersionV2, Version: "1.0.0"},
			Values:   parseValues("root: chart\nchild:\n  fromparent: chart"),
		}
		publishChart.AddDependency(&chart.Chart{
			Metadata: &chart.Metadata{Name: "child", APIVersion: chart.APIVersionV2, Version: "1.0.0"},
			Values:   parseValues("password: child"),
		})

		bundleVals, _, err := makeBundleValuesAndSecretValues(context.Background(), publishChart, &values.Options{Values: []string{"child.fromset=set"}}, nil, nil, nil)
		Expect(err).To(Succeed())
		Expect(bundleVals).To(Equal(parseValues("root: chart\nchild:\n  fromparent: chart\n  fromset: set")))
	})
})

func convergeValues(ctx context.Context, c bundleValuesCase) map[string]interface{} {
	valsOpts := &values.Options{Values: c.set, JSONValues: c.setJSON}
	fileVals := map[string]interface{}{}
	if c.valuesFile != "" {
		fileVals = parseValues(c.valuesFile)
	}

	withSecrets, err := valsOpts.MergeSetValues(ctx, getter.Getters(), chartcommon.MergeMaps(fileVals, parseValues(c.customSecretVals)), values.MergeSetValuesOptions{SetFilesCache: map[string]string{}})
	Expect(err).To(Succeed())

	convergeChart := &chart.Chart{
		Metadata:           &chart.Metadata{Name: "root", APIVersion: chart.APIVersionV2, Version: "1.0.0"},
		Values:             parseValues(c.chartValues),
		SecretsRuntimeData: &fakeSecretsRuntimeData{defaultSecretValues: parseValues(c.chartSecretVals)},
	}

	convergedVals, err := chartcommonutil.CoalesceValues(convergeChart, chartcommon.MergeMaps(withSecrets, parseValues(c.serviceVals)))
	Expect(err).To(Succeed())

	return convergedVals
}
