package render

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/flant/logboek"

	"github.com/werf/werf/cmd/werf/common"
	helm_common "github.com/werf/werf/cmd/werf/helm/common"
	"github.com/werf/werf/pkg/deploy"
	"github.com/werf/werf/pkg/deploy/helm"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/images_manager"
	"github.com/werf/werf/pkg/tmp_manager"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/werf"
)

var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	var outputFilePath string

	cmd := &cobra.Command{
		Use:                   "render",
		Short:                 "Render werf chart templates to stdout",
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfSecretKey),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			return runRender(outputFilePath)
		},
	}

	common.SetupDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)

	common.SetupHelmChartDir(&commonCmdData, cmd)
	common.SetupNamespace(&commonCmdData, cmd)
	common.SetupRelease(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)
	common.SetupDockerConfig(&commonCmdData, cmd, "")
	common.SetupAddAnnotations(&commonCmdData, cmd)
	common.SetupAddLabels(&commonCmdData, cmd)

	common.SetupSet(&commonCmdData, cmd)
	common.SetupSetString(&commonCmdData, cmd)
	common.SetupValues(&commonCmdData, cmd)
	common.SetupSecretValues(&commonCmdData, cmd)
	common.SetupIgnoreSecretKey(&commonCmdData, cmd)

	common.SetupImagesRepoOptions(&commonCmdData, cmd)

	common.SetupTag(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)

	cmd.Flags().StringVarP(&outputFilePath, "output-file-path", "o", "", "Write to file instead of stdout")

	return cmd
}

func runRender(outputFilePath string) error {
	tmp_manager.AutoGCEnabled = false

	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := true_git.Init(true_git.Options{Out: logboek.GetOutStream(), Err: logboek.GetErrStream(), LiveGitOutput: *commonCmdData.LogVerbose || *commonCmdData.LogDebug}); err != nil {
		return err
	}

	if err := deploy.Init(deploy.InitOptions{HelmInitOptions: helm.InitOptions{WithoutKube: true}}); err != nil {
		return err
	}

	if err := docker.Init(*commonCmdData.DockerConfig, *commonCmdData.LogVerbose, *commonCmdData.LogDebug); err != nil {
		return err
	}

	projectDir, err := common.GetProjectDir(&commonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	helmChartDir, err := common.GetHelmChartDir(projectDir, &commonCmdData)
	if err != nil {
		return fmt.Errorf("getting helm chart dir failed: %s", err)
	}

	werfConfig, err := common.GetRequiredWerfConfig(projectDir, &commonCmdData, false)
	if err != nil {
		return fmt.Errorf("unable to load werf config: %s", err)
	}

	projectName := werfConfig.Meta.Project

	optionalImagesRepo, err := common.GetOptionalImagesRepoAddress(projectName, &commonCmdData)
	if err != nil {
		return err
	}

	withoutImagesRepo := true
	if optionalImagesRepo != "" {
		withoutImagesRepo = false
	}

	imagesRepo, err := common.GetImagesRepoWithOptionalStubRepoAddress(projectName, &commonCmdData)
	if err != nil {
		return err
	}

	env := helm_common.GetEnvironmentOrStub(*commonCmdData.Environment)

	release, err := common.GetHelmRelease(*commonCmdData.Release, env, werfConfig)
	if err != nil {
		return err
	}

	namespace, err := common.GetKubernetesNamespace(*commonCmdData.Namespace, env, werfConfig)
	if err != nil {
		return err
	}

	tag, tagStrategy, err := helm_common.GetTagOrStub(&commonCmdData)
	if err != nil {
		return err
	}

	userExtraAnnotations, err := common.GetUserExtraAnnotations(&commonCmdData)
	if err != nil {
		return err
	}

	userExtraLabels, err := common.GetUserExtraLabels(&commonCmdData)
	if err != nil {
		return err
	}

	var imagesInfoGetters []images_manager.ImageInfoGetter
	var imagesNames []string
	for _, imageConfig := range werfConfig.StapelImages {
		imagesNames = append(imagesNames, imageConfig.Name)
	}
	for _, imageConfig := range werfConfig.ImagesFromDockerfile {
		imagesNames = append(imagesNames, imageConfig.Name)
	}
	for _, imageName := range imagesNames {
		d := &images_manager.ImageInfo{
			ImagesRepo:      imagesRepo,
			Tag:             tag,
			Name:            imageName,
			WithoutRegistry: withoutImagesRepo,
		}
		imagesInfoGetters = append(imagesInfoGetters, d)
	}

	buf := bytes.NewBuffer([]byte{})
	if err := deploy.RunRender(buf, projectDir, helmChartDir, werfConfig, imagesRepo.String(), imagesInfoGetters, tag, tagStrategy, deploy.RenderOptions{
		ReleaseName:          release,
		Namespace:            namespace,
		WithoutImagesRepo:    withoutImagesRepo,
		Values:               *commonCmdData.Values,
		SecretValues:         *commonCmdData.SecretValues,
		Set:                  *commonCmdData.Set,
		SetString:            *commonCmdData.SetString,
		Env:                  env,
		UserExtraAnnotations: userExtraAnnotations,
		UserExtraLabels:      userExtraLabels,
		IgnoreSecretKey:      *commonCmdData.IgnoreSecretKey,
	}); err != nil {
		return err
	}

	if outputFilePath != "" {
		if err := saveRenderedChart(outputFilePath, buf); err != nil {
			return err
		}
	} else {
		fmt.Printf("%s", buf.String())
	}

	return nil
}

func saveRenderedChart(outputFilePath string, buf *bytes.Buffer) error {
	if err := os.MkdirAll(filepath.Dir(outputFilePath), 0777); err != nil {
		return err
	}

	if err := ioutil.WriteFile(outputFilePath, buf.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}
