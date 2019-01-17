package deploy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/flant/werf/pkg/deploy/secret"
	"github.com/flant/werf/pkg/werf"
	"github.com/ghodss/yaml"
	"github.com/otiai10/copy"
	uuid "github.com/satori/go.uuid"
)

const (
	ProjectHelmChartDir            = ".helm"
	ProjectDefaultSecretValuesFile = ProjectHelmChartDir + "/secret-values.yaml"
	ProjectSecretDir               = ProjectHelmChartDir + "/secret"

	WerfChartDecodedSecretDir = "decoded-secret"
	WerfChartMoreValuesDir    = "more-values"
)

type WerfChart struct {
	ChartDir  string
	Values    []string
	Set       []string
	SetString []string

	moreValuesCounter uint
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

	decodedData, err := m.ExtractYamlData(data)
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

func (chart *WerfChart) Deploy(releaseName string, namespace string, opts HelmChartOptions) error {
	return DeployHelmChart(chart.ChartDir, releaseName, namespace, HelmChartOptions{
		CommonHelmOptions: CommonHelmOptions{KubeContext: opts.KubeContext},
		Set:               append(chart.Set, opts.Set...),
		SetString:         append(chart.SetString, opts.SetString...),
		Values:            append(chart.Values, opts.Values...),
		DryRun:            opts.DryRun,
		Debug:             opts.Debug,
	})
}

func (chart *WerfChart) Render(namespace string) (string, error) {
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

	stdout, stderr, err := HelmCmd(args...)
	if err != nil {
		return "", fmt.Errorf("%s\n%s", stdout, stderr)
	}

	return stdout, nil
}

type ChartConfig struct {
	Name string `json:"name"`
}

func (chart *WerfChart) Lint() error {
	tmpLintPath := filepath.Join(werf.GetTmpDir(), fmt.Sprintf("werf-lint-%s", uuid.NewV4().String()))
	defer os.RemoveAll(tmpLintPath)

	err := os.MkdirAll(tmpLintPath, os.ModePerm)
	if err != nil {
		return err
	}

	chartConfigPath := filepath.Join(chart.ChartDir, "Chart.yaml")

	data, err := ioutil.ReadFile(chartConfigPath)
	if err != nil {
		return fmt.Errorf("error reading %s: %s", chartConfigPath, err)
	}

	if debug() {
		fmt.Printf("Read chart config:\n%s\n", data)
	}

	var cc ChartConfig
	err = yaml.Unmarshal(data, &cc)
	if err != nil {
		return fmt.Errorf("bad chart config %s: %s", chartConfigPath, err)
	}

	tmpChartDir := filepath.Join(tmpLintPath, cc.Name)
	err = copy.Copy(chart.ChartDir, tmpChartDir)

	args := []string{"lint", tmpChartDir, "--strict"}
	for _, set := range chart.Set {
		args = append(args, "--set", set)
	}
	for _, setString := range chart.SetString {
		args = append(args, "--set-string", setString)
	}
	for _, values := range chart.Values {
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

func GenerateWerfChart(projectDir string, m secret.Manager) (*WerfChart, error) {
	tmpChartPath := filepath.Join(werf.GetTmpDir(), fmt.Sprintf("werf-chart-%s", uuid.NewV4().String()))
	return PrepareWerfChart(projectDir, tmpChartPath, m)
}

func PrepareWerfChart(projectDir string, targetDir string, m secret.Manager) (*WerfChart, error) {
	werfChart := &WerfChart{ChartDir: targetDir}

	projectHelmDir := filepath.Join(projectDir, ".helm")
	err := copy.Copy(projectHelmDir, targetDir)
	if err != nil {
		return nil, fmt.Errorf("unable to copy project helm dir %s into %s: %s", projectHelmDir, targetDir, err)
	}

	templatesDir := filepath.Join(targetDir, "templates")
	err = os.MkdirAll(templatesDir, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("unable to create %s: %s", templatesDir, err)
	}

	helpersTplPath := filepath.Join(templatesDir, "_werf_helpers.tpl")
	f, err := os.Create(helpersTplPath)
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

			decodedData, err := m.Extract([]byte(strings.TrimRightFunc(string(data), unicode.IsSpace)))
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
