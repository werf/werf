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

	"github.com/flant/dapp/pkg/dapp"
	"github.com/flant/dapp/pkg/deploy/secret"
	"github.com/ghodss/yaml"
	"github.com/otiai10/copy"
	"github.com/satori/go.uuid"
)

const (
	ProjectHelmChartDir            = ".helm"
	ProjectDefaultSecretValuesFile = ProjectHelmChartDir + "/secret-values.yaml"
	ProjectSecretDir               = ProjectHelmChartDir + "/secret"

	DappChartDecodedSecretDir = "decoded-secret"
	DappChartMoreValuesDir    = "more-values"
)

type DappChart struct {
	ChartDir  string
	Values    []string
	Set       []string
	SetString []string

	moreValuesCounter uint
}

func (chart *DappChart) SetGlobalAnnotation(name, value string) error {
	// TODO: https://github.com/flant/dapp/issues/1069
	return nil
}

func (chart *DappChart) SetValues(values map[string]interface{}) error {
	path := filepath.Join(chart.ChartDir, DappChartMoreValuesDir, fmt.Sprintf("%d.yaml", chart.moreValuesCounter))
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

func (chart *DappChart) SetValuesSet(set string) error {
	chart.Set = append(chart.Set, set)
	return nil
}

func (chart *DappChart) SetValuesSetString(setString string) error {
	chart.SetString = append(chart.SetString, setString)
	return nil
}

func (chart *DappChart) SetValuesFile(path string) error {
	chart.Values = append(chart.Values, path)
	return nil
}

func (chart *DappChart) SetSecretValuesFile(path string, m secret.Manager) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read secret values file %s: %s", path, err)
	}

	decodedData, err := m.ExtractYamlData(data)
	if err != nil {
		return fmt.Errorf("cannot decode secret values file %s data: %s", path, err)
	}

	newPath := filepath.Join(chart.ChartDir, DappChartMoreValuesDir, fmt.Sprintf("%d.yaml", chart.moreValuesCounter))

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

func (chart *DappChart) Deploy(releaseName string, namespace string, opts HelmChartOptions) error {
	return DeployHelmChart(chart.ChartDir, releaseName, namespace, HelmChartOptions{
		CommonHelmOptions: CommonHelmOptions{KubeContext: opts.KubeContext},
		Set:               append(chart.Set, opts.Set...),
		SetString:         append(chart.SetString, opts.SetString...),
		Values:            append(chart.Values, opts.Values...),
		DryRun:            opts.DryRun,
		Debug:             opts.Debug,
	})
}

func (chart *DappChart) Render(namespace string) (string, error) {
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

func (chart *DappChart) Lint() error {
	tmpLintPath := filepath.Join(dapp.GetTmpDir(), fmt.Sprintf("dapp-lint-%s", uuid.NewV4().String()))
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

func GenerateDappChart(projectDir string, m secret.Manager) (*DappChart, error) {
	tmpChartPath := filepath.Join(dapp.GetTmpDir(), fmt.Sprintf("dapp-chart-%s", uuid.NewV4().String()))
	return PrepareDappChart(projectDir, tmpChartPath, m)
}

func PrepareDappChart(projectDir string, targetDir string, m secret.Manager) (*DappChart, error) {
	dappChart := &DappChart{ChartDir: targetDir}

	projectHelmDir := filepath.Join(projectDir, ".helm")
	err := copy.Copy(projectHelmDir, targetDir)
	if err != nil {
		return nil, fmt.Errorf("unable to copy project helm dir %s into %s: %s", projectHelmDir, targetDir, err)
	}

	helpersTplPath := filepath.Join(targetDir, "templates/_dapp_helpers.tpl")
	f, err := os.Create(helpersTplPath)
	if err != nil {
		return nil, fmt.Errorf("unable to create %s: %s", helpersTplPath, err)
	}
	_, err = f.Write(DappChartHelpersTpl)
	if err != nil {
		return nil, fmt.Errorf("unable to write %s: %s", helpersTplPath, err)
	}

	defaultSecretValues := filepath.Join(projectDir, ProjectDefaultSecretValuesFile)
	if _, err := os.Stat(defaultSecretValues); !os.IsNotExist(err) {
		err := dappChart.SetSecretValuesFile(defaultSecretValues, m)
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
			newPath := filepath.Join(targetDir, DappChartDecodedSecretDir, relativePath)

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

	return dappChart, nil
}
