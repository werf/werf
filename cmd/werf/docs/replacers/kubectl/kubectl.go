package kubectl

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/werf/werf/cmd/werf/common"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	flagAuthProvider    = "auth-provider"
	flagAuthProviderArg = "auth-provider-arg"

	flagExecCommand    = "exec-command"
	flagExecAPIVersion = "exec-api-version"
	flagExecArg        = "exec-arg"
	flagExecEnv        = "exec-env"
)

func ReplaceKubectlDocs(cmd *cobra.Command) *cobra.Command {
	if len(cmd.Commands()) > 0 {
		for _, c := range cmd.Commands() {
			if len(c.Commands()) > 0 {
				ReplaceKubectlDocs(c)
			}
			setNewDocs(c)
		}
	}
	setNewDocs(cmd)
	return cmd
}

func setNewDocs(cmd *cobra.Command) {
	switch cmd.Use {
	case "events [--for TYPE/NAME] [--watch]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetAlphaEventsDocs().LongMD,
		}
	case "kubectl":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetKubectlDocs().LongMD,
		}
	case "alpha":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetAlphaDocs().LongMD,
		}
	case "annotate [--overwrite] (-f FILENAME | TYPE NAME) KEY_1=VAL_1 ... KEY_N=VAL_N [--resource-version=version]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetAnnotateDocs().LongMD,
		}
	case "api-resources":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetApiResourcesDocs().LongMD,
		}
	case "api-versions":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetApiVersionsDocs().LongMD,
		}
	case "apply (-f FILENAME | -k DIRECTORY)":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetApplyDocs().LongMD,
		}
	case "edit-last-applied (RESOURCE/NAME | -f FILENAME)":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetApplyEditLastAppliedDocs().LongMD,
		}
	case "set-last-applied -f FILENAME":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetApplySetLastAppliedDocs().LongMD,
		}
	case "view-last-applied (TYPE [NAME | -l label] | TYPE/NAME | -f FILENAME)":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetApplyViewLastAppliedDocs().LongMD,
		}
	case "attach (POD | TYPE/NAME) -c CONTAINER":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetAttachDocs().LongMD,
		}
	case "auth":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetAuthDocs().LongMD,
		}
	case "can-i VERB [TYPE | TYPE/NAME | NONRESOURCEURL]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetAuthCanIDocs().LongMD,
		}
	case "reconcile -f FILENAME":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetAuthReconcileDocs().LongMD,
		}
	case "autoscale (-f FILENAME | TYPE NAME | TYPE/NAME) [--min=MINPODS] --max=MAXPODS [--cpu-percent=CPU]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetAutoscaleDocs().LongMD,
		}
	case "certificate SUBCOMMAND":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetCertificateDocs().LongMD,
		}
	case "approve (-f FILENAME | NAME)":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetCertificateApproveDocs().LongMD,
		}
	case "deny (-f FILENAME | NAME)":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetCertificateDenyDocs().LongMD,
		}
	case "cluster-info":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetClusterInfoDocs().LongMD,
		}
	case "dump":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetClusterInfoDumpDocs().LongMD,
		}
	case "completion SHELL":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetCompletionDocs().LongMD,
		}
	case "config SUBCOMMAND":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetConfigDocs(clientcmd.NewDefaultPathOptions()).LongMD,
		}
	case "current-context":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetConfigCurrentContextDocs().LongMD,
		}
	case "delete-cluster NAME":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetConfigDeleteClusterDocs().LongMD,
		}
	case "delete-context NAME":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetConfigDeleteContextDocs().LongMD,
		}
	case "delete-user NAME":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetConfigDeleteUserDocs().LongMD,
		}
	case "get-clusters":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetConfigGetClustersDocs().LongMD,
		}
	case "get-contexts [(-o|--output=)name)]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetConfigGetContextsDocs().LongMD,
		}
	case "get-users":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetConfigGetUsersDocs().LongMD,
		}
	case "rename-context CONTEXT_NAME NEW_NAME":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetConfigRenameContextDocs().LongMD,
		}
	case "set PROPERTY_NAME PROPERTY_VALUE":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetConfigSetDocs().LongMD,
		}
	case fmt.Sprintf("set-cluster NAME [--%v=server] [--%v=path/to/certificate/authority] "+
		"[--%v=true] [--%v=example.com]", clientcmd.FlagAPIServer, clientcmd.FlagCAFile,
		clientcmd.FlagInsecure, clientcmd.FlagTLSServerName):
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetConfigSetClusterDocs().LongMD,
		}
	case fmt.Sprintf("set-context [NAME | --current] [--%v=cluster_nickname] [--%v=user_nickname] "+
		"[--%v=namespace]", clientcmd.FlagClusterName, clientcmd.FlagAuthInfoName, clientcmd.FlagNamespace):
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetConfigSetContextDocs().LongMD,
		}
	case fmt.Sprintf(
		"set-credentials NAME [--%v=path/to/certfile] "+
			"[--%v=path/to/keyfile] "+
			"[--%v=bearer_token] "+
			"[--%v=basic_user] "+
			"[--%v=basic_password] "+
			"[--%v=provider_name] "+
			"[--%v=key=value] "+
			"[--%v=exec_command] "+
			"[--%v=exec_api_version] "+
			"[--%v=arg] "+
			"[--%v=key=value]",
		clientcmd.FlagCertFile,
		clientcmd.FlagKeyFile,
		clientcmd.FlagBearerToken,
		clientcmd.FlagUsername,
		clientcmd.FlagPassword,
		flagAuthProvider,
		flagAuthProviderArg,
		flagExecCommand,
		flagExecAPIVersion,
		flagExecArg,
		flagExecEnv,
	):
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetConfigSetCredentialsDocs().LongMD,
		}
	case "unset PROPERTY_NAME":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetConfigUnsetDocs().LongMD,
		}
	case "use-context CONTEXT_NAME":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetConfigUseContextDocs().LongMD,
		}
	case "view":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetConfigViewDocs().LongMD,
		}
	}
}
