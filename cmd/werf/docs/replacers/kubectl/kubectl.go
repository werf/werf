package kubectl

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/werf/werf/cmd/werf/common"
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
	case "whoami":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetWhoamiDocs().LongMD,
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
	case "cordon NODE":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetCordonDocs().LongMD,
		}
	case "cp <file-spec-src> <file-spec-dest>":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetCpDocs().LongMD,
		}
	case "create -f FILENAME":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetCreateDocs().LongMD,
		}
	case "token SERVICE_ACCOUNT_NAME":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetCreateTokenDocs().LongMD,
		}
	case "clusterrole NAME --verb=verb --resource=resource.group [--resource-name=resourcename] " +
		"[--dry-run=server|client|none]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetCreateClusterRoleDocs().LongMD,
		}
	case "clusterrolebinding NAME --clusterrole=NAME [--user=username] [--group=groupname] " +
		"[--serviceaccount=namespace:serviceaccountname] [--dry-run=server|client|none]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetCreateClusterRoleBindingDocs().LongMD,
		}
	case "configmap NAME [--from-file=[key=]source] [--from-literal=key1=value1] [--dry-run=server|client|none]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetCreateConfigMapDocs().LongMD,
		}
	case "cronjob NAME --image=image --schedule='0/5 * * * ?' -- [COMMAND] [args...]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetCreateCronJobDocs().LongMD,
		}
	case "deployment NAME --image=image -- [COMMAND] [args...]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetCreateDeploymentDocs().LongMD,
		}
	case "ingress NAME --rule=host/path=service:port[,tls[=secret]] ":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetCreateIngressDocs().LongMD,
		}
	case "job NAME --image=image [--from=cronjob/name] -- [COMMAND] [args...]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetCreateJobDocs().LongMD,
		}
	case "namespace NAME [--dry-run=server|client|none]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetCreateNamespaceDocs().LongMD,
		}
	case "poddisruptionbudget NAME --selector=SELECTOR --min-available=N [--dry-run=server|client|none]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetCreatePodDisruptionBudgetDocs().LongMD,
		}
	case "priorityclass NAME --value=VALUE --global-default=BOOL [--dry-run=server|client|none]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetCreatePriorityClassDocs().LongMD,
		}
	case "quota NAME [--hard=key1=value1,key2=value2] [--scopes=Scope1,Scope2] [--dry-run=server|client|none]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetCreateQuotaDocs().LongMD,
		}
	case "role NAME --verb=verb --resource=resource.group/subresource " +
		"[--resource-name=resourcename] [--dry-run=server|client|none]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetCreateRoleDocs().LongMD,
		}
	case "rolebinding NAME --clusterrole=NAME|--role=NAME [--user=username] [--group=groupname] " +
		"[--serviceaccount=namespace:serviceaccountname] [--dry-run=server|client|none]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetCreateRoleBindingDocs().LongMD,
		}
	case "secret (docker-registry | generic | tls)":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetCreateSecretDocs().LongMD,
		}
	case "docker-registry NAME --docker-username=user --docker-password=password " +
		"--docker-email=email [--docker-server=string] [--from-file=[key=]source] " +
		"[--dry-run=server|client|none]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetCreateSecretDockerRegistryDocs().LongMD,
		}
	case "generic NAME [--type=string] [--from-file=[key=]source] [--from-literal=key1=value1] [--dry-run=server|client|none]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetCreateSecretGenericDocs().LongMD,
		}
	case "tls NAME --cert=path/to/cert/file --key=path/to/key/file [--dry-run=server|client|none]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetCreateSecretTLSDocs().LongMD,
		}
	case "service":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetCreateServiceDocs().LongMD,
		}
	case "clusterip NAME [--tcp=<port>:<targetPort>] [--dry-run=server|client|none]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetCreateServiceClusterIPDocs().LongMD,
		}
	case "externalname NAME --external-name external.name [--dry-run=server|client|none]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetCreateServiceExternalNameDocs().LongMD,
		}
	case "loadbalancer NAME [--tcp=port:targetPort] [--dry-run=server|client|none]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetCreateServiceLoadBalancerDocs().LongMD,
		}
	case "nodeport NAME [--tcp=port:targetPort] [--dry-run=server|client|none]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetCreateServiceNodePortDocs().LongMD,
		}
	case "serviceaccount NAME [--dry-run=server|client|none]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetCreateServiceAccountDocs().LongMD,
		}
	case "debug (POD | TYPE[[.VERSION].GROUP]/NAME) [ -- COMMAND [args...] ]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetDebugDocs().LongMD,
		}
	case "delete ([-f FILENAME] | [-k DIRECTORY] | TYPE [(NAME | -l label | --all)])":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetDeleteDocs().LongMD,
		}
	case "describe (-f FILENAME | TYPE [NAME_PREFIX | -l label] | TYPE/NAME)":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetDescribeDocs().LongMD,
		}
	case "diff -f FILENAME":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetDiffDocs().LongMD,
		}
	case "drain NODE":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetDrainDocs().LongMD,
		}
	case "edit (RESOURCE/NAME | -f FILENAME)":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetEditDocs().LongMD,
		}
	case "exec (POD | TYPE/NAME) [-c CONTAINER] [flags] -- COMMAND [args...]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetExecDocs().LongMD,
		}
	case "explain TYPE [--recursive=FALSE|TRUE] [--api-version=api-version-group] [--output=plaintext|plaintext-openapiv2]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetExplainDocs().LongMD,
		}
	case "expose (-f FILENAME | TYPE NAME) [--port=port] [--protocol=TCP|UDP|SCTP] " +
		"[--target-port=number-or-name] [--name=name] [--external-ip=external-ip-of-service] " +
		"[--type=type]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetExposeDocs().LongMD,
		}
	case "get [(-o|--output=)json|yaml|name|go-template|go-template-file|template|templatefile|jsonpath|" +
		"jsonpath-as-json|jsonpath-file|custom-columns|custom-columns-file|wide] (TYPE[.VERSION][.GROUP] " +
		"[NAME | -l label] | TYPE[.VERSION][.GROUP]/NAME ...) [flags]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetGetDocs().LongMD,
		}
	case "kustomize DIR":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetKustomizeDocs().LongMD,
		}
	case "label [--overwrite] (-f FILENAME | TYPE NAME) KEY_1=VAL_1 ... KEY_N=VAL_N [--resource-version=version]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetLabelDocs().LongMD,
		}
	case "logs [-f] [-p] (POD | TYPE/NAME) [-c CONTAINER]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetLogsDocs().LongMD,
		}
	case "options":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetOptionsDocs().LongMD,
		}
	case "patch (-f FILENAME | TYPE NAME) [-p PATCH|--patch-file FILE]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetPatchDocs().LongMD,
		}
	case "plugin [flags]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetPluginDocs().LongMD,
		}
	case "list":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetPluginListDocs().LongMD,
		}
	case "port-forward TYPE/NAME [options] [LOCAL_PORT:]REMOTE_PORT [...[LOCAL_PORT_N:]REMOTE_PORT_N]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetPortForwardDocs().LongMD,
		}
	case "proxy [--port=PORT] [--www=static-dir] [--www-prefix=prefix] [--api-prefix=prefix]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetProxyDocs().LongMD,
		}
	case "replace -f FILENAME":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetReplaceDocs().LongMD,
		}
	case "rollout SUBCOMMAND":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetRolloutDocs().LongMD,
		}
	case "history (TYPE NAME | TYPE/NAME) [flags]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetRolloutHistoryDocs().LongMD,
		}
	case "pause RESOURCE":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetRolloutPauseDocs().LongMD,
		}
	case "resume RESOURCE":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetRolloutResumeDocs().LongMD,
		}
	case "undo (TYPE NAME | TYPE/NAME) [flags]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetRolloutUndoDocs().LongMD,
		}
	case "status (TYPE NAME | TYPE/NAME) [flags]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetRolloutStatusDocs().LongMD,
		}
	case "restart RESOURCE":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetRolloutRestartDocs().LongMD,
		}
	case "run NAME --image=image [--env=\"key=value\"] [--port=port] [--dry-run=server|client] " +
		"[--overrides=inline-json] [--command] -- [COMMAND] [args...]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetRunDocs().LongMD,
		}
	case "scale [--resource-version=version] [--current-replicas=count] --replicas=COUNT (-f FILENAME | TYPE NAME)":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetScaleDocs().LongMD,
		}
	case "set SUBCOMMAND":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetSetDocs().LongMD,
		}
	case "image (-f FILENAME | TYPE NAME) CONTAINER_NAME_1=CONTAINER_IMAGE_1 ... CONTAINER_NAME_N=CONTAINER_IMAGE_N":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetSetImageDocs().LongMD,
		}
	case "resources (-f FILENAME | TYPE NAME)  ([--limits=LIMITS & --requests=REQUESTS]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetSetResourceDocs().LongMD,
		}
	case "selector (-f FILENAME | TYPE NAME) EXPRESSIONS [--resource-version=version]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetSetSelectorDocs().LongMD,
		}
	case "subject (-f FILENAME | TYPE NAME) [--user=username] [--group=groupname] [--serviceaccount=namespace:serviceaccountname] [--dry-run=server|client|none]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetSetSubjectDocs().LongMD,
		}
	case "serviceaccount (-f FILENAME | TYPE NAME) SERVICE_ACCOUNT":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetSetServiceAccountDocs().LongMD,
		}
	case "env RESOURCE/NAME KEY_1=VAL_1 ... KEY_N=VAL_N":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetSetEnvDocs().LongMD,
		}
	case "taint NODE NAME KEY_1=VAL_1:TAINT_EFFECT_1 ... KEY_N=VAL_N:TAINT_EFFECT_N":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetTaintDocs().LongMD,
		}
	case "top":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetTopDocs().LongMD,
		}
	case "node [NAME | -l label]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetTopNodeDocs().LongMD,
		}
	case "pod [NAME | -l label]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetTopPodDocs().LongMD,
		}
	case "uncordon NODE":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetUncordonDocs().LongMD,
		}
	case "version":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetVersionDocs().LongMD,
		}
	case "wait ([-f FILENAME] | resource.group/resource.name | resource.group [(-l label | --all)]) [--for=delete|--for condition=available|--for=jsonpath='{}'[=value]]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetWaitDocs().LongMD,
		}
	case "events [(-o|--output=)json|yaml|name|go-template|go-template-file|template|templatefile|jsonpath|" +
		"jsonpath-as-json|jsonpath-file] [--for TYPE/NAME] [--watch] [--types=Normal,Warning]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetEventsDocs().LongMD,
		}
	}
}
