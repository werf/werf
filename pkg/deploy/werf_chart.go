package deploy

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/deploy/secret"
	"github.com/flant/werf/pkg/deploy/werf_chart"
	"github.com/flant/werf/pkg/werf"
	uuid "github.com/satori/go.uuid"
)

func GetTmpWerfChartPath(projectName string) string {
	return filepath.Join(werf.GetTmpDir(), fmt.Sprintf("werf-chart-%s", uuid.NewV4().String()), projectName)
}

func ReleaseTmpWerfChart(tmpWerfChartPath string) {
	// NOTE: remove parent dir, see GetTmpWerfChartPath func
	os.RemoveAll(filepath.Dir(tmpWerfChartPath))
}

func PrepareWerfChart(targetDir string, projectName, projectDir string, m secret.Manager, secretValues []string, serviceValues map[string]interface{}) (*werf_chart.WerfChart, error) {
	werfChart, err := werf_chart.CreateNewWerfChart(projectName, projectDir, targetDir, m)
	if err != nil {
		return nil, err
	}

	for _, path := range secretValues {
		err = werfChart.SetSecretValuesFile(path, m)
		if err != nil {
			return nil, err
		}
	}

	if serviceValues != nil {
		err = werfChart.SetValues(serviceValues)
		if err != nil {
			return nil, err
		}
	}

	if debug() {
		fmt.Fprintf(logboek.GetOutStream(), "Werf chart: %#v\n", werfChart)
	}

	return werfChart, nil
}

func debug() bool {
	return os.Getenv("WERF_DEPLOY_DEBUG") == "1"
}
