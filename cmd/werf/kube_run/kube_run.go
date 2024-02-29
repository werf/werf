package kube_run

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/containers/image/v5/docker/reference"
	config2 "github.com/containers/image/v5/pkg/docker/config"
	imgtypes "github.com/containers/image/v5/types"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/kubectl/pkg/scheme"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/logboek"
	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/build"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/config/deploy_params"
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/deploy/helm"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/git_repo/gitdata"
	"github.com/werf/werf/pkg/giterminism_manager"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/logging"
	"github.com/werf/werf/pkg/ssh_agent"
	"github.com/werf/werf/pkg/storage/lrumeta"
	"github.com/werf/werf/pkg/storage/manager"
	"github.com/werf/werf/pkg/tmp_manager"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
	"github.com/werf/werf/pkg/werf/global_warnings"
)

type cmdDataType struct {
	Interactive     bool
	AllocateTty     bool
	Rm              bool
	RmWithNamespace bool
	AutoPullSecret  bool

	Pod             string
	Command         []string
	ImageName       string
	Overrides       string
	RunExtraOptions string
	CopyFrom        []string
	CopyTo          []string

	registryCredsFound bool
}

type copyFromTo struct {
	Src string
	Dst string
}

var (
	cmdData       cmdDataType
	commonCmdData common.CmdData
)

type dockerConfigJson struct {
	Auths map[string]dockerAuthJson `json:"auths"`
}

type dockerAuthJson struct {
	Auth          string `json:"auth,omitempty"`
	IdentityToken string `json:"identitytoken,omitempty"`
}

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "kube-run [options] [IMAGE_NAME] [-- COMMAND ARG...]",
		Short:                 "Run container for project image in Kubernetes",
		Long:                  common.GetLongCommandDescription(GetKubeRunDocs().Long),
		DisableFlagsInUseLine: true,
		Example: `  # Run interactive shell in the image
  $ werf kube-run --repo test/test -it -- sh

  # Run image with specified command
  $ werf kube-run --repo test/test application -- /app/run.sh

  # Run multiple commands
  $ werf kube-run --repo test/test application -- sh -euc 'test -d /tmp && touch /tmp/file'
`,
		Annotations: map[string]string{
			common.DisableOptionsInUseLineAnno: "1",
			common.DocsLongMD:                  GetKubeRunDocs().LongMD,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			defer global_warnings.PrintGlobalWarnings(ctx)

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			if err := processArgs(cmd, args); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			if cmdData.RmWithNamespace && !cmdData.Rm {
				return fmt.Errorf("option --rm-with-namespace requires --rm to be set")
			}

			if cmdData.AllocateTty && !cmdData.Interactive {
				return fmt.Errorf("option --tty requires --interactive to be set")
			}

			if cmdData.Pod != "" {
				if errMsgs := validation.IsDNS1123Subdomain(cmdData.Pod); len(errMsgs) > 0 {
					return fmt.Errorf("--pod name is not a valid subdomain:\n%s", strings.Join(errMsgs, "\n"))
				}
			}

			if *commonCmdData.Follow {
				if cmdData.Interactive || cmdData.AllocateTty {
					return fmt.Errorf("--follow mode does not work with -i or -t options")
				}
			}

			if err := validateCopyFrom(); err != nil {
				return fmt.Errorf("error validating --copy-from: %w", err)
			}

			if err := validateCopyTo(); err != nil {
				return fmt.Errorf("error validating --copy-to: %w", err)
			}

			return runMain(ctx)
		},
	})

	common.SetupDir(&commonCmdData, cmd)
	common.SetupGitWorkTree(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupGiterminismConfigPath(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)
	common.SetupNamespace(&commonCmdData, cmd)
	common.SetupAddAnnotations(&commonCmdData, cmd)
	common.SetupAddLabels(&commonCmdData, cmd)

	common.SetupGiterminismOptions(&commonCmdData, cmd)

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})
	common.SetupSSHKey(&commonCmdData, cmd)

	common.SetupSecondaryStagesStorageOptions(&commonCmdData, cmd)
	common.SetupCacheStagesStorageOptions(&commonCmdData, cmd)
	common.SetupRepoOptions(&commonCmdData, cmd, common.RepoDataOptions{})
	common.SetupFinalRepo(&commonCmdData, cmd)

	common.SetupSkipBuild(&commonCmdData, cmd)
	common.SetupRequireBuiltImages(&commonCmdData, cmd)

	common.SetupFollow(&commonCmdData, cmd)

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read and pull images from the specified repo")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupSynchronization(&commonCmdData, cmd)
	common.SetupKubeConfig(&commonCmdData, cmd)
	common.SetupKubeConfigBase64(&commonCmdData, cmd)
	common.SetupKubeContext(&commonCmdData, cmd)

	common.SetupDryRun(&commonCmdData, cmd)

	common.SetupVirtualMerge(&commonCmdData, cmd)

	commonCmdData.SetupPlatform(cmd)

	cmd.Flags().StringVarP(&cmdData.Pod, "pod", "", os.Getenv("WERF_POD"), "Set created pod name (default $WERF_POD or autogenerated if not specified)")
	cmd.Flags().StringVarP(&cmdData.Overrides, "overrides", "", os.Getenv("WERF_OVERRIDES"), "Inline JSON to override/extend any fields in created Pod, e.g. to add imagePullSecrets field (default $WERF_OVERRIDES). %pod_name% and %container_name% will be replaced with names of a created pod and a container.")
	cmd.Flags().StringVarP(&cmdData.RunExtraOptions, "extra-options", "", os.Getenv("WERF_EXTRA_OPTIONS"), "Pass extra options to \"kubectl run\" command, which will create a Pod (default $WERF_EXTRA_OPTIONS)")
	cmd.Flags().BoolVarP(&cmdData.Rm, "rm", "", util.GetBoolEnvironmentDefaultTrue("WERF_RM"), "Remove pod and other created resources after command completion (default $WERF_RM or true if not specified)")
	cmd.Flags().BoolVarP(&cmdData.RmWithNamespace, "rm-with-namespace", "", util.GetBoolEnvironmentDefaultFalse("WERF_RM_WITH_NAMESPACE"), "Remove also a namespace after command completion (default $WERF_RM_WITH_NAMESPACE or false if not specified)")
	cmd.Flags().BoolVarP(&cmdData.Interactive, "interactive", "i", util.GetBoolEnvironmentDefaultFalse("WERF_INTERACTIVE"), "Enable interactive mode (default $WERF_INTERACTIVE or false if not specified)")
	cmd.Flags().BoolVarP(&cmdData.AllocateTty, "tty", "t", util.GetBoolEnvironmentDefaultFalse("WERF_TTY"), "Allocate a TTY (default $WERF_TTY or false if not specified)")
	cmd.Flags().BoolVarP(&cmdData.AutoPullSecret, "auto-pull-secret", "", util.GetBoolEnvironmentDefaultTrue("WERF_AUTO_PULL_SECRET"), "Automatically create docker config secret in the namespace and plug it via pod's imagePullSecrets for private registry access (default $WERF_AUTO_PULL_SECRET or true if not specified)")
	cmd.Flags().StringArrayVarP(&cmdData.CopyFrom, "copy-from", "", []string{}, "Copy file/dir from container to local machine after user command execution. Example: \"/from/file:to\". Can be specified multiple times. Can also be defined with \"$WERF_COPY_FROM_*\", e.g. \"WERF_COPY_FROM_1=from:to\".")
	cmd.Flags().StringArrayVarP(&cmdData.CopyTo, "copy-to", "", []string{}, "Copy file/dir from local machine to container before user command execution. Example: \"from:/to/file\". Can be specified multiple times. Can also be defined with \"$WERF_COPY_TO_*\", e.g. \"WERF_COPY_TO_1=from:to\".")

	return cmd
}

