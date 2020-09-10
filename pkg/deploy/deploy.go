package deploy

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ghodss/yaml"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/deploy/helm"
	"github.com/werf/werf/pkg/deploy/lock_manager"
	"github.com/werf/werf/pkg/deploy/werf_chart"
	"github.com/werf/werf/pkg/images_manager"
	"github.com/werf/werf/pkg/tag_strategy"
	"github.com/werf/werf/pkg/util/secretvalues"
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
	ThreeWayMergeMode    helm.ThreeWayMergeModeType
	DryRun               bool
}

func Deploy(ctx context.Context, projectName, projectDir, helmChartDir string, imagesRepository string, images []images_manager.ImageInfoGetter, release, namespace, commonTag string, tagStrategy tag_strategy.TagStrategy, werfConfig *config.WerfConfig, helmReleaseStorageNamespace, helmReleaseStorageType string, opts DeployOptions) error {
	if opts.DryRun {
		fmt.Printf("Deploy DryRun\n")
		return nil
	}

	lockManager, err := lock_manager.NewLockManager(namespace)
	if err != nil {
		return err
	}

	if lock, err := lockManager.LockRelease(ctx, release); err != nil {
		return err
	} else {
		defer lockManager.Unlock(lock)
	}

	var werfChart *werf_chart.WerfChart

	if err := logboek.Context(ctx).Default().LogBlock("Deploy options").DoError(func() error {
		if kube.Context != "" {
			logboek.Context(ctx).LogF("Kube-config context: %s\n", kube.Context)
		}
		logboek.Context(ctx).LogF("Kubernetes namespace: %s\n", namespace)
		logboek.Context(ctx).LogF("Helm release storage namespace: %s\n", helmReleaseStorageNamespace)
		logboek.Context(ctx).LogF("Helm release storage type: %s\n", helmReleaseStorageType)
		logboek.Context(ctx).LogF("Helm release name: %s\n", release)

		m, err := GetSafeSecretManager(ctx, projectDir, helmChartDir, opts.SecretValues, opts.IgnoreSecretKey)
		if err != nil {
			return err
		}

		serviceValues, err := GetServiceValues(ctx, werfConfig.Meta.Project, imagesRepository, namespace, commonTag, tagStrategy, images, ServiceValuesOptions{Env: opts.Env})
		if err != nil {
			return fmt.Errorf("error creating service values: %s", err)
		}

		serviceValuesRaw, _ := yaml.Marshal(serviceValues)
		serviceValuesRawStr := strings.TrimRight(string(serviceValuesRaw), "\n")
		logboek.Context(ctx).Info().LogBlock(fmt.Sprintf("Service values")).Do(func() {
			logboek.Context(ctx).Info().LogLn(serviceValuesRawStr)
		})

		werfChart, err = PrepareWerfChart(ctx, werfConfig.Meta.Project, helmChartDir, opts.Env, m, opts.SecretValues, serviceValues)
		if err != nil {
			return err
		}
		helm.SetReleaseLogSecretValuesToMask(werfChart.SecretValuesToMask)

		werfChart.MergeExtraAnnotations(opts.UserExtraAnnotations)
		werfChart.MergeExtraLabels(opts.UserExtraLabels)
		werfChart.LogExtraAnnotations(ctx)
		werfChart.LogExtraLabels(ctx)

		return nil
	}); err != nil {
		logboek.Context(ctx).LogOptionalLn()
		return err
	}

	logboek.Context(ctx).LogOptionalLn()

	helm.WerfTemplateEngine.InitWerfEngineExtraTemplatesFunctions(werfChart.DecodedSecretFilesData)
	patchLoadChartfile(werfChart.Name)

	err = helm.WerfTemplateEngineWithExtraAnnotationsAndLabels(werfChart.ExtraAnnotations, werfChart.ExtraLabels, func() error {
		return werfChart.Deploy(ctx, release, namespace, helm.ChartOptions{
			Timeout: opts.Timeout,
			ChartValuesOptions: helm.ChartValuesOptions{
				Set:       opts.Set,
				SetString: opts.SetString,
				Values:    opts.Values,
			},
			ThreeWayMergeMode: opts.ThreeWayMergeMode,
		})
	})

	if err != nil {
		return fmt.Errorf("%s", secretvalues.MaskSecretValuesInString(werfChart.SecretValuesToMask, err.Error()))
	}

	return nil
}

func patchLoadChartfile(chartName string) {
	boundedFunc := helm.LoadChartfileFunc
	helm.LoadChartfileFunc = func(chartPath string) (*chart.Chart, error) {
		var c *chart.Chart

		if err := chartutil.WithSkipChartYamlFileValidation(true, func() error {
			var err error
			if c, err = boundedFunc(chartPath); err != nil {
				return err
			}

			return nil
		}); err != nil {
			return nil, err
		}

		c.Metadata = &chart.Metadata{
			Name:    chartName,
			Version: "0.1.0",
			Engine:  helm.WerfTemplateEngineName,
		}

		return c, nil
	}
}
