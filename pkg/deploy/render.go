package deploy

import (
	"io"

	"github.com/flant/logboek"

	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/deploy/helm"
	"github.com/werf/werf/pkg/images_manager"
	"github.com/werf/werf/pkg/tag_strategy"
)

type RenderOptions struct {
	ReleaseName          string
	Namespace            string
	WithoutImagesRepo    bool
	Values               []string
	SecretValues         []string
	Set                  []string
	SetString            []string
	Env                  string
	UserExtraAnnotations map[string]string
	UserExtraLabels      map[string]string
	IgnoreSecretKey      bool
}

func RunRender(out io.Writer, projectDir, helmChartDir string, werfConfig *config.WerfConfig, imagesRepository string, images []images_manager.ImageInfoGetter, commonTag string, tagStrategy tag_strategy.TagStrategy, opts RenderOptions) error {
	logboek.Debug.LogF("Render options: %#v\n", opts)

	m, err := GetSafeSecretManager(projectDir, helmChartDir, opts.SecretValues, opts.IgnoreSecretKey)
	if err != nil {
		return err
	}

	serviceValues, err := GetServiceValues(werfConfig.Meta.Project, imagesRepository, opts.Namespace, commonTag, tagStrategy, images, ServiceValuesOptions{Env: opts.Env})
	if err != nil {
		return err
	}

	werfChart, err := PrepareWerfChart(werfConfig.Meta.Project, helmChartDir, opts.Env, m, opts.SecretValues, serviceValues)
	if err != nil {
		return err
	}

	werfChart.MergeExtraAnnotations(opts.UserExtraAnnotations)
	werfChart.MergeExtraLabels(opts.UserExtraLabels)
	werfChart.LogExtraAnnotations()
	werfChart.LogExtraLabels()

	renderOptions := helm.RenderOptions{
		ShowNotes: false,
	}

	helm.WerfTemplateEngine.InitWerfEngineExtraTemplatesFunctions(werfChart.DecodedSecretFilesData)
	patchLoadChartfile(werfChart.Name)

	return helm.WerfTemplateEngineWithExtraAnnotationsAndLabels(werfChart.ExtraAnnotations, werfChart.ExtraLabels, func() error {
		return helm.Render(
			out,
			werfChart.ChartDir,
			opts.ReleaseName,
			opts.Namespace,
			append(werfChart.Values, opts.Values...),
			werfChart.SecretValues,
			append(werfChart.Set, opts.Set...),
			append(werfChart.SetString, opts.SetString...),
			renderOptions)
	})
}
