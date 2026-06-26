package cleaning

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/exec"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/cmd/werf/common"
)

const (
	kubeTokenFilePath     = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	kubeNamespaceFilePath = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
)

type KubeConfigOptions struct {
	Context             string
	ConfigPath          string
	ConfigDataBase64    string
	ConfigPathMergeList []string

	BearerToken     string
	BearerTokenFile string

	APIServerURL string
	Insecure     bool
	CADataBase64 string
}

type KubeConfig struct {
	Config           *rest.Config
	Context          string
	DefaultNamespace string
}

type GetAllContextsClientsOptions struct {
	ConfigPath          string
	ConfigDataBase64    string
	ConfigPathMergeList []string
	BearerToken         string
	BearerTokenFile     string

	APIServerURL string
	Insecure     bool
	CADataBase64 string
}

type ContextClient struct {
	ContextName      string
	ContextNamespace string
	Client           kubernetes.Interface
}

func GetAllContextsClients(opts GetAllContextsClientsOptions) ([]*ContextClient, error) {
	// Try to load contexts from kubeconfig in flags or from ~/.kube/config
	var outOfClusterErr error

	contexts, outOfClusterErr := getOutOfClusterContextsClients(KubeConfigOptions{
		ConfigPath:          opts.ConfigPath,
		ConfigDataBase64:    opts.ConfigDataBase64,
		ConfigPathMergeList: opts.ConfigPathMergeList,
	})
	if len(contexts) > 0 {
		return contexts, nil
	}

	if hasInClusterConfig() {
		contextClient, err := getInClusterContextClient()
		if err != nil {
			return nil, err
		}
		return []*ContextClient{contextClient}, nil
	}

	tokenClient, err := getTokenContextClient(KubeConfigOptions{
		ConfigPath:          opts.ConfigPath,
		ConfigDataBase64:    opts.ConfigDataBase64,
		ConfigPathMergeList: opts.ConfigPathMergeList,
		BearerToken:         opts.BearerToken,
		BearerTokenFile:     opts.BearerTokenFile,
		APIServerURL:        opts.APIServerURL,
		Insecure:            opts.Insecure,
		CADataBase64:        opts.CADataBase64,
	})
	if err != nil {
		return nil, err
	}
	if tokenClient != nil {
		return []*ContextClient{tokenClient}, nil
	}

	if outOfClusterErr != nil {
		return nil, outOfClusterErr
	}

	return nil, nil
}

func makeOutOfClusterClientConfigError(configPath, context string, err error) error {
	baseErrMsg := "out-of-cluster configuration problem"

	if configPath != "" {
		baseErrMsg += fmt.Sprintf(", custom kube config path is %q", configPath)
	}

	if context != "" {
		baseErrMsg += fmt.Sprintf(", custom kube context is %q", context)
	}

	return fmt.Errorf("%s: %w", baseErrMsg, err)
}

func setConfigPathMergeListEnvironment(configPathMergeList []string) error {
	configPathEnvVar := strings.Join(configPathMergeList, string(filepath.ListSeparator))
	if err := os.Setenv(clientcmd.RecommendedConfigPathEnvVar, configPathEnvVar); err != nil {
		return fmt.Errorf("unable to set env var %q: %w", clientcmd.RecommendedConfigPathEnvVar, err)
	}
	return nil
}

func GetClientConfig(context, configPath string, configData []byte, configPathMergeList []string, overrides *clientcmd.ConfigOverrides) (clientcmd.ClientConfig, error) {
	if context != "" {
		overrides.CurrentContext = context
	}

	if configData != nil {
		config, err := clientcmd.Load(configData)
		if err != nil {
			return nil, fmt.Errorf("unable to load config data: %w", err)
		}

		return clientcmd.NewDefaultClientConfig(*config, overrides), nil
	}

	if len(configPathMergeList) > 0 {
		if err := setConfigPathMergeListEnvironment(configPathMergeList); err != nil {
			return nil, err
		}
	}

	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	if configPath != "" {
		rules.ExplicitPath = configPath
	}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides), nil
}

func hasInClusterConfig() bool {
	token, _ := fileExists(kubeTokenFilePath)
	ns, _ := fileExists(kubeNamespaceFilePath)
	return token && ns
}

func parseConfigDataBase64(configDataBase64 string) ([]byte, error) {
	var configData []byte

	if configDataBase64 != "" {
		if data, err := base64.StdEncoding.DecodeString(configDataBase64); err != nil {
			return nil, fmt.Errorf("unable to decode base64 config data: %w", err)
		} else {
			configData = data
		}
	}

	return configData, nil
}