func processArgs(cmd *cobra.Command, args []string) error {
	doubleDashInd := cmd.ArgsLenAtDash()
	doubleDashExist := cmd.ArgsLenAtDash() != -1

	if !doubleDashExist {
		return fmt.Errorf("-- <command> should be specified")
	}

	if doubleDashInd == len(args) {
		return fmt.Errorf("unsupported position args format")
	}

	switch doubleDashInd {
	case 0:
		cmdData.Command = args[doubleDashInd:]
	case 1:
		cmdData.ImageName = args[0]
		cmdData.Command = args[doubleDashInd:]
	default:
		return fmt.Errorf("unsupported position args format")
	}

	return nil
}

func runMain(ctx context.Context) error {
	global_warnings.PostponeMultiwerfNotUpToDateWarning()

	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %w", err)
	}

	containerBackend, processCtx, err := common.InitProcessContainerBackend(ctx, &commonCmdData)
	if err != nil {
		return err
	}
	ctx = processCtx

	gitDataManager, err := gitdata.GetHostGitDataManager(ctx)
	if err != nil {
		return fmt.Errorf("error getting host git data manager: %w", err)
	}

	if err := git_repo.Init(gitDataManager); err != nil {
		return err
	}

	if err := image.Init(); err != nil {
		return err
	}

	if err := lrumeta.Init(); err != nil {
		return err
	}

	if err := true_git.Init(ctx, true_git.Options{LiveGitOutput: *commonCmdData.LogDebug}); err != nil {
		return err
	}

	if err := common.DockerRegistryInit(ctx, &commonCmdData); err != nil {
		return err
	}

	if err := ssh_agent.Init(ctx, common.GetSSHKey(&commonCmdData)); err != nil {
		return fmt.Errorf("cannot initialize ssh agent: %w", err)
	}
	defer func() {
		if err := ssh_agent.Terminate(); err != nil {
			logboek.Context(ctx).Warn().LogF("WARNING: ssh agent termination failed: %s\n", err)
		}
	}()

	giterminismManager, err := common.GetGiterminismManager(ctx, &commonCmdData)
	if err != nil {
		return err
	}

	common.ProcessLogProjectDir(&commonCmdData, giterminismManager.ProjectDir())

	var pod string
	if cmdData.Pod == "" {
		pod = fmt.Sprintf("werf-run-%d", rand.Int())
	} else {
		pod = cmdData.Pod
	}
	secret := pod

	cmdData.Overrides = templateOverrides(cmdData.Overrides, pod, pod)

	_, werfConfig, err := common.GetRequiredWerfConfig(ctx, &commonCmdData, giterminismManager, common.GetWerfConfigOptions(&commonCmdData, false))
	if err != nil {
		return fmt.Errorf("unable to load werf config: %w", err)
	}

	common.SetupOndemandKubeInitializer(*commonCmdData.KubeContext, *commonCmdData.KubeConfig, *commonCmdData.KubeConfigBase64, *commonCmdData.KubeConfigPathMergeList)
	if err := common.GetOndemandKubeInitializer().Init(ctx); err != nil {
		return err
	}

	namespace, err := deploy_params.GetKubernetesNamespace(*commonCmdData.Namespace, *commonCmdData.Environment, werfConfig)
	if err != nil {
		return err
	}

	defer func() {
		cleanupResources(ctx, pod, secret, namespace)
	}()

	if *commonCmdData.Follow {
		return common.FollowGitHead(ctx, &commonCmdData, func(ctx context.Context, headCommitGiterminismManager giterminism_manager.Interface) error {
			cleanupResources(ctx, pod, secret, namespace)

			if err := run(ctx, pod, secret, namespace, werfConfig, containerBackend, giterminismManager); err != nil {
				return err
			}

			cleanupResources(ctx, pod, secret, namespace)

			return nil
		})
	} else {
		if err := run(ctx, pod, secret, namespace, werfConfig, containerBackend, giterminismManager); err != nil {
			return err
		}
	}

	return nil
}

