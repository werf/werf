package deploy

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/deploy/helm"
	"github.com/flant/werf/pkg/tag_strategy"
)

type RenderOptions struct {
	Values               []string
	SecretValues         []string
	Set                  []string
	SetString            []string
	Env                  string
	UserExtraAnnotations map[string]string
	UserExtraLabels      map[string]string
	IgnoreSecretKey      bool
}

func RunRender(projectDir string, werfConfig *config.WerfConfig, opts RenderOptions) error {
	if debug() {
		fmt.Fprintf(logboek.GetOutStream(), "Render options: %#v\n", opts)
	}

	m, err := GetSafeSecretManager(projectDir, opts.SecretValues, opts.IgnoreSecretKey)
	if err != nil {
		return err
	}

	releaseName := "RELEASE_NAME"
	imagesRepo := "REPO"
	tag := "GIT_BRANCH"
	tagStrategy := tag_strategy.GitBranch
	namespace := "NAMESPACE"

	images := GetImagesInfoGetters(werfConfig.Images, imagesRepo, tag, true)

	serviceValues, err := GetServiceValues(werfConfig.Meta.Project, imagesRepo, namespace, tag, tagStrategy, images, ServiceValuesOptions{Env: opts.Env})

	werfChart, err := PrepareWerfChart(GetTmpWerfChartPath(werfConfig.Meta.Project), werfConfig.Meta.Project, projectDir, opts.Env, m, opts.SecretValues, serviceValues)
	if err != nil {
		return err
	}
	defer ReleaseTmpWerfChart(werfChart.ChartDir)

	werfChart.MergeExtraAnnotations(opts.UserExtraAnnotations)
	werfChart.MergeExtraLabels(opts.UserExtraLabels)
	werfChart.LogExtraAnnotations()
	werfChart.LogExtraLabels()

	renderOptions := helm.RenderOptions{
		ShowNotes: false,
	}

	if err := helm.WithExtra(werfChart.ExtraAnnotations, werfChart.ExtraLabels, func() error {
		return helm.Render(
			os.Stdout,
			werfChart.ChartDir,
			releaseName,
			namespace,
			append(werfChart.Values, opts.Values...),
			append(werfChart.Set, opts.Set...),
			append(werfChart.SetString, opts.SetString...),
			renderOptions)
	}); err != nil {
		replaceOld := fmt.Sprintf("%s/", werfChart.Name)
		replaceNew := fmt.Sprintf("%s/", ".helm")
		errMsg := strings.Replace(err.Error(), replaceOld, replaceNew, -1)
		return errors.New(errMsg)
	}

	return nil
}
