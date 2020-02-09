package lint

import (
	"fmt"
	"path/filepath"

	"github.com/flant/logboek"
	"github.com/flant/shluz"
	"github.com/spf13/cobra"

	"github.com/flant/werf/cmd/werf/common"
	helm_common "github.com/flant/werf/cmd/werf/helm/common"
	"github.com/flant/werf/pkg/deploy"
	"github.com/flant/werf/pkg/deploy/helm"
	"github.com/flant/werf/pkg/deploy/helm/matrix_tests"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/tmp_manager"
	"github.com/flant/werf/pkg/true_git"
	"github.com/flant/werf/pkg/werf"
)

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "matrix-tests",
		Short:                 "Execute matrix-tests for the werf chart (matrix_test.yaml must exist in chart folder).",
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMatrixTests()
		},
	}

	common.SetupDir(&CommonCmdData, cmd)
	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)

	common.SetupNamespace(&CommonCmdData, cmd)
	common.SetupRelease(&CommonCmdData, cmd)
	common.SetupEnvironment(&CommonCmdData, cmd)
	common.SetupDockerConfig(&CommonCmdData, cmd, "")

	common.SetupImagesRepo(&CommonCmdData, cmd)
	common.SetupImagesRepoMode(&CommonCmdData, cmd)
	common.SetupTag(&CommonCmdData, cmd)

	return cmd
}

func runMatrixTests() error {
	tmp_manager.AutoGCEnabled = false

	if err := werf.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := shluz.Init(filepath.Join(werf.GetServiceDir(), "locks")); err != nil {
		return err
	}

	if err := true_git.Init(true_git.Options{Out: logboek.GetOutStream(), Err: logboek.GetErrStream()}); err != nil {
		return err
	}

	if err := deploy.Init(deploy.InitOptions{HelmInitOptions: helm.InitOptions{WithoutKube: true}}); err != nil {
		return err
	}

	if err := docker.Init(*CommonCmdData.DockerConfig); err != nil {
		return err
	}

	projectDir, err := common.GetProjectDir(&CommonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	werfConfig, err := common.GetWerfConfig(projectDir)
	if err != nil {
		return fmt.Errorf("bad config: %s", err)
	}

	optionalImagesRepo, err := common.GetOptionalImagesRepo(werfConfig.Meta.Project, &CommonCmdData)
	if err != nil {
		return err
	}

	withoutImagesRepo := true
	if optionalImagesRepo != "" {
		withoutImagesRepo = false
	}

	imagesRepo := helm_common.GetImagesRepoOrStub(optionalImagesRepo)

	imagesRepoMode, err := common.GetImagesRepoMode(&CommonCmdData)
	if err != nil {
		return err
	}

	imagesRepoManager, err := common.GetImagesRepoManager(imagesRepo, imagesRepoMode)
	if err != nil {
		return err
	}

	env := helm_common.GetEnvironmentOrStub(*CommonCmdData.Environment)

	release, err := common.GetHelmRelease(*CommonCmdData.Release, env, werfConfig)
	if err != nil {
		return err
	}

	namespace, err := common.GetKubernetesNamespace(*CommonCmdData.Namespace, env, werfConfig)
	if err != nil {
		return err
	}

	tag, tagStrategy, err := helm_common.GetTagOrStub(&CommonCmdData)
	if err != nil {
		return err
	}

	return matrix_tests.RunMatrixTests("", projectDir, werfConfig, deploy.RenderOptions{
		ReleaseName:       release,
		Tag:               tag,
		TagStrategy:       tagStrategy,
		Namespace:         namespace,
		ImagesRepoManager: imagesRepoManager,
		WithoutImagesRepo: withoutImagesRepo,
		Values:            []string{},
		SecretValues:      []string{},
		Set:               []string{},
		SetString:         []string{},
		Env:               env,
		IgnoreSecretKey:   true,
	})
}