func run(ctx context.Context, pod, secret, namespace string, werfConfig *config.WerfConfig, containerBackend container_backend.ContainerBackend, giterminismManager giterminism_manager.Interface) error {
	projectName := werfConfig.Meta.Project

	userExtraAnnotations, err := common.GetUserExtraAnnotations(&commonCmdData)
	if err != nil {
		return err
	}
	userExtraAnnotations = util.MergeMaps(userExtraAnnotations, helm.WerfRuntimeAnnotations)

	userExtraLabels, err := common.GetUserExtraLabels(&commonCmdData)
	if err != nil {
		return err
	}
	userExtraLabels = util.MergeMaps(userExtraLabels, helm.WerfRuntimeLabels)

	projectTmpDir, err := tmp_manager.CreateProjectDir(ctx)
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %w", err)
	}
	defer tmp_manager.ReleaseProjectDir(projectTmpDir)

	imageName := cmdData.ImageName
	if imageName == "" && len(werfConfig.GetAllImages()) == 1 {
		imageName = werfConfig.GetAllImages()[0].GetName()
	}

	if !werfConfig.HasImage(imageName) {
		return fmt.Errorf("image %q is not defined in werf.yaml", logging.ImageLogName(imageName, false))
	}

	stagesStorage, err := common.GetStagesStorage(ctx, containerBackend, &commonCmdData)
	if err != nil {
		return err
	}
	finalStagesStorage, err := common.GetOptionalFinalStagesStorage(ctx, containerBackend, &commonCmdData)
	if err != nil {
		return err
	}
	synchronization, err := common.GetSynchronization(ctx, &commonCmdData, projectName, stagesStorage)
	if err != nil {
		return err
	}
	storageLockManager, err := common.GetStorageLockManager(ctx, synchronization)
	if err != nil {
		return err
	}
	secondaryStagesStorageList, err := common.GetSecondaryStagesStorageList(ctx, stagesStorage, containerBackend, &commonCmdData)
	if err != nil {
		return err
	}
	cacheStagesStorageList, err := common.GetCacheStagesStorageList(ctx, containerBackend, &commonCmdData)
	if err != nil {
		return err
	}

	storageManager := manager.NewStorageManager(projectName, stagesStorage, finalStagesStorage, secondaryStagesStorageList, cacheStagesStorageList, storageLockManager)

	logboek.Context(ctx).Info().LogOptionalLn()

	imagesToProcess := build.NewImagesToProcess([]string{imageName}, false)

	conveyorOptions, err := common.GetConveyorOptions(ctx, &commonCmdData, imagesToProcess)
	if err != nil {
		return err
	}

	conveyorWithRetry := build.NewConveyorWithRetryWrapper(werfConfig, giterminismManager, giterminismManager.ProjectDir(), projectTmpDir, ssh_agent.SSHAuthSock, containerBackend, storageManager, storageLockManager, conveyorOptions)
	defer conveyorWithRetry.Terminate()

	var image string
	if err := conveyorWithRetry.WithRetryBlock(ctx, func(c *build.Conveyor) error {
		if common.GetRequireBuiltImages(ctx, &commonCmdData) {
			if err := c.ShouldBeBuilt(ctx, build.ShouldBeBuiltOptions{}); err != nil {
				return err
			}
		} else {
			if err := c.Build(ctx, build.BuildOptions{}); err != nil {
				return err
			}
		}

		image, err = c.GetFullImageName(ctx, imageName)
		if err != nil {
			return fmt.Errorf("unable to get full name for image %q: %w", imageName, err)
		}
		return nil
	}); err != nil {
		return err
	}

	var dockerAuthConf imgtypes.DockerAuthConfig
	var namedRef reference.Named
	if cmdData.AutoPullSecret {
		var err error
		namedRef, dockerAuthConf, err = getDockerConfigCredentials(image)
		if err != nil {
			return fmt.Errorf("unable to get docker config credentials: %w", err)
		}

		if dockerAuthConf == (imgtypes.DockerAuthConfig{}) {
			logboek.Context(ctx).Debug().LogF("No credentials for werf repo found in Docker's config.json. No image pull secret will be created.\n")
		} else {
			cmdData.registryCredsFound = true
		}
	}

	commonArgs, err := createCommonKubectlArgs(namespace)
	if err != nil {
		return fmt.Errorf("error creating common kubectl args: %w", err)
	}

	if err := createNamespace(ctx, namespace); err != nil {
		return fmt.Errorf("unable to create namespace: %w", err)
	}

	if err := createDockerRegistrySecret(ctx, secret, namespace, namedRef, dockerAuthConf); err != nil {
		return fmt.Errorf("unable to create docker registry secret: %w", err)
	}

	return logboek.Streams().DoErrorWithoutProxyStreamDataFormatting(func() error {
		return common.WithoutTerminationSignalsTrap(func() error {
			if err := createPod(ctx, namespace, pod, image, secret, commonArgs, userExtraAnnotations, userExtraLabels); err != nil {
				return fmt.Errorf("error creating Pod: %w", err)
			}

			if err := waitPodReadiness(ctx, namespace, pod, commonArgs); err != nil {
				return fmt.Errorf("error waiting for Pod readiness: %w", err)
			}

			defer stopContainer(ctx, namespace, pod, pod, commonArgs)

			for _, copyTo := range getCopyTo() {
				if err := copyToPod(ctx, namespace, pod, pod, copyTo, commonArgs); err != nil {
					return fmt.Errorf("error copying to Pod: %w", err)
				}
			}

			defer func() {
				for _, copyFrom := range getCopyFrom() {
					copyFromPod(ctx, namespace, pod, pod, copyFrom, commonArgs)
				}
			}()

			if err := execCommandInPod(ctx, namespace, pod, pod, cmdData.Command, commonArgs); err != nil {
				return fmt.Errorf("error running command in Pod: %w", err)
			}

			return nil
		})
	})
}

