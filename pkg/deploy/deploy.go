package deploy

import (
	"fmt"
	"time"

	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/deploy/helm"
	"github.com/flant/werf/pkg/logger"
	"github.com/flant/werf/pkg/tag_strategy"
	"github.com/ghodss/yaml"
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
	logger.LogInfoF("Using helm release name: %s\n", release)
	logger.LogInfoF("Using kubernetes namespace: %s\n", namespace)

	images := GetImagesInfoGetters(werfConfig.Images, imagesRepo, tag, false)

	m, err := GetSafeSecretManager(projectDir, opts.SecretValues)
	if err != nil {
		return fmt.Errorf("cannot get project secret: %s", err)
	}

	serviceValues, err := GetServiceValues(werfConfig.Meta.Project, imagesRepo, namespace, tag, tagStrategy, images, ServiceValuesOptions{Env: opts.Env})
	if err != nil {
		return fmt.Errorf("error creating service values: %s", err)
	}

	serviceValuesRaw, _ := yaml.Marshal(serviceValues)
	logger.LogInfoF("Using service values:\n%s", serviceValuesRaw)
	logger.OptionalLnModeOn()

	werfChart, err := PrepareWerfChart(GetTmpWerfChartPath(werfConfig.Meta.Project), werfConfig.Meta.Project, projectDir, m, opts.SecretValues, serviceValues)
	if err != nil {
		return err
	}
	defer ReleaseTmpWerfChart(werfChart.ChartDir)

	logger.OptionalLnModeOn()
	return werfChart.Deploy(release, namespace, helm.HelmChartOptions{
		Timeout: opts.Timeout,
		HelmChartValuesOptions: helm.HelmChartValuesOptions{
			Set:       opts.Set,
			SetString: opts.SetString,
			Values:    opts.Values,
		},
	})
}
