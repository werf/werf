package deploy

import (
	"fmt"
	"time"

	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/deploy/helm"
	"github.com/flant/werf/pkg/tag_strategy"
)

type DeployOptions struct {
	Values       []string
	SecretValues []string
	Set          []string
	SetString    []string
	Timeout      time.Duration
	Env          string
}

func Deploy(projectDir, imagesRepo, release, namespace, tag string, tagStrategy tag_strategy.TagStrategy, werfConfig *config.WerfConfig, opts DeployOptions) error {
	images := GetImagesInfoGetters(werfConfig.Images, imagesRepo, tag, false)

	m, err := GetSafeSecretManager(projectDir, opts.SecretValues)
	if err != nil {
		return fmt.Errorf("cannot get project secret: %s", err)
	}

	serviceValues, err := GetServiceValues(werfConfig.Meta.Project, imagesRepo, namespace, tag, tagStrategy, images, ServiceValuesOptions{Env: opts.Env})
	if err != nil {
		return fmt.Errorf("error creating service values: %s", err)
	}

	werfChart, err := PrepareWerfChart(GetTmpWerfChartPath(werfConfig.Meta.Project), werfConfig.Meta.Project, projectDir, m, opts.SecretValues, serviceValues)
	if err != nil {
		return err
	}
	defer ReleaseTmpWerfChart(werfChart.ChartDir)

	logboek.OptionalLnModeOn()
	return werfChart.Deploy(release, namespace, helm.HelmChartOptions{
		Timeout: opts.Timeout,
		HelmChartValuesOptions: helm.HelmChartValuesOptions{
			Set:       opts.Set,
			SetString: opts.SetString,
			Values:    opts.Values,
		},
	})
}
