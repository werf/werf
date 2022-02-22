package helm

import (
	"encoding/base64"
	"fmt"

	"k8s.io/client-go/restmapper"

	"k8s.io/client-go/discovery/cached/memory"

	"github.com/werf/kubedog/pkg/kube"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type ClientGetterFromConfigData struct {
	Context          string
	ConfigDataBase64 string

	ClientConfig clientcmd.ClientConfig
}

func NewClientGetterFromConfigData(context, configDataBase64 string) (*ClientGetterFromConfigData, error) {
	getter := &ClientGetterFromConfigData{Context: context, ConfigDataBase64: configDataBase64}

	if clientConfig, err := getter.getRawKubeConfigLoader(); err != nil {
		return nil, err
	} else {
		getter.ClientConfig = clientConfig
	}

	return getter, nil
}

func (getter *ClientGetterFromConfigData) ToRESTConfig() (*rest.Config, error) {
	return getter.ClientConfig.ClientConfig()
}

func (getter *ClientGetterFromConfigData) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	config, err := getter.ClientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}

	return memory.NewMemCacheClient(discoveryClient), nil
}

func (getter *ClientGetterFromConfigData) ToRESTMapper() (meta.RESTMapper, error) {
	discoveryClient, err := getter.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)
	expander := restmapper.NewShortcutExpander(mapper, discoveryClient)
	return expander, nil
}

func (getter *ClientGetterFromConfigData) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return getter.ClientConfig
}

func (getter *ClientGetterFromConfigData) getRawKubeConfigLoader() (clientcmd.ClientConfig, error) {
	if data, err := base64.StdEncoding.DecodeString(getter.ConfigDataBase64); err != nil {
		return nil, fmt.Errorf("unable to decode base64 config data: %s", err)
	} else {
		return kube.GetClientConfig(getter.Context, "", data, []string{})
	}
}