func getOutOfClusterContextsClients(opts KubeConfigOptions) ([]*ContextClient, error) {
	var res []*ContextClient

	configData, err := parseConfigDataBase64(opts.ConfigDataBase64)
	if err != nil {
		return nil, fmt.Errorf("unable to parse base64 config data: %w", err)
	}

	overrides := &clientcmd.ConfigOverrides{
		ClusterDefaults: clientcmd.ClusterDefaults,
		AuthInfo: api.AuthInfo{
			Token:     opts.BearerToken,
			TokenFile: opts.BearerTokenFile,
		},
	}

	clientConfig, err := GetClientConfig(
		"",
		opts.ConfigPath,
		configData,
		opts.ConfigPathMergeList,
		overrides,
	)
	if err != nil {
		return nil, err
	}

	rc, err := clientConfig.RawConfig()
	if err != nil {
		return nil, err
	}

	for contextName, context := range rc.Contexts {
		clientConfig, err := GetClientConfig(
			contextName,
			opts.ConfigPath,
			configData,
			opts.ConfigPathMergeList,
			overrides,
		)
		if err != nil {
			return nil, makeOutOfClusterClientConfigError(opts.ConfigPath, contextName, err)
		}

		config, err := clientConfig.ClientConfig()
		if err != nil {
			return nil, makeOutOfClusterClientConfigError(opts.ConfigPath, contextName, err)
		}

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			return nil, err
		}

		res = append(res, &ContextClient{
			ContextName:      contextName,
			ContextNamespace: context.Namespace,
			Client:           clientset,
		})
	}

	return res, nil
}

func getInClusterConfig() (*KubeConfig, error) {
	res := &KubeConfig{}

	if config, err := rest.InClusterConfig(); err != nil {
		return nil, fmt.Errorf("in-cluster configuration problem: %w", err)
	} else {
		res.Config = config
	}

	if data, err := os.ReadFile(kubeNamespaceFilePath); err != nil {
		return nil, fmt.Errorf("in-cluster configuration problem: cannot determine default kubernetes namespace: error reading %s: %w", kubeNamespaceFilePath, err)
	} else {
		res.DefaultNamespace = string(data)
	}

	return res, nil
}

func getInClusterContextClient() (*ContextClient, error) {
	kubeConfig, err := getInClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(kubeConfig.Config)
	if err != nil {
		return nil, err
	}

	return &ContextClient{
		ContextName:      "inClusterContext",
		ContextNamespace: kubeConfig.DefaultNamespace,
		Client:           clientset,
	}, nil
}

func getTokenContextClient(opts KubeConfigOptions) (*ContextClient, error) {
	if opts.BearerToken == "" {
		return nil, fmt.Errorf("missing bearer token")
	}
	if opts.APIServerURL == "" {
		return nil, fmt.Errorf("missing API server URL")
	}

	var caData []byte
	var err error

	if opts.CADataBase64 != "" {
		caData, err = base64.StdEncoding.DecodeString(opts.CADataBase64)
		if err != nil {
			return nil, fmt.Errorf("invalid CADataBase64: %w", err)
		}
	}

	cfg := &rest.Config{
		Host:        opts.APIServerURL,
		BearerToken: opts.BearerToken,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: opts.Insecure,
			CAData:   caData,
		},
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("cannot create kubernetes client: %w", err)
	}

	return &ContextClient{
		ContextName:      "token",
		ContextNamespace: "",
		Client:           clientset,
	}, nil
}

func GetKubernetesContextClients(configPath, configDataBase64 string, configPathMergeList []string, kubeContext, kubeBearerTokenData, kubeBearerTokenPath, apiServerURL, caDataBase64 string, insecure bool) ([]*ContextClient, error) {
	var res []*ContextClient
	if contextClients, err := GetAllContextsClients(GetAllContextsClientsOptions{
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

func GetKubernetesNamespaceRestrictionByContext(cmdData *common.CmdData, contextClients []*ContextClient) map[string]string {
	res := map[string]string{}
	for _, contextClient := range contextClients {
		if *cmdData.ScanContextNamespaceOnly {
			res[contextClient.ContextName] = contextClient.ContextNamespace
		} else {
			// "" - cluster scope, therefore all namespaces
			res[contextClient.ContextName] = ""
		}
	}

	for contextName, restrictionNamespace := range res {
		logboek.Debug().LogF("GetKubernetesNamespaceRestrictionByContext -- context %q restriction namespace %q\n", contextName, restrictionNamespace)
	}

	return res
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
