package get

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/flant/logboek"
	"github.com/flant/shluz"
	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/deploy"
	"github.com/flant/werf/pkg/deploy/helm"
	"github.com/flant/werf/pkg/true_git"
	"github.com/flant/werf/pkg/werf"
	"github.com/spf13/cobra"
)

var CommonCmdData common.CmdData

var CmdData struct {
	helm.GetOptions
}

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "get RELEASE_NAME",
		Short:                 "Get release templates and values by release name",
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ValidateArgumentCount(1, args, cmd); err != nil {
				return err
			}
			return runGet(args[0])
		},
	}

	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)

	common.SetupKubeConfig(&CommonCmdData, cmd)
	common.SetupKubeContext(&CommonCmdData, cmd)
	common.SetupHelmReleaseStorageNamespace(&CommonCmdData, cmd)
	common.SetupHelmReleaseStorageType(&CommonCmdData, cmd)

	cmd.Flags().Int32Var(&CmdData.Revision, "revision", 0, "Get the named release by revision (use latest revision by default)")
	cmd.Flags().StringVar(&CmdData.Template, "template", "", "Go template for formatting the output, eg: {{.Release.Name}}")

	return cmd
}

func runGet(releaseName string) error {
	if err := werf.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := shluz.Init(filepath.Join(werf.GetServiceDir(), "locks")); err != nil {
		return err
	}

	if err := true_git.Init(true_git.Options{Out: logboek.GetOutStream(), Err: logboek.GetErrStream()}); err != nil {
		return err
	}

	helmReleaseStorageType, err := common.GetHelmReleaseStorageType(*CommonCmdData.HelmReleaseStorageType)
	if err != nil {
		return err
	}

	deployInitOptions := deploy.InitOptions{
		HelmInitOptions: helm.InitOptions{
			KubeConfig:                  *CommonCmdData.KubeConfig,
			KubeContext:                 *CommonCmdData.KubeContext,
			HelmReleaseStorageNamespace: *CommonCmdData.HelmReleaseStorageNamespace,
			HelmReleaseStorageType:      helmReleaseStorageType,
			ReleasesMaxHistory:          0,
		},
	}
	if err := deploy.Init(deployInitOptions); err != nil {
		return err
	}

	if err := helm.Get(os.Stdout, releaseName, CmdData.GetOptions); err != nil {
		return err
	}

	return nil
}
