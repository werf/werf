package werf_chart

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/ghodss/yaml"
	"github.com/otiai10/copy"

	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/deploy/helm"
	"github.com/flant/werf/pkg/deploy/secret"
	"github.com/flant/werf/pkg/werf"
)

const (
	DefaultSecretValuesFile = "secret-values.yaml"
	SecretDir               = "secret"

	WerfChartMoreValuesDir    = "werf.values"
	WerfChartDecodedSecretDir = "werf.secret"
)

var (
	ProjectHelmChartDir            = ".helm"
	ProjectDefaultSecretValuesFile = filepath.Join(ProjectHelmChartDir, DefaultSecretValuesFile)
	ProjectSecretDir               = filepath.Join(ProjectHelmChartDir, SecretDir)
)

func LoadWerfChart(werfChartDir string) (*WerfChart, error) {
	path := filepath.Join(werfChartDir, "werf-chart.yaml")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &WerfChart{ChartDir: werfChartDir}, nil
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read file %s: %s", path, err)
	}

	werfChart := &WerfChart{}
	err = yaml.Unmarshal(data, werfChart)
	if err != nil {
		return nil, fmt.Errorf("bad yaml %s: %s", path, err)
	}

	return werfChart, nil
}

type WerfChart struct {
	Name             string            `yaml:"Name"`
	ChartDir         string            `yaml:"ChartDir"`
	Values           []string          `yaml:"Values"`
	Set              []string          `yaml:"Set"`
	SetString        []string          `yaml:"SetString"`
	ExtraAnnotations map[string]string `yaml:"ExtraAnnotations"`
	ExtraLabels      map[string]string `yaml:"ExtraLabels"`

	moreValuesCounter uint `yaml:"moreValuesCounter"`
}

func (chart *WerfChart) Save() error {
	path := filepath.Join(chart.ChartDir, "werf-chart.yaml")

	data, err := yaml.Marshal(chart)
	if err != nil {
		return fmt.Errorf("cannot marshal werf chart %#v to yaml: %s", chart, err)
	}

	err = ioutil.WriteFile(path, data, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error writing %s: %s", path, err)
	}

	return nil
}

func (chart *WerfChart) SetGlobalAnnotation(name, value string) error {
	// TODO: https://github.com/flant/werf/issues/1069
	return nil
}

func (chart *WerfChart) SetValues(values map[string]interface{}) error {
	path := filepath.Join(chart.ChartDir, WerfChartMoreValuesDir, fmt.Sprintf("%d.yaml", chart.moreValuesCounter))
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(values)
	if err != nil {
		return fmt.Errorf("cannot marshal values %#v to yaml: %s", values, err)
	}

	err = ioutil.WriteFile(path, data, 0400)
	if err != nil {
		return fmt.Errorf("error writing values file %s: %s", path, data)
	}

	chart.Values = append(chart.Values, path)
	chart.moreValuesCounter++

	return nil
}

func (chart *WerfChart) SetValuesSet(set string) error {
	chart.Set = append(chart.Set, set)
	return nil
}

func (chart *WerfChart) SetValuesSetString(setString string) error {
	chart.SetString = append(chart.SetString, setString)
	return nil
}

func (chart *WerfChart) SetValuesFile(path string) error {
	chart.Values = append(chart.Values, path)
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

	newPath := filepath.Join(chart.ChartDir, WerfChartMoreValuesDir, fmt.Sprintf("%d.yaml", chart.moreValuesCounter))

	err = os.MkdirAll(filepath.Dir(newPath), os.ModePerm)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(newPath, decodedData, 0400)
	if err != nil {
		return fmt.Errorf("cannot write decoded secret values file %s: %s", newPath, err)
	}

	chart.Values = append(chart.Values, newPath)
	chart.moreValuesCounter++

	return nil
}