func createCommonKubectlArgs(namespace string) ([]string, error) {
	commonArgs := []string{
		"--namespace", namespace,
	}

	if *commonCmdData.KubeContext != "" {
		commonArgs = append(commonArgs, "--context", *commonCmdData.KubeContext)
	}

	if *commonCmdData.KubeConfigBase64 != "" {
		commonArgs = append(commonArgs, "--kube-config-base64", *commonCmdData.KubeConfigBase64)
	} else if *commonCmdData.KubeConfig != "" {
		if err := os.Setenv("KUBECONFIG", *commonCmdData.KubeConfig); err != nil {
			return nil, fmt.Errorf("unable to set $KUBECONFIG env var: %w", err)
		}
	} else if len(*commonCmdData.KubeConfigPathMergeList) > 0 {
		if err := os.Setenv("KUBECONFIG", common.GetFirstExistingKubeConfigEnvVar()); err != nil {
			return nil, fmt.Errorf("unable to set $KUBECONFIG env var: %w", err)
		}
	}

	return commonArgs, nil
}

func createPod(ctx context.Context, namespace, pod, image, secret string, extraArgs []string, extraAnnos, extraLabels map[string]string) error {
	logboek.Context(ctx).LogF("Running pod %q ...\n", pod)

	args, err := createKubectlRunArgs(pod, image, secret, extraArgs, extraAnnos, extraLabels)
	if err != nil {
		return fmt.Errorf("error creating kubectl run args: %w", err)
	}

	cmd := util.ExecKubectlCmd(args...)

	if *commonCmdData.DryRun {
		fmt.Println(cmd.String())
		return nil
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running pod: %w", err)
	}

	return nil
}

