package kubectl

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/cmd"
	"k8s.io/kubectl/pkg/cmd/plugin"
	"k8s.io/kubectl/pkg/cmd/util"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/tmp_manager"
	util2 "github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
)

var (
	commonCmdData common.CmdData
	configFlags   *genericclioptions.ConfigFlags
)

func NewCmd(ctx context.Context) *cobra.Command {
	configFlags = genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()

	kubectlCmd := cmd.NewDefaultKubectlCommandWithArgs(cmd.KubectlOptions{
		PluginHandler: cmd.NewDefaultPluginHandler(plugin.ValidPluginFilenamePrefixes),
		Arguments:     os.Args,
		ConfigFlags:   configFlags,
		IOStreams:     genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr},
	})

	common.SetupHomeDir(&commonCmdData, kubectlCmd, common.SetupHomeDirOptions{Persistent: true})
	common.SetupTmpDir(&commonCmdData, kubectlCmd, common.SetupTmpDirOptions{Persistent: true})
	common.SetupKubeConfigBase64(&commonCmdData, kubectlCmd)

	kubeConfigFlag := kubectlCmd.Flag("kubeconfig")
	kubeConfigFlag.Usage = "Path to the kubeconfig file to use for CLI requests (default $WERF_KUBE_CONFIG, or $WERF_KUBECONFIG, or $KUBECONFIG). Ignored if kubeconfig passed as base64."

	kubeContextFlag := kubectlCmd.Flag("context")
	kubeContextFlag.Usage = "The name of the kubeconfig context to use (default $WERF_KUBE_CONTEXT)"
	if werfKubeContextEnvVar := os.Getenv("WERF_KUBE_CONTEXT"); werfKubeContextEnvVar != "" {
		if err := kubectlCmd.Flags().Set("context", werfKubeContextEnvVar); err != nil {
			util.CheckErr(fmt.Errorf("unable to set context flag for kubectl: %w", err))
		}
	}

	skipTlsVerifyRegistryFlag := kubectlCmd.Flag("insecure-skip-tls-verify")
	skipTlsVerifyRegistryFlag.Usage = "If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure (default $WERF_SKIP_TLS_VERIFY_REGISTRY)"
	if isSkipTlsVerifyRegistry := util2.GetBoolEnvironmentDefaultFalse("WERF_SKIP_TLS_VERIFY_REGISTRY"); isSkipTlsVerifyRegistry {
		if err := kubectlCmd.Flags().Set("insecure-skip-tls-verify", "true"); err != nil {
			util.CheckErr(fmt.Errorf("unable to set insecure-skip-tls-verify flag for kubectl: %w", err))
		}
	}

	wrapPreRun(kubectlCmd)

	return kubectlCmd
}

func wrapPreRun(kubectlCmd *cobra.Command) {
	switch {
	case kubectlCmd.PersistentPreRunE != nil:
		oldFunc := kubectlCmd.PersistentPreRunE
		kubectlCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := prePreRun(ctx); err != nil {
				return err
			}
			return oldFunc(cmd, args)
		}
	case kubectlCmd.PersistentPreRun != nil:
		oldFunc := kubectlCmd.PersistentPreRun
		kubectlCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()

			if err := prePreRun(ctx); err != nil {
				util.CheckErr(err)
			}
			oldFunc(cmd, args)
		}
	default:
		kubectlCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := prePreRun(ctx); err != nil {
				return err
			}
			return nil
		}
	}
}

func prePreRun(ctx context.Context) error {
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %w", err)
	}

	return setupKubeconfig(ctx)
}

func setupKubeconfig(ctx context.Context) error {
	var kubeConfigPath string
	if *commonCmdData.KubeConfigBase64 != "" {
		var err error
		kubeConfigPath, err = tmp_manager.CreateKubeConfigFromBase64(ctx, strings.NewReader(*commonCmdData.KubeConfigBase64))
		if err != nil {
			return fmt.Errorf("unable to create kubeconfig from base64: %w", err)
		}
		*configFlags.KubeConfig = ""
	}

	if kubeConfigEnv := common.GetFirstExistingKubeConfigEnvVar(); kubeConfigEnv != "" && *configFlags.KubeConfig == "" {
		kubeConfigPath = kubeConfigEnv
	}

	if kubeConfigPath != "" {
		if err := os.Setenv("KUBECONFIG", kubeConfigPath); err != nil {
			return fmt.Errorf("unable to set $KUBECONFIG env var: %w", err)
		}
	}

	return nil
}
