package storage

import (
	"errors"
	"strings"
)

var (
	ErrBadKubernetesSynchronizationAddress = errors.New("bad kubernetes synchronization address")
)

type KubernetesSynchronizationParams struct {
	ConfigContext       string
	ConfigPath          string
	ConfigDataBase64    string
	ConfigPathMergeList []string
	Namespace           string
}

func ParseKubernetesSynchronization(address string) (*KubernetesSynchronizationParams, error) {
	if !strings.HasPrefix(address, "kubernetes://") {
		return nil, ErrBadKubernetesSynchronizationAddress
	}
	addressWithoutScheme := strings.TrimPrefix(address, "kubernetes://")

	res := &KubernetesSynchronizationParams{}

	namespaceWithConextAndConfigParts := strings.SplitN(addressWithoutScheme, "@", 2)
	var namespaceWithContext, config string
	if len(namespaceWithConextAndConfigParts) == 2 {
		namespaceWithContext, config = namespaceWithConextAndConfigParts[0], namespaceWithConextAndConfigParts[1]
	} else {
		namespaceWithContext = namespaceWithConextAndConfigParts[0]
	}

	namespaceAndContextParts := strings.SplitN(namespaceWithContext, ":", 2)
	if len(namespaceAndContextParts) == 2 {
		res.Namespace, res.ConfigContext = namespaceAndContextParts[0], namespaceAndContextParts[1]
	} else {
		res.Namespace = namespaceAndContextParts[0]
	}

	if config != "" {
		if strings.HasPrefix(config, "base64:") {
			configBase64 := strings.TrimPrefix(config, "base64:")
			res.ConfigDataBase64 = configBase64
		} else {
			res.ConfigPath = config
		}
	}

	return res, nil
}