func createKubectlRunArgs(pod, image, secret string, extraArgs []string, extraAnnos, extraLabels map[string]string) ([]string, error) {
	args := []string{
		"run",
		pod,
		"--image", image,
		"--command",
		"--restart", "Never",
		"--quiet",
		"--pod-running-timeout=6h",
	}

	args = append(args, extraArgs...)

	if overrides, err := generateOverrides(pod, secret, extraAnnos, extraLabels); err != nil {
		return nil, fmt.Errorf("error generating --overrides: %w", err)
	} else if overrides != nil {
		args = append(args, "--overrides", string(overrides), "--override-type", "strategic")
	}

	if cmdData.RunExtraOptions != "" {
		args = append(args, strings.Fields(cmdData.RunExtraOptions)...)
	}

	return args, nil
}

// Can return nil overrides.
func generateOverrides(container, secret string, extraAnnos, extraLabels map[string]string) ([]byte, error) {
	codec := runtime.NewCodec(scheme.DefaultJSONEncoder(), scheme.Codecs.UniversalDeserializer())

	podOverrides := &corev1.Pod{}
	if err := runtime.DecodeInto(codec, []byte(cmdData.Overrides), podOverrides); err != nil {
		return nil, fmt.Errorf("error decoding --overrides: %w", err)
	}

	createMainContainer(podOverrides, container)

	addAnnotations(extraAnnos, podOverrides)
	addLabels(extraLabels, podOverrides)

	if cmdData.AutoPullSecret && cmdData.registryCredsFound {
		if err := addImagePullSecret(secret, podOverrides); err != nil {
			return nil, fmt.Errorf("error adding imagePullSecret to --overrides: %w", err)
		}
	}

	overrides, err := runtime.Encode(codec, podOverrides)
	if err != nil {
		return nil, fmt.Errorf("error encoding generated --overrides: %w", err)
	}

	overrides, err = cleanPodManifest(overrides)
	if err != nil {
		return nil, fmt.Errorf("error cleaning --overrides: %w", err)
	}

	return overrides, nil
}

func createMainContainer(pod *corev1.Pod, container string) {
	if util.FirstMatchInSliceIndex(pod.Spec.Containers, func(i int, val corev1.Container) bool {
		return val.Name == container
	}) == nil {
		pod.Spec.Containers = append(pod.Spec.Containers, corev1.Container{Name: container})
	}

	pod.Spec.Containers[getContainerIndex(pod, container)].Command = []string{
		"sh", "-euc",
	}
	pod.Spec.Containers[getContainerIndex(pod, container)].Args = []string{
		"until [ -f /tmp/werf-kube-run-quit ]; do sleep 1; done",
	}
}

func getContainerIndex(pod *corev1.Pod, container string) int {
	return *util.FirstMatchInSliceIndex(pod.Spec.Containers, func(i int, val corev1.Container) bool {
		return val.Name == container
	})
}

func cleanPodManifest(podJsonManifest []byte) ([]byte, error) {
	podJsonManifest = []byte(strings.TrimSpace(string(podJsonManifest)))

	var pod map[string]interface{}
	if err := json.Unmarshal(podJsonManifest, &pod); err != nil {
		return nil, fmt.Errorf("error unmarshaling pod json manifest: %w", err)
	}

	if pod["spec"].(map[string]interface{})["containers"] != nil {
		return podJsonManifest, nil
	}

	delete(pod["spec"].(map[string]interface{}), "containers")
	if result, err := json.Marshal(pod); err != nil {
		return nil, fmt.Errorf("error marshaling cleaned pod json manifest: %w", err)
	} else {
		return result, nil
	}
}

func waitPodReadiness(ctx context.Context, namespace, pod string, extraArgs []string) error {
	if *commonCmdData.DryRun {
		return nil
	}

	logboek.Context(ctx).LogF("Waiting for pod to be ready ...\n")

	for {
		phase, err := getPodPhase(namespace, pod, extraArgs)
		if err != nil {
			return fmt.Errorf("error getting Pod phase: %w", err)
		}

		switch phase {
		case corev1.PodFailed:
			return fmt.Errorf("pod %s/%s failed", namespace, pod)
		case corev1.PodSucceeded:
			return fmt.Errorf("pod %s/%s stopped too early", namespace, pod)
		case corev1.PodRunning:
			if ready, err := isPodReady(namespace, pod, extraArgs); err != nil {
				return fmt.Errorf("error checking pod readiness: %w", err)
			} else if ready {
				return nil
			} else {
				continue
			}
		default:
			continue
		}
	}
}

func getPodPhase(namespace, pod string, extraArgs []string) (corev1.PodPhase, error) {
	args := []string{
		"get", "pod", "--template", "{{.status.phase}}", pod,
	}

	args = append(args, extraArgs...)

	cmd := util.ExecKubectlCmd(args...)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error getting pod %s/%s spec: %w", namespace, pod, err)
	}

	return corev1.PodPhase(strings.TrimSpace(stdout.String())), nil
}

func isPodReady(namespace, pod string, extraArgs []string) (bool, error) {
	args := []string{
		"get", "pod", "--template", "{{range .status.conditions}}{{if eq .type \"Ready\"}}{{.status}}{{end}}{{end}}", pod,
	}

	args = append(args, extraArgs...)

	cmd := util.ExecKubectlCmd(args...)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("error getting pod %s/%s spec: %w", namespace, pod, err)
	}

	switch strings.TrimSpace(stdout.String()) {
	case "True":
		return true, nil
	default:
		return false, nil
	}
}

