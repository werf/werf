package kubectl

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/cmd"
	"k8s.io/kubectl/pkg/cmd/plugin"

	"github.com/werf/werf/cmd/werf/common"
)

var (
	configFlags *genericclioptions.ConfigFlags

	result, textsResult      bool
	msgAnnotations, msgTexts string
)

func TestReplaceKubectlDocs(t *testing.T) {
	configFlags = genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	cmd := ReplaceKubectlDocs(cmd.NewDefaultKubectlCommandWithArgs(cmd.KubectlOptions{
		PluginHandler: cmd.NewDefaultPluginHandler(plugin.ValidPluginFilenamePrefixes),
		Arguments:     os.Args,
		ConfigFlags:   configFlags,
		IOStreams:     genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr},
	}))

	msgAnnotations = "There are no annotations in the following commands:\n\n"
	msgTexts = "The following commands do not match the text in the annotation:\n\n"
	result = true
	textsResult = true
	checkCmd(cmd, "")
	if !result || !textsResult {
		if !result {
			t.Error(msgAnnotations + "\n\n")
		}
		if !textsResult {
			t.Error(msgTexts + "\n\n")
		}
	}
}

func checkCmd(cmd *cobra.Command, previuosCmd string) {
	if ann, ok := cmd.Annotations[common.DocsLongMD]; !ok {
		result = false
		msgAnnotations += previuosCmd + " " + cmd.Use + "\n\n"
	} else {
		if !checkText(ann) {
			textsResult = false
			msgTexts += previuosCmd + " " + cmd.Use + "\n\n"
		}
	}
	if len(cmd.Commands()) > 0 {
		for _, c := range cmd.Commands() {
			checkCmd(c, previuosCmd+" "+cmd.Use)
		}
	}
}

