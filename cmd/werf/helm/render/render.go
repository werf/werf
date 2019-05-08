package render

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flant/logboek"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/deploy"
	"github.com/flant/werf/pkg/deploy/helm"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/true_git"
	"github.com/flant/werf/pkg/werf"
)

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "render",
		Short:                 "Render Werf chart templates to stdout",
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfSecretKey),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRender()
		},
	}

	common.SetupDir(&CommonCmdData, cmd)
	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)

	common.SetupEnvironment(&CommonCmdData, cmd)
	common.SetupDockerConfig(&CommonCmdData, cmd, "")
	common.SetupAddAnnotations(&CommonCmdData, cmd)
	common.SetupAddLabels(&CommonCmdData, cmd)

	common.SetupSet(&CommonCmdData, cmd)
	common.SetupSetString(&CommonCmdData, cmd)
	common.SetupValues(&CommonCmdData, cmd)
	common.SetupSecretValues(&CommonCmdData, cmd)
	common.SetupIgnoreSecretKey(&CommonCmdData, cmd)

	return cmd
}

func runRender() error {
	if err := werf.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := lock.Init(); err != nil {
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

	userExtraAnnotations, err := common.GetUserExtraAnnotations(&CommonCmdData)
	if err != nil {
		return err
	}

	userExtraLabels, err := common.GetUserExtraLabels(&CommonCmdData)
	if err != nil {
		return err
	}

	return deploy.RunRender(projectDir, werfConfig, deploy.RenderOptions{
		Values:               *CommonCmdData.Values,
		SecretValues:         *CommonCmdData.SecretValues,
		Set:                  *CommonCmdData.Set,
		SetString:            *CommonCmdData.SetString,
		Env:                  *CommonCmdData.Environment,
		UserExtraAnnotations: userExtraAnnotations,
		UserExtraLabels:      userExtraLabels,
		IgnoreSecretKey:      *CommonCmdData.IgnoreSecretKey,
	})
}