func copyFromPod(ctx context.Context, namespace, pod, container string, copyFrom copyFromTo, extraArgs []string) {
	logboek.Context(ctx).LogF("Copying %q from pod to %q ...\n", copyFrom.Src, copyFrom.Dst)

	args := []string{
		"cp", fmt.Sprint(namespace, "/", pod, ":", copyFrom.Src), copyFrom.Dst, "-c", container,
	}

	args = append(args, extraArgs...)

	cmd := util.ExecKubectlCmd(args...)

	if *commonCmdData.DryRun {
		fmt.Println(cmd.String())
		return
	}

	if err := cmd.Run(); err != nil {
		logboek.Context(ctx).Warn().LogF("Error copying %q from pod %s/s: %s\n", copyFrom.Src, namespace, pod, err)
	}
}

func copyToPod(ctx context.Context, namespace, pod, container string, copyFrom copyFromTo, extraArgs []string) error {
	logboek.Context(ctx).LogF("Copying %q to %q in pod ...\n", copyFrom.Src, copyFrom.Dst)

	args := []string{
		"cp", copyFrom.Src, fmt.Sprint(namespace, "/", pod, ":", copyFrom.Dst), "-c", container,
	}

	args = append(args, extraArgs...)

	cmd := util.ExecKubectlCmd(args...)

	if *commonCmdData.DryRun {
		fmt.Println(cmd.String())
		return nil
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error copying %q to pod %s/%s: %w", copyFrom.Src, namespace, pod, err)
	}

	return nil
}

func stopContainer(ctx context.Context, namespace, pod, container string, extraArgs []string) {
	logboek.Context(ctx).LogF("Stopping container %q in pod ...\n", container)

	args := []string{
		"exec", pod, "-q", "--pod-running-timeout", "5h", "-c", container,
	}

	args = append(args, extraArgs...)
	args = append(args, "--", "touch", "/tmp/werf-kube-run-quit")

	cmd := util.ExecKubectlCmd(args...)

	if *commonCmdData.DryRun {
		fmt.Println(cmd.String())
		return
	}

	if err := cmd.Run(); err != nil {
		logboek.Context(ctx).Warn().LogF("Error stopping service container %s/%s/%s for copying files: %s\n", namespace, pod, container, err)
	}
}

func execCommandInPod(ctx context.Context, namespace, pod, container string, command, extraArgs []string) error {
	logboek.Context(ctx).LogF("Execing into pod ...\n")

	args := []string{
		"exec", pod, "-q", "--pod-running-timeout", "5h", "-c", container,
	}

	args = append(args, extraArgs...)

	if cmdData.Interactive {
		args = append(args, "-i")
	}

	if cmdData.AllocateTty {
		args = append(args, "-t")
	}

	args = append(args, "--")
	args = append(args, command...)

	cmd := util.ExecKubectlCmd(args...)

	if *commonCmdData.DryRun {
		fmt.Println(cmd.String())
		return nil
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running command %q in pod %s/%s: %w", cmd, namespace, pod, err)
	}

	return nil
}

func cleanupResources(ctx context.Context, pod, secret, namespace string) {
	if !cmdData.Rm || *commonCmdData.DryRun {
		return
	}

	if isNsExist, err := isNamespaceExist(ctx, namespace); err != nil {
		logboek.Context(ctx).Warn().LogF("WARNING: unable to check for namespace existence: %s\n", err)
		return
	} else if !isNsExist {
		return
	}

	if isPodExist, err := isPodExist(ctx, pod, namespace); err != nil {
		logboek.Context(ctx).Warn().LogF("WARNING: unable to check for pod existence: %s\n", err)
	} else if isPodExist {
		logboek.Context(ctx).LogF("Cleaning up pod %q ...\n", pod)
		if err := kube.Client.CoreV1().Pods(namespace).Delete(ctx, pod, v1.DeleteOptions{}); err != nil {
			logboek.Context(ctx).Warn().LogF("WARNING: pod cleaning up failed: %s\n", err)
		}
	}

	if cmdData.AutoPullSecret && cmdData.registryCredsFound {
		if isSecretExist, err := isSecretExist(ctx, secret, namespace); err != nil {
			logboek.Context(ctx).Warn().LogF("WARNING: unable to check for secret existence: %s\n", err)
		} else if isSecretExist {
			logboek.Context(ctx).LogF("Cleaning up secret %q ...\n", secret)
			if err := kube.Client.CoreV1().Secrets(namespace).Delete(ctx, secret, v1.DeleteOptions{}); err != nil {
				logboek.Context(ctx).Warn().LogF("WARNING: secret cleaning up failed: %s\n", err)
			}
		}
	}

	if cmdData.RmWithNamespace {
		logboek.Context(ctx).LogF("Cleaning up namespace %q ...\n", namespace)
		if err := kube.Client.CoreV1().Namespaces().Delete(ctx, namespace, v1.DeleteOptions{}); err != nil {
			logboek.Context(ctx).Warn().LogF("WARNING: namespace cleaning up failed: %s\n", err)
		}
	}
}

