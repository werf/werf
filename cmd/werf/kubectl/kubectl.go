package kubectl

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/cmd"
	"k8s.io/kubectl/pkg/cmd/util"

	"github.com/werf/werf/cmd/werf/common"
)

func NewCmd() *cobra.Command {
	kubectlCmd := cmd.NewDefaultKubectlCommand()

	kubeConfigFlag := kubectlCmd.Flag("kubeconfig")
	kubeConfigFlag.Usage = "Path to the kubeconfig file to use for CLI requests (default $WERF_KUBE_CONFIG, or $WERF_KUBECONFIG, or $KUBECONFIG)"
	if kubeConfigEnvVar := common.GetFirstExistingKubeConfigEnvVar(); kubeConfigEnvVar != "" {
		if err := os.Setenv("KUBECONFIG", kubeConfigEnvVar); err != nil {
			util.CheckErr(fmt.Errorf("unable to set $KUBECONFIG env var: %w", err))
		}
	}

	kubeContextFlag := kubectlCmd.Flag("context")
	kubeContextFlag.Usage = "The name of the kubeconfig context to use (default $WERF_KUBE_CONTEXT)"
	if werfKubeContextEnvVar := os.Getenv("WERF_KUBE_CONTEXT"); werfKubeContextEnvVar != "" {
		if err := kubectlCmd.Flags().Set("context", werfKubeContextEnvVar); err != nil {
			util.CheckErr(fmt.Errorf("unable to set context flag for kubectl: %w", err))
		}
	}

	skipTlsVerifyRegistryFlag := kubectlCmd.Flag("insecure-skip-tls-verify")
	skipTlsVerifyRegistryFlag.Usage = "If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure (default $WERF_SKIP_TLS_VERIFY_REGISTRY)"
	if isSkipTlsVerifyRegistry := common.GetBoolEnvironmentDefaultFalse("WERF_SKIP_TLS_VERIFY_REGISTRY"); isSkipTlsVerifyRegistry {
		if err := kubectlCmd.Flags().Set("insecure-skip-tls-verify", "true"); err != nil {
			util.CheckErr(fmt.Errorf("unable to set insecure-skip-tls-verify flag for kubectl: %w", err))
		}
	}

	return kubectlCmd
}
