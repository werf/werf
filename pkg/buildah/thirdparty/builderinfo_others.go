//go:build !linux
// +build !linux

package thirdparty

import (
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

//
// BuilderInfo is copied from github.com/containers/buildah
// to provide crossplatform types definitions in the github.com/werf/werf/pkg/buildah package.
//

// BuilderInfo are used as objects to display container information
type BuilderInfo struct {
	Type                  string
	FromImage             string
	FromImageID           string
	FromImageDigest       string
	Config                string
	Manifest              string
	Container             string
	ContainerID           string
	MountPoint            string
	ProcessLabel          string
	MountLabel            string
	ImageAnnotations      map[string]string
	ImageCreatedBy        string
	OCIv1                 v1.Image
	Docker                V2Image
	DefaultMountsFilePath string
	Isolation             string
	// NamespaceOptions      define.NamespaceOptions
	Capabilities     []string
	ConfigureNetwork string
	CNIPluginPath    string
	CNIConfigDir     string
	// IDMappingOptions      define.IDMappingOptions
	History []v1.History
	// Devices               define.ContainerDevices
}
