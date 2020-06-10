package deploy

import (
	"fmt"
	"os"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/deploy/helm"
	"github.com/werf/werf/pkg/images_manager"
	"github.com/werf/werf/pkg/tag_strategy"
	"github.com/werf/werf/pkg/util/secretvalues"
)

type LintOptions struct {
	Values          []string
	SecretValues    []string
	Set             []string
	SetString       []string
	Env             string
	IgnoreSecretKey bool
}

func RunLint(projectDir, helmChartDir string, werfConfig *config.WerfConfig, imagesRepository string, images []images_manager.ImageInfoGetter, commonTag string, tagStrategy tag_strategy.TagStrategy, opts LintOptions) error {
	logboek.Debug.LogF("Lint options: %#v\n", opts)

	m, err := GetSafeSecretManager(projectDir, helmChartDir, opts.SecretValues, opts.IgnoreSecretKey)
	if err != nil {
		return err
	}

	namespace := "NAMESPACE"

	serviceValues, err := GetServiceValues(werfConfig.Meta.Project, imagesRepository, namespace, commonTag, tagStrategy, images, ServiceValuesOptions{Env: opts.Env})
	if err != nil {
		return fmt.Errorf("error creating service values: %s", err)
	}

	werfChart, err := PrepareWerfChart(werfConfig.Meta.Project, helmChartDir, opts.Env, m, opts.SecretValues, serviceValues)
	if err != nil {
		return err
	}
	helm.SetReleaseLogSecretValuesToMask(werfChart.SecretValuesToMask)

	helm.WerfTemplateEngine.InitWerfEngineExtraTemplatesFunctions(werfChart.DecodedSecretFilesData)
	patchLoadChartfile(werfChart.Name)

	if err := helm.Lint(
		os.Stdout,
		werfChart.ChartDir,
		namespace,
		append(werfChart.Values, opts.Values...),
		werfChart.SecretValues,
		append(werfChart.Set, opts.Set...),
		append(werfChart.SetString, opts.SetString...),
		helm.LintOptions{Strict: true},
	); err != nil {
		return fmt.Errorf("%s", secretvalues.MaskSecretValuesInString(werfChart.SecretValuesToMask, err.Error()))
	}

	return nil
}
