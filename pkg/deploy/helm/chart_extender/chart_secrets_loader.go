package chart_extender

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/werf/werf/pkg/util"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"

	"github.com/werf/werf/pkg/deploy/secret"
	"sigs.k8s.io/yaml"
)

type LoadChartSecretValueFilesOptions struct {
	CustomFiles []string
}

func LoadChartSecretValueFiles(chartDir string, loadedChartFiles []*chart.ChartExtenderBufferedFile, secretsManager secret.Manager, opts LoadChartSecretValueFilesOptions) (map[string]interface{}, error) {
	var res map[string]interface{}

	valuesFilePaths := []string{DefaultSecretValuesFileName}
	for _, path := range opts.CustomFiles {
		relPath := util.GetRelativeToBaseFilepath(chartDir, path)
		valuesFilePaths = append(valuesFilePaths, relPath)
	}

	for _, file := range loadedChartFiles {
		for _, valuesFilePath := range valuesFilePaths {
			if file.Name == valuesFilePath {
				decodedData, err := secretsManager.DecryptYamlData(file.Data)
				if err != nil {
					return nil, fmt.Errorf("cannot decode file %q secret data: %s", file.Name, err)
				}

				rawValues := map[string]interface{}{}
				if err := yaml.Unmarshal(decodedData, &rawValues); err != nil {
					return nil, fmt.Errorf("cannot unmarshal secret values file %s: %s", file.Name, err)
				}

				res = chartutil.CoalesceTables(rawValues, res)
			}
		}
	}

	return res, nil
}

func LoadChartSecretFilesData(chartDir string, loadedChartFiles []*chart.ChartExtenderBufferedFile, secretsManager secret.Manager) (map[string]string, error) {
	res := make(map[string]string)

	for _, file := range loadedChartFiles {
		if !util.IsSubpathOfBasePath(SecretDirName, file.Name) {
			continue
		}

		decodedData, err := secretsManager.Decrypt([]byte(strings.TrimRightFunc(string(file.Data), unicode.IsSpace)))
		if err != nil {
			return nil, fmt.Errorf("error decoding %s: %s", filepath.Join(chartDir, file.Name), err)
		}

		relPath := util.GetRelativeToBaseFilepath(SecretDirName, file.Name)
		res[filepath.ToSlash(relPath)] = string(decodedData)
	}

	return res, nil
}
