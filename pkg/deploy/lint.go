package deploy

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/flant/logboek"

	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/deploy/helm"
	"github.com/flant/werf/pkg/deploy/werf_chart"
	"github.com/flant/werf/pkg/images_manager"
	"github.com/flant/werf/pkg/tag_strategy"
	"github.com/flant/werf/pkg/util/secretvalues"
)

type LintOptions struct {
	Values          []string
	SecretValues    []string
	Set             []string
	SetString       []string
	Env             string
	IgnoreSecretKey bool
}

func RunLint(projectDir string, werfConfig *config.WerfConfig, imagesRepoManager images_manager.ImagesRepoManager, images []images_manager.ImageInfoGetter, commonTag string, tagStrategy tag_strategy.TagStrategy, opts LintOptions) error {
	logboek.Debug.LogF("Lint options: %#v\n", opts)

	m, err := GetSafeSecretManager(projectDir, opts.SecretValues, opts.IgnoreSecretKey)
	if err != nil {
		return err
	}

	namespace := "NAMESPACE"

	serviceValues, err := GetServiceValues(werfConfig.Meta.Project, imagesRepoManager, namespace, commonTag, tagStrategy, images, ServiceValuesOptions{Env: opts.Env})
	if err != nil {
		return fmt.Errorf("error creating service values: %s", err)
	}

	projectChartDir := filepath.Join(projectDir, werf_chart.ProjectHelmChartDirName)
	werfChart, err := PrepareWerfChart(werfConfig.Meta.Project, projectChartDir, opts.Env, m, opts.SecretValues, serviceValues)
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