func createNamespace(ctx context.Context, namespace string) error {
	if *commonCmdData.DryRun {
		return nil
	}

	if isNsExist, err := isNamespaceExist(ctx, namespace); err != nil {
		return fmt.Errorf("unable to check for namespace existence: %w", err)
	} else if isNsExist {
		return nil
	}

	logboek.Context(ctx).LogF("Creating namespace %q ...\n", namespace)

	if _, err := kube.Client.CoreV1().Namespaces().Create(
		ctx,
		&corev1.Namespace{
			ObjectMeta: v1.ObjectMeta{
				Name: namespace,
			},
		},
		v1.CreateOptions{},
	); err != nil {
		return fmt.Errorf("error creating namespace %q: %w", namespace, err)
	}

	return nil
}

func createDockerRegistrySecret(ctx context.Context, name, namespace string, ref reference.Named, dockerAuthConf imgtypes.DockerAuthConfig) error {
	if *commonCmdData.DryRun || !cmdData.registryCredsFound {
		return nil
	}

	secret := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: map[string][]byte{},
		Type: corev1.SecretTypeDockerConfigJson,
	}

	var authJson dockerAuthJson
	switch {
	case dockerAuthConf.IdentityToken != "":
		authJson.IdentityToken = dockerAuthConf.IdentityToken
	case dockerAuthConf.Username != "" && dockerAuthConf.Password != "":
		authJson.Auth = base64.StdEncoding.EncodeToString([]byte(dockerAuthConf.Username + ":" + dockerAuthConf.Password))
	default:
		panic("unexpected dockerAuthConf")
	}

	dockerConfJson := &dockerConfigJson{
		Auths: map[string]dockerAuthJson{
			ref.Name(): authJson,
		},
	}

	dockerConf, err := json.Marshal(dockerConfJson)
	if err != nil {
		return fmt.Errorf("unable to marshal docker config json: %w", err)
	}

	secret.Data[corev1.DockerConfigJsonKey] = dockerConf

	logboek.Context(ctx).LogF("Creating secret %q ...\n", name)
	if _, err := kube.Client.CoreV1().Secrets(namespace).Create(ctx, secret, v1.CreateOptions{}); err != nil {
		return fmt.Errorf("error creating secret %s/%s: %w", namespace, secret, err)
	}

	return nil
}

func isNamespaceExist(ctx context.Context, namespace string) (bool, error) {
	if matchedNamespaces, err := kube.Client.CoreV1().Namespaces().List(ctx, v1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("metadata.name", namespace).String(),
	}); err != nil {
		return false, fmt.Errorf("unable to list namespaces: %w", err)
	} else if len(matchedNamespaces.Items) > 0 {
		return true, nil
	}

	return false, nil
}

func isPodExist(ctx context.Context, pod, namespace string) (bool, error) {
	if matchedPods, err := kube.Client.CoreV1().Pods(namespace).List(ctx, v1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("metadata.name", pod).String(),
	}); err != nil {
		return false, fmt.Errorf("unable to list pods: %w", err)
	} else if len(matchedPods.Items) > 0 {
		return true, nil
	}

	return false, nil
}

func isSecretExist(ctx context.Context, secret, namespace string) (bool, error) {
	if matchedSecrets, err := kube.Client.CoreV1().Secrets(namespace).List(ctx, v1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("metadata.name", secret).String(),
	}); err != nil {
		return false, fmt.Errorf("unable to list secrets: %w", err)
	} else if len(matchedSecrets.Items) > 0 {
		return true, nil
	}

	return false, nil
}

// Might return empty DockerAuthConfig.
func getDockerConfigCredentials(ref string) (reference.Named, imgtypes.DockerAuthConfig, error) {
	namedRef, err := reference.ParseNormalizedNamed(ref)
	if err != nil {
		return nil, imgtypes.DockerAuthConfig{}, fmt.Errorf("unable to parse docker config registry reference %q: %w", ref, err)
	}

	sysContext := &imgtypes.SystemContext{}
	if *commonCmdData.DockerConfig != "" {
		sysContext.AuthFilePath = filepath.Join(*commonCmdData.DockerConfig, "config.json")
	}

	dockerAuthConf, err := config2.GetCredentialsForRef(sysContext, namedRef)
	if err != nil {
		return nil, imgtypes.DockerAuthConfig{}, fmt.Errorf("unable to get docker registry creds for ref %q: %w", ref, err)
	}

	return namedRef, dockerAuthConf, nil
}

func addAnnotations(annotations map[string]string, podOverrides *corev1.Pod) {
	podOverrides.Annotations = util.MergeMaps(annotations, podOverrides.Annotations)
}

