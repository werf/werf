package deploy

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ghodss/yaml"

	"github.com/flant/kubedog/pkg/kube"
	"github.com/flant/logboek"

	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/deploy/helm"
	"github.com/flant/werf/pkg/deploy/werf_chart"
	"github.com/flant/werf/pkg/tag_strategy"
)

type DeployOptions struct {
	Values               []string
	SecretValues         []string
	Set                  []string
	SetString            []string
	Timeout              time.Duration
	Env                  string
	UserExtraAnnotations map[string]string
	UserExtraLabels      map[string]string
	IgnoreSecretKey      bool
}

func Deploy(projectDir, imagesRepo, release, namespace, tag string, tagStrategy tag_strategy.TagStrategy, werfConfig *config.WerfConfig, helmReleaseStorageNamespace, helmReleaseStorageType string, opts DeployOptions) error {
	var logBlockErr error
	var werfChart *werf_chart.WerfChart

	logboek.LogBlock("Deploy options", logboek.LogBlockOptions{}, func() {
		if kube.Context != "" {
			logboek.LogF("Using kube context: %s\n", kube.Context)
		}
		logboek.LogF("Using helm release storage namespace: %s\n", helmReleaseStorageNamespace)
		logboek.LogF("Using helm release storage type: %s\n", helmReleaseStorageType)
		logboek.LogF("Using helm release name: %s\n", release)
		logboek.LogF("Using kubernetes namespace: %s\n", namespace)
		logboek.LogLn()

		images := GetImagesInfoGetters(werfConfig.Images, imagesRepo, tag, false)

		m, err := GetSafeSecretManager(projectDir, opts.SecretValues, opts.IgnoreSecretKey)
		if err != nil {
			logBlockErr = err
			return
		}

		serviceValues, err := GetServiceValues(werfConfig.Meta.Project, imagesRepo, namespace, tag, tagStrategy, images, ServiceValuesOptions{Env: opts.Env})
		if err != nil {
			logBlockErr = fmt.Errorf("error creating service values: %s", err)
			return
		}

		serviceValuesRaw, _ := yaml.Marshal(serviceValues)
		logboek.LogLn()
		logboek.LogLn("Using service values:")
		logboek.LogLn(logboek.FitText(string(serviceValuesRaw), logboek.FitTextOptions{ExtraIndentWidth: 2}))

		werfChart, err = PrepareWerfChart(GetTmpWerfChartPath(werfConfig.Meta.Project), werfConfig.Meta.Project, projectDir, opts.Env, m, opts.SecretValues, serviceValues)
		if err != nil {
			logBlockErr = err
			return
		}

		werfChart.MergeExtraAnnotations(opts.UserExtraAnnotations)
		werfChart.MergeExtraLabels(opts.UserExtraLabels)
		werfChart.LogExtraAnnotations()
		werfChart.LogExtraLabels()
	})
	logboek.LogOptionalLn()

	if werfChart != nil {
		defer ReleaseTmpWerfChart(werfChart.ChartDir)
	}

	if logBlockErr != nil {
		return logBlockErr
	}

	if err := helm.WithExtra(werfChart.ExtraAnnotations, werfChart.ExtraLabels, func() error {
		return werfChart.Deploy(release, namespace, helm.ChartOptions{
			Timeout: opts.Timeout,
			ChartValuesOptions: helm.ChartValuesOptions{
				Set:       opts.Set,
				SetString: opts.SetString,
				Values:    opts.Values,
			},
		})
	}); err != nil {
		replaceOld := fmt.Sprintf("%s/", werfChart.Name)
		replaceNew := fmt.Sprintf("%s/", ".helm")
		errMsg := strings.Replace(err.Error(), replaceOld, replaceNew, -1)
		return errors.New(errMsg)
	}

	return nil
}