func (chart *WerfChart) Deploy(releaseName string, namespace string, opts helm.ChartOptions) error {
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

func CreateNewWerfChart(projectName, projectDir string, targetDir, env string, m secret.Manager) (*WerfChart, error) {
	werfChart := &WerfChart{}
	werfChart.ChartDir = targetDir
	werfChart.ExtraAnnotations = map[string]string{
		"werf.io/version":      werf.Version,
		"project.werf.io/name": projectName,
	}

	if env != "" {
		werfChart.ExtraAnnotations["project.werf.io/env"] = env
	}

	werfChart.ExtraLabels = map[string]string{}

	projectHelmDir := filepath.Join(projectDir, ".helm")
	err := copy.Copy(projectHelmDir, targetDir)
	if err != nil {
		return nil, fmt.Errorf("unable to copy project helm dir %s into %s: %s", projectHelmDir, targetDir, err)
	}

	werfChart.Name = projectName

	chartFile := filepath.Join(projectHelmDir, "Chart.yaml")
	if _, err := os.Stat(chartFile); !os.IsNotExist(err) {
		logboek.LogErrorF("WARNING: %s will be generated by werf automatically! To skip the warning please delete .helm/Chart.yaml from project.\n", chartFile)
	}

	targetChartFile := filepath.Join(targetDir, "Chart.yaml")
	f, err := os.Create(targetChartFile)
	if err != nil {
		return nil, fmt.Errorf("unable to create %s: %s", targetChartFile, err)
	}

	chartData := fmt.Sprintf("name: %s\nversion: 0.1.0\nengine: %s\n", werfChart.Name, helm.WerfTemplateEngineName)

	_, err = f.Write([]byte(chartData))
	if err != nil {
		return nil, fmt.Errorf("unable to write %s: %s", targetChartFile, err)
	}

	templatesDir := filepath.Join(targetDir, "templates")
	err = os.MkdirAll(templatesDir, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("unable to create %s: %s", templatesDir, err)
	}

	helpersTplPath := filepath.Join(templatesDir, "_werf_helpers.tpl")
	f, err = os.Create(helpersTplPath)
	if err != nil {
		return nil, fmt.Errorf("unable to create %s: %s", helpersTplPath, err)
	}
	_, err = f.Write(WerfChartHelpersTpl)
	if err != nil {
		return nil, fmt.Errorf("unable to write %s: %s", helpersTplPath, err)
	}

	defaultSecretValues := filepath.Join(projectDir, ProjectDefaultSecretValuesFile)
	if _, err := os.Stat(defaultSecretValues); !os.IsNotExist(err) {
		if err := werfChart.SetSecretValuesFile(defaultSecretValues, m); err != nil {
			return nil, err
		}

		if err := os.Remove(filepath.Join(targetDir, DefaultSecretValuesFile)); err != nil {
			return nil, err
		}
	}

	secretDir := filepath.Join(projectDir, ProjectSecretDir)
	if _, err := os.Stat(secretDir); !os.IsNotExist(err) {
		if err := filepath.Walk(secretDir, func(path string, info os.FileInfo, accessErr error) error {
			if accessErr != nil {
				return fmt.Errorf("error accessing file %s: %s", path, accessErr)
			}

			if info.Mode().IsDir() {
				return nil
			}

			relativePath := strings.TrimPrefix(path, secretDir)
			newPath := filepath.Join(targetDir, WerfChartDecodedSecretDir, relativePath)

			data, err := ioutil.ReadFile(path)
			if err != nil {
				return fmt.Errorf("error reading file %s: %s", path, err)
			}

			decodedData, err := m.Decrypt([]byte(strings.TrimRightFunc(string(data), unicode.IsSpace)))
			if err != nil {
				return fmt.Errorf("error decoding %s: %s", path, err)
			}

			err = os.MkdirAll(filepath.Dir(newPath), os.ModePerm)
			if err != nil {
				return err
			}
			err = ioutil.WriteFile(newPath, decodedData, 0400)
			if err != nil {
				return fmt.Errorf("error writing file %s: %s", newPath, err)
			}

			return nil
		}); err != nil {
			return nil, err
		}

		if err := os.RemoveAll(filepath.Join(targetDir, SecretDir)); err != nil {
			return nil, err
		}
	}

	return werfChart, nil
}