func addLabels(labels map[string]string, podOverrides *corev1.Pod) {
	podOverrides.Labels = util.MergeMaps(labels, podOverrides.Labels)
}

func addImagePullSecret(secret string, podOverrides *corev1.Pod) error {
	if secret == "" {
		panic("secret name can't be empty")
	}

	for _, imagePullSecret := range podOverrides.Spec.ImagePullSecrets {
		if imagePullSecret.Name == secret {
			return nil
		}
	}

	podOverrides.Spec.ImagePullSecrets = append(podOverrides.Spec.ImagePullSecrets, corev1.LocalObjectReference{Name: secret})

	return nil
}

func templateOverrides(line, podName, containerName string) string {
	result := strings.ReplaceAll(line, "%container_name%", containerName)
	return strings.ReplaceAll(result, "%pod_name%", podName)
}

func validateCopyFrom() error {
	rawCopyFrom := getCopyFromRaw()

	for _, copyFrom := range rawCopyFrom {
		parts := strings.Split(copyFrom, ":")
		if len(parts) != 2 {
			return fmt.Errorf("wrong format: %s", copyFrom)
		}

		src := cleanCopyPodPath(parts[0])
		dst, err := cleanCopyFromLocalPath(parts[1], path.Base(src))
		if err != nil {
			return fmt.Errorf("error cleaning destination path: %w", err)
		}

		if strings.TrimSpace(src) == "" || strings.TrimSpace(dst) == "" {
			return fmt.Errorf("invalid value: %s", copyFrom)
		}

		if !path.IsAbs(src) {
			return fmt.Errorf("invalid value %q: source should be an absolute path", copyFrom)
		}
	}

	return nil
}

func validateCopyTo() error {
	rawCopyTo := getCopyToRaw()

	for _, copyTo := range rawCopyTo {
		parts := strings.Split(copyTo, ":")
		if len(parts) != 2 {
			return fmt.Errorf("wrong format: %s", copyTo)
		}

		src, err := cleanCopyToLocalPath(parts[0])
		if err != nil {
			return fmt.Errorf("error cleaning source path: %w", err)
		}
		dst := cleanCopyPodPath(parts[1])

		if strings.TrimSpace(src) == "" || strings.TrimSpace(dst) == "" {
			return fmt.Errorf("invalid value: %s", copyTo)
		}

		if !path.IsAbs(dst) {
			return fmt.Errorf("invalid value %q: destination should be an absolute path", copyTo)
		}
	}

	return nil
}

func getCopyFrom() []copyFromTo {
	rawCopyFrom := getCopyFromRaw()

	var result []copyFromTo
	for _, rawcf := range rawCopyFrom {
		parts := strings.Split(rawcf, ":")
		src := cleanCopyPodPath(parts[0])
		dst, err := cleanCopyFromLocalPath(parts[1], path.Base(src))
		if err != nil {
			panic("error cleaning destination path")
		}

		cf := copyFromTo{
			Src: src,
			Dst: dst,
		}

		result = append(result, cf)
	}

	return result
}

func getCopyTo() []copyFromTo {
	rawCopyTo := getCopyToRaw()

	var result []copyFromTo
	for _, rawct := range rawCopyTo {
		parts := strings.Split(rawct, ":")
		src, err := cleanCopyToLocalPath(parts[0])
		if err != nil {
			panic("error cleaning source path")
		}
		dst := cleanCopyPodPath(parts[1])

		ct := copyFromTo{
			Src: src,
			Dst: dst,
		}

		result = append(result, ct)
	}

	return result
}

func cleanCopyPodPath(rawPath string) string {
	return filepath.ToSlash(filepath.Clean(rawPath))
}

func cleanCopyFromLocalPath(rawPath, srcBaseName string) (string, error) {
	rawPath = filepath.Clean(rawPath)

	if rawPath == "." {
		rawPath = filepath.Join(".", srcBaseName)
	}

	rawPath = filepath.Clean(util.ExpandPath(rawPath))

	var err error
	rawPath, err = filepath.Abs(rawPath)
	if err != nil {
		return "", fmt.Errorf("error converting path %q to an absolute path: %w", rawPath, err)
	}

	return rawPath, nil
}

func cleanCopyToLocalPath(rawPath string) (string, error) {
	rawPath = filepath.Clean(util.ExpandPath(rawPath))

	var err error
	rawPath, err = filepath.Abs(rawPath)
	if err != nil {
		return "", fmt.Errorf("error converting path %q to an absolute path: %w", rawPath, err)
	}

	return rawPath, nil
}

func getCopyFromRaw() []string {
	return append(util.PredefinedValuesByEnvNamePrefix("WERF_COPY_FROM_"), cmdData.CopyFrom...)
}

func getCopyToRaw() []string {
	return append(util.PredefinedValuesByEnvNamePrefix("WERF_COPY_TO_"), cmdData.CopyTo...)
}
