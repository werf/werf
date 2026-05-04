package common

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/logboek"
)

func SetupScanContextNamespaceOnly(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.ScanContextNamespaceOnly = new(bool)
	cmd.Flags().BoolVarP(cmdData.ScanContextNamespaceOnly, "scan-context-namespace-only", "", util.GetBoolEnvironmentDefaultFalse("WERF_SCAN_CONTEXT_NAMESPACE_ONLY"), "Scan for used images only in namespace linked with context for each available context in kube-config (or only for the context specified with option --kube-context). When disabled will scan all namespaces in all contexts (or only for the context specified with option --kube-context). (Default $WERF_SCAN_CONTEXT_NAMESPACE_ONLY)")
}

func SetupKubeScanNamespaces(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.KubeScanNamespaces = new([]string)
	cmd.Flags().StringArrayVarP(cmdData.KubeScanNamespaces, "kube-scan-namespaces", "", []string{}, "Kubernetes namespaces to scan for each selected context (can specify multiple). Overrides --scan-context-namespace-only when set.")
}

func GetKubernetesContextClients(configPath, configDataBase64 string, configPathMergeList []string, kubeContext, kubeBearerTokenData, kubeBearerTokenPath, apiServerURL, caDataBase64 string, insecure bool) ([]*kube.ContextClient, error) {
	var res []*kube.ContextClient
	if contextClients, err := kube.GetAllContextsClients(kube.GetAllContextsClientsOptions{
		ConfigPath:          configPath,
		ConfigDataBase64:    configDataBase64,
		ConfigPathMergeList: configPathMergeList,
		BearerToken:         kubeBearerTokenData,
		BearerTokenFile:     kubeBearerTokenPath,
		APIServerURL:        apiServerURL,
		CADataBase64:        caDataBase64,
		Insecure:            insecure,
	}); err != nil {
		return nil, err
	} else {
		if kubeContext != "" {
			for _, cc := range contextClients {
				if cc.ContextName == kubeContext {
					res = append(res, cc)
					break
				}
			}

			if len(res) == 0 {
				return nil, fmt.Errorf("cannot find specified kube context %q", kubeContext)
			}
		} else {
			res = contextClients
		}
	}

	for _, contextClient := range res {
		logboek.Debug().LogF("GetKubernetesContextClients -- context %q namespace %q\n", contextClient.ContextName, contextClient.ContextNamespace)
	}

	return res, nil
}

func GetKubernetesNamespacesByContext(cmdData *CmdData, contextClients []*kube.ContextClient) map[string][]string {
	res := map[string][]string{}
	scanNamespaces := []string{}
	if cmdData.KubeScanNamespaces != nil {
		scanNamespaces = *cmdData.KubeScanNamespaces
	}

	for _, contextClient := range contextClients {
		if len(scanNamespaces) > 0 {
			res[contextClient.ContextName] = append([]string(nil), scanNamespaces...)
			continue
		}

		if *cmdData.ScanContextNamespaceOnly {
			if contextClient.ContextNamespace != "" {
				res[contextClient.ContextName] = []string{contextClient.ContextNamespace}
			} else {
				res[contextClient.ContextName] = nil
			}
			continue
		}

		if contextClient.ContextNamespace != "" {
			res[contextClient.ContextName] = []string{contextClient.ContextNamespace}
		} else {
			res[contextClient.ContextName] = nil
		}
	}

	for contextName, namespaces := range res {
		logboek.Debug().LogF("GetKubernetesNamespacesByContext -- context %q namespaces %q\n", contextName, namespaces)
	}

	return res
}
