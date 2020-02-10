package deploy

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/flant/logboek"

	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/deploy/helm"
	"github.com/flant/werf/pkg/deploy/werf_chart"
	"github.com/flant/werf/pkg/images_manager"
	"github.com/flant/werf/pkg/tag_strategy"
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

func RunRender(out io.Writer, projectDir string, werfConfig *config.WerfConfig, imagesRepoManager images_manager.ImagesRepoManager, images []images_manager.ImageInfoGetter, commonTag string, tagStrategy tag_strategy.TagStrategy, opts RenderOptions) error {
	if debug() {
		fmt.Fprintf(logboek.GetOutStream(), "Render options: %#v\n", opts)
	}

	m, err := GetSafeSecretManager(projectDir, opts.SecretValues, opts.IgnoreSecretKey)
	if err != nil {
		return err
	}

	serviceValues, err := GetServiceValues(werfConfig.Meta.Project, imagesRepoManager, opts.Namespace, commonTag, tagStrategy, images, ServiceValuesOptions{Env: opts.Env})
	if err != nil {
		return err
	}

	projectChartDir := filepath.Join(projectDir, werf_chart.ProjectHelmChartDirName)
	werfChart, err := PrepareWerfChart(werfConfig.Meta.Project, projectChartDir, opts.Env, m, opts.SecretValues, serviceValues)
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
