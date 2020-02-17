package werf_chart

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/ghodss/yaml"

	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/deploy/helm"
	"github.com/flant/werf/pkg/deploy/secret"
	"github.com/flant/werf/pkg/util"
	"github.com/flant/werf/pkg/util/secretvalues"
	"github.com/flant/werf/pkg/werf"
)

const (
	ProjectHelmChartDirName = ".helm"

	DefaultSecretValuesFileName = "secret-values.yaml"
	SecretDirName               = "secret"
)

type WerfChart struct {
	Name             string
	ChartDir         string
	SecretValues     []map[string]interface{}
	Values           []string
	Set              []string
	SetString        []string
	ExtraAnnotations map[string]string
	ExtraLabels      map[string]string

	DecodedSecretFilesData map[string]string
	SecretValuesToMask     []string
}

func (chart *WerfChart) SetGlobalAnnotation(name, value string) error {
	// TODO: https://github.com/flant/werf/issues/1069
	return nil
}

func (chart *WerfChart) SetServiceValues(values map[string]interface{}) error {
	chart.SecretValues = append(chart.SecretValues, values)
	return nil
}

func (chart *WerfChart) SetSecretValuesFile(path string, m secret.Manager) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read secret values file %s: %s", path, err)
	}

	decodedData, err := m.DecryptYamlData(data)
	if err != nil {
		return fmt.Errorf("cannot decode secret values file %s data: %s", path, err)
	}

	var values map[string]interface{}
	if err := yaml.UnmarshalStrict(decodedData, &values); err != nil {
		return fmt.Errorf("cannot unmarshal secret values file %s: %s", path, err)
	}
	chart.SecretValues = append(chart.SecretValues, values)
	chart.SecretValuesToMask = secretvalues.ExtractSecretValuesFromMap(values)

	return nil
}

func valuesToStrvals(values map[string]interface{}) []string {
	var result []string

	for key, value := range values {
		result = append(result, valueToStrvals(key, value)...)
	}

	return result
}

func valueToStrvals(valuePath string, value interface{}) []string {
	var result []string

	switch x := value.(type) {
	case []interface{}:
		for ind, v := range x {
			result = append(result, valueToStrvals(fmt.Sprintf("%s[%d]", valuePath, ind), v)...)
		}
	case map[string]interface{}:
		for k, v := range x {
			result = append(result, valueToStrvals(strings.Join([]string{valuePath, k}, "."), v)...)
		}
	default:
		result = append(result, fmt.Sprintf("%s=%v", valuePath, x))
	}

	return result
}

func (chart *WerfChart) Deploy(releaseName string, namespace string, opts helm.ChartOptions) error {
	opts.SecretValues = append(chart.SecretValues, opts.SecretValues...)
	opts.Set = append(chart.Set, opts.Set...)
	opts.SetString = append(chart.SetString, opts.SetString...)
	opts.Values = append(chart.Values, opts.Values...)

	return helm.DeployHelmChart(chart.ChartDir, releaseName, namespace, opts)
}

func (chart *WerfChart) MergeExtraAnnotations(extraAnnotations map[string]string) {
	for annoName, annoValue := range extraAnnotations {
		chart.ExtraAnnotations[annoName] = annoValue
	}
}

func (chart *WerfChart) MergeExtraLabels(extraLabels map[string]string) {
	for labelName, labelValue := range extraLabels {
		chart.ExtraLabels[labelName] = labelValue
	}
}

func (chart *WerfChart) LogExtraAnnotations() {
	if len(chart.ExtraAnnotations) == 0 {
		return
	}

	res, _ := yaml.Marshal(chart.ExtraAnnotations)

	annotations := strings.TrimRight(string(res), "\n")
	logboek.LogLn("Using extra annotations:")
	logboek.LogF(logboek.FitText(annotations, logboek.FitTextOptions{ExtraIndentWidth: 2}))
	logboek.LogLn()
	logboek.LogOptionalLn()
}

func (chart *WerfChart) LogExtraLabels() {
	if len(chart.ExtraLabels) == 0 {
		return
	}

	res, _ := yaml.Marshal(chart.ExtraLabels)

	labels := strings.TrimRight(string(res), "\n")
	logboek.LogLn("Using extra labels:")
	logboek.LogF(logboek.FitText(labels, logboek.FitTextOptions{ExtraIndentWidth: 2}))
	logboek.LogLn()
	logboek.LogOptionalLn()
}

type ChartConfig struct {
	Name string `json:"name"`
}

func InitWerfChart(projectName, chartDir string, env string, m secret.Manager) (*WerfChart, error) {
	werfChart := &WerfChart{}
	werfChart.Name = projectName
	werfChart.ChartDir = chartDir
	werfChart.ExtraAnnotations = map[string]string{
		"werf.io/version":      werf.Version,
		"project.werf.io/name": projectName,
	}
	werfChart.DecodedSecretFilesData = make(map[string]string, 0)

	if env != "" {
		werfChart.ExtraAnnotations["project.werf.io/env"] = env
	}

	werfChart.ExtraLabels = map[string]string{}

	chartYamlFile := filepath.Join(chartDir, "Chart.yaml")
	if exist, err := util.FileExists(chartYamlFile); err != nil {
		return nil, fmt.Errorf("check file %s existence failed: %s", chartYamlFile, err)
	} else if exist {
		logboek.LogErrorF("WARNING: werf generates Chart metadata based on project werf.yaml! To skip the warning please delete .helm/Chart.yaml.\n")
	}

	defaultSecretValues := filepath.Join(chartDir, DefaultSecretValuesFileName)
	if _, err := os.Stat(defaultSecretValues); !os.IsNotExist(err) {
		if err := werfChart.SetSecretValuesFile(defaultSecretValues, m); err != nil {
			return nil, err
		}
	}

	secretDir := filepath.Join(chartDir, SecretDirName)
	if _, err := os.Stat(secretDir); !os.IsNotExist(err) {
		if err := filepath.Walk(secretDir, func(path string, info os.FileInfo, accessErr error) error {
			if accessErr != nil {
				return fmt.Errorf("error accessing file %s: %s", path, accessErr)
			}

			if info.Mode().IsDir() {
				return nil
			}

			data, err := ioutil.ReadFile(path)
			if err != nil {
				return fmt.Errorf("error reading file %s: %s", path, err)
			}

			decodedData, err := m.Decrypt([]byte(strings.TrimRightFunc(string(data), unicode.IsSpace)))
			if err != nil {
				return fmt.Errorf("error decoding %s: %s", path, err)
			}

			relativePath, err := filepath.Rel(secretDir, path)
			if err != nil {
				panic(err)
			}

			werfChart.DecodedSecretFilesData[filepath.ToSlash(relativePath)] = string(decodedData)
			werfChart.SecretValuesToMask = append(werfChart.SecretValuesToMask, string(decodedData))

			return nil
		}); err != nil {
			return nil, err
		}
	}

	return werfChart, nil
}
