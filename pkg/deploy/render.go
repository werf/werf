package deploy

import (
	"context"
	"io"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/deploy/helm"
	"github.com/werf/werf/pkg/image"
)

type RenderOptions struct {
	ReleaseName          string
	Namespace            string
	Values               []string
	SecretValues         []string
	Set                  []string
	SetString            []string
	Env                  string
	UserExtraAnnotations map[string]string
	UserExtraLabels      map[string]string
	IgnoreSecretKey      bool
}

func RunRender(ctx context.Context, out io.Writer, projectDir, helmChartDir string, werfConfig *config.WerfConfig, imagesRepository string, images []*image.InfoGetter, opts RenderOptions) error {
	logboek.Context(ctx).Debug().LogF("Render options: %#v\n", opts)

	m, err := GetSafeSecretManager(ctx, projectDir, helmChartDir, opts.SecretValues, opts.IgnoreSecretKey)
	if err != nil {
		return err
	}

	serviceValues, err := GetServiceValues(ctx, werfConfig.Meta.Project, imagesRepository, opts.Namespace, images, ServiceValuesOptions{Env: opts.Env})
	if err != nil {
		return err
	}

	werfChart, err := PrepareWerfChart(ctx, werfConfig.Meta.Project, helmChartDir, opts.Env, m, opts.SecretValues, serviceValues)
	if err != nil {
		return err
	}

	werfChart.MergeExtraAnnotations(opts.UserExtraAnnotations)
	werfChart.MergeExtraLabels(opts.UserExtraLabels)
	werfChart.LogExtraAnnotations(ctx)
	werfChart.LogExtraLabels(ctx)

	renderOptions := helm.RenderOptions{
		ShowNotes: false,
	}

	helm.WerfTemplateEngine.InitWerfEngineExtraTemplatesFunctions(werfChart.DecodedSecretFilesData)
	patchLoadChartfile(werfChart.Name)

	return helm.WerfTemplateEngineWithExtraAnnotationsAndLabels(werfChart.ExtraAnnotations, werfChart.ExtraLabels, func() error {
		return helm.Render(
			ctx,
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