func checkText(ann string) bool {
	if ann != GetAlphaEventsDocs().LongMD &&
		ann != GetKubectlDocs().LongMD &&
		ann != GetAlphaDocs().LongMD &&
		ann != GetAnnotateDocs().LongMD &&
		ann != GetApiResourcesDocs().LongMD &&
		ann != GetApiVersionsDocs().LongMD &&
		ann != GetApplyDocs().LongMD &&
		ann != GetApplyEditLastAppliedDocs().LongMD &&
		ann != GetApplySetLastAppliedDocs().LongMD &&
		ann != GetApplyViewLastAppliedDocs().LongMD &&
		ann != GetAttachDocs().LongMD &&
		ann != GetAuthDocs().LongMD &&
		ann != GetWhoamiDocs().LongMD &&
		ann != GetAuthCanIDocs().LongMD &&
		ann != GetAuthReconcileDocs().LongMD &&
		ann != GetAutoscaleDocs().LongMD &&
		ann != GetCertificateDocs().LongMD &&
		ann != GetCertificateApproveDocs().LongMD &&
		ann != GetCertificateDenyDocs().LongMD &&
		ann != GetClusterInfoDocs().LongMD &&
		ann != GetClusterInfoDumpDocs().LongMD &&
		ann != GetCompletionDocs().LongMD &&
		ann != GetConfigDocs(clientcmd.NewDefaultPathOptions()).LongMD &&
		ann != GetConfigCurrentContextDocs().LongMD &&
		ann != GetConfigDeleteClusterDocs().LongMD &&
		ann != GetConfigDeleteContextDocs().LongMD &&
		ann != GetConfigDeleteUserDocs().LongMD &&
		ann != GetConfigGetClustersDocs().LongMD &&
		ann != GetConfigGetContextsDocs().LongMD &&
		ann != GetConfigGetUsersDocs().LongMD &&
		ann != GetConfigRenameContextDocs().LongMD &&
		ann != GetConfigSetDocs().LongMD &&
		ann != GetConfigSetClusterDocs().LongMD &&
		ann != GetConfigSetContextDocs().LongMD &&
		ann != GetConfigSetCredentialsDocs().LongMD &&
		ann != GetConfigUnsetDocs().LongMD &&
		ann != GetConfigUseContextDocs().LongMD &&
		ann != GetConfigViewDocs().LongMD &&
		ann != GetCordonDocs().LongMD &&
		ann != GetCpDocs().LongMD &&
		ann != GetCreateDocs().LongMD &&
		ann != GetCreateClusterRoleDocs().LongMD &&
		ann != GetCreateClusterRoleBindingDocs().LongMD &&
		ann != GetCreateConfigMapDocs().LongMD &&
		ann != GetCreateCronJobDocs().LongMD &&
		ann != GetCreateDeploymentDocs().LongMD &&
		ann != GetCreateIngressDocs().LongMD &&
		ann != GetCreateJobDocs().LongMD &&
		ann != GetCreateNamespaceDocs().LongMD &&
		ann != GetCreatePodDisruptionBudgetDocs().LongMD &&
		ann != GetCreatePriorityClassDocs().LongMD &&
		ann != GetCreateQuotaDocs().LongMD &&
		ann != GetCreateRoleDocs().LongMD &&
		ann != GetCreateRoleBindingDocs().LongMD &&
		ann != GetCreateSecretDocs().LongMD &&
		ann != GetCreateSecretDockerRegistryDocs().LongMD &&
		ann != GetCreateSecretGenericDocs().LongMD &&
		ann != GetCreateSecretTLSDocs().LongMD &&
		ann != GetCreateServiceDocs().LongMD &&
		ann != GetCreateServiceClusterIPDocs().LongMD &&
		ann != GetCreateServiceExternalNameDocs().LongMD &&
		ann != GetCreateServiceLoadBalancerDocs().LongMD &&
		ann != GetCreateServiceNodePortDocs().LongMD &&
		ann != GetCreateServiceAccountDocs().LongMD &&
		ann != GetCreateTokenDocs().LongMD &&
		ann != GetDebugDocs().LongMD &&
		ann != GetDeleteDocs().LongMD &&
		ann != GetDescribeDocs().LongMD &&
		ann != GetDiffDocs().LongMD &&
		ann != GetDrainDocs().LongMD &&
		ann != GetEditDocs().LongMD &&
		ann != GetExecDocs().LongMD &&
		ann != GetExplainDocs().LongMD &&
		ann != GetExposeDocs().LongMD &&
		ann != GetGetDocs().LongMD &&
		ann != GetKustomizeDocs().LongMD &&
		ann != GetLabelDocs().LongMD &&
		ann != GetLogsDocs().LongMD &&
		ann != GetOptionsDocs().LongMD &&
		ann != GetPatchDocs().LongMD &&
		ann != GetPluginDocs().LongMD &&
		ann != GetPluginListDocs().LongMD &&
		ann != GetPortForwardDocs().LongMD &&
		ann != GetProxyDocs().LongMD &&
		ann != GetReplaceDocs().LongMD &&
		ann != GetRolloutDocs().LongMD &&
		ann != GetRolloutHistoryDocs().LongMD &&
		ann != GetRolloutPauseDocs().LongMD &&
		ann != GetRolloutResumeDocs().LongMD &&
		ann != GetRolloutUndoDocs().LongMD &&
		ann != GetRolloutStatusDocs().LongMD &&
		ann != GetRolloutRestartDocs().LongMD &&
		ann != GetRunDocs().LongMD &&
		ann != GetScaleDocs().LongMD &&
		ann != GetSetDocs().LongMD &&
		ann != GetSetImageDocs().LongMD &&
		ann != GetSetResourceDocs().LongMD &&
		ann != GetSetSelectorDocs().LongMD &&
		ann != GetSetSubjectDocs().LongMD &&
		ann != GetSetServiceAccountDocs().LongMD &&
		ann != GetSetEnvDocs().LongMD &&
		ann != GetTaintDocs().LongMD &&
		ann != GetTopDocs().LongMD &&
		ann != GetTopNodeDocs().LongMD &&
		ann != GetTopPodDocs().LongMD &&
		ann != GetUncordonDocs().LongMD &&
		ann != GetVersionDocs().LongMD &&
		ann != GetEventsDocs().LongMD &&
		ann != GetWaitDocs().LongMD {
		return false
	}
	return true
}
