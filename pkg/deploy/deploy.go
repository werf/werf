package deploy

import (
	"fmt"
	"time"

	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/deploy/helm"
	"github.com/flant/werf/pkg/logger"
	"github.com/flant/werf/pkg/tag_scheme"
)

type DeployOptions struct {
	Values       []string
	SecretValues []string
	Set          []string
	SetString    []string
	Timeout      time.Duration
	KubeContext  string
}

type DockerAuthorizer interface {
	Login(repo string) error
}

func Deploy(projectDir, imagesRepo, release, namespace string, tag string, tagScheme tag_scheme.TagScheme, werfConfig *config.WerfConfig, dockerAuthorizer DockerAuthorizer, opts DeployOptions) error {
	logger.LogInfoF("Using Helm release name: %s\n", release)
	logger.LogInfoF("Using Kubernetes namespace: %s\n", namespace)

	images := GetImagesInfoGetters(werfConfig.Images, imagesRepo, tag, false)

	m, err := GetSafeSecretManager(projectDir, opts.SecretValues)
	if err != nil {
		return fmt.Errorf("cannot get project secret: %s", err)
	}

	if err := dockerAuthorizer.Login(imagesRepo); err != nil {
		return fmt.Errorf("docker login failed: %s", err)
	}

	serviceValues, err := GetServiceValues(werfConfig.Meta.Project, imagesRepo, namespace, tag, tagScheme, images)
	if err != nil {
		return fmt.Errorf("error creating service values: %s", err)
	}

	werfChart, err := PrepareWerfChart(GetTmpWerfChartPath(werfConfig.Meta.Project), werfConfig.Meta.Project, projectDir, m, opts.SecretValues, serviceValues)
	if err != nil {
		return err
	}

	return werfChart.Deploy(release, namespace, helm.HelmChartOptions{
		CommonHelmOptions: helm.CommonHelmOptions{KubeContext: opts.KubeContext},
		Timeout:           opts.Timeout,
		Set:               opts.Set,
		SetString:         opts.SetString,
		Values:            opts.Values,
	})
}
