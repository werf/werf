package werf_chart

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/deploy/helm"
	"github.com/flant/werf/pkg/deploy/secret"
	"github.com/ghodss/yaml"
	"github.com/otiai10/copy"
)

const (
	ProjectHelmChartDir            = ".helm"
	ProjectDefaultSecretValuesFile = ProjectHelmChartDir + "/secret-values.yaml"
	ProjectSecretDir               = ProjectHelmChartDir + "/secret"

	WerfChartDecodedSecretDir = "decoded-secret"
	WerfChartMoreValuesDir    = "more-values"
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
	Name      string   `yaml:"Name"`
	ChartDir  string   `yaml:"ChartDir"`
	Values    []string `yaml:"Values"`
	Set       []string `yaml:"Set"`
	SetString []string `yaml:"SetString"`

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

func (chart *WerfChart) Deploy(releaseName string, namespace string, opts helm.HelmChartOptions) error {
	return helm.DeployHelmChart(chart.ChartDir, releaseName, namespace, helm.HelmChartOptions{
		HelmChartValuesOptions: helm.HelmChartValuesOptions{
			Set:       append(chart.Set, opts.Set...),
			SetString: append(chart.SetString, opts.SetString...),
			Values:    append(chart.Values, opts.Values...),
		},
		DryRun: opts.DryRun,
		Debug:  opts.Debug,
	})
}

func (chart *WerfChart) Render(namespace string, opts helm.HelmChartValuesOptions) (string, error) {
	args := []string{"template", chart.ChartDir}

	args = append(args, "--namespace", namespace)

	for _, set := range chart.Set {
		args = append(args, "--set", set)
	}
	for _, setString := range chart.SetString {
		args = append(args, "--set-string", setString)
	}
	for _, values := range chart.Values {
		args = append(args, "--values", values)
	}

	for _, set := range opts.Set {
		args = append(args, "--set", set)
	}
	for _, setString := range opts.SetString {
		args = append(args, "--set-string", setString)
	}
	for _, values := range opts.Values {
		args = append(args, "--values", values)
	}

	stdout, stderr, err := helm.HelmCmd(args...)
	if err != nil {
		return "", helm.FormatHelmCmdError(stdout, stderr, err)
	}

	return stdout, nil
}

type ChartConfig struct {
	Name string `json:"name"`
}

func (chart *WerfChart) Lint(opts helm.HelmChartValuesOptions) error {
	args := []string{"lint", chart.ChartDir} // TODO: opts.Strict

	for _, set := range chart.Set {
		args = append(args, "--set", set)
	}
	for _, setString := range chart.SetString {
		args = append(args, "--set-string", setString)
	}
	for _, values := range chart.Values {
		args = append(args, "--values", values)
	}

	for _, set := range opts.Set {
		args = append(args, "--set", set)
	}
	for _, setString := range opts.SetString {
		args = append(args, "--set-string", setString)
	}
	for _, values := range opts.Values {
		args = append(args, "--values", values)
	}

	cmd := exec.Command("helm", args...)
	cmd.Env = os.Environ()

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("helm lint failed: %s\n%s", err, output.String())
	}

	fmt.Printf("%s", output.String())

	return nil
}

func CreateNewWerfChart(projectName, projectDir string, targetDir string, m secret.Manager) (*WerfChart, error) {
	werfChart := &WerfChart{ChartDir: targetDir}

	projectHelmDir := filepath.Join(projectDir, ".helm")
	err := copy.Copy(projectHelmDir, targetDir)
	if err != nil {
		return nil, fmt.Errorf("unable to copy project helm dir %s into %s: %s", projectHelmDir, targetDir, err)
	}

	werfChart.Name = projectName

	chartFile := filepath.Join(projectHelmDir, "Chart.yaml")
	if _, err := os.Stat(chartFile); !os.IsNotExist(err) {
		logboek.LogErrorLn("WARNING: To skip the warning please delete .helm/Chart.yaml from project")
	}

	targetChartFile := filepath.Join(targetDir, "Chart.yaml")
	f, err := os.Create(targetChartFile)
	if err != nil {
		return nil, fmt.Errorf("unable to create %s: %s", targetChartFile, err)
	}

	chartData := fmt.Sprintf("name: %s\nversion: 0.1.0\n", werfChart.Name)

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
		err := werfChart.SetSecretValuesFile(defaultSecretValues, m)
		if err != nil {
			return nil, err
		}
	}

	secretDir := filepath.Join(projectDir, ProjectSecretDir)
	if _, err := os.Stat(secretDir); !os.IsNotExist(err) {
		err := filepath.Walk(secretDir, func(path string, info os.FileInfo, accessErr error) error {
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
		})

		if err != nil {
			return nil, err
		}
	}

	return werfChart, nil
}
