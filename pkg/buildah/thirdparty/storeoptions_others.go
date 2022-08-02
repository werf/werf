// This is from "github.com/containers/storage".
//go:build !linux
// +build !linux

package thirdparty

// IDMap contains a single entry for user namespace range remapping. An array
// of IDMap entries represents the structure that will be provided to the Linux
// kernel for creating a user namespace.
type IDMap struct {
	ContainerID int `json:"container_id"`
	HostID      int `json:"host_id"`
	Size        int `json:"size"`
}

// StoreOptions is used for passing initialization options to GetStore(), for
// initializing a Store object and the underlying storage that it controls.
type StoreOptions struct {
	RunRoot             string            `json:"runroot,omitempty"`
	GraphRoot           string            `json:"root,omitempty"`
	RootlessStoragePath string            `toml:"rootless_storage_path"`
	GraphDriverName     string            `json:"driver,omitempty"`
	GraphDriverOptions  []string          `json:"driver-options,omitempty"`
	UIDMap              []IDMap           `json:"uidmap,omitempty"`
	GIDMap              []IDMap           `json:"gidmap,omitempty"`
	RootAutoNsUser      string            `json:"root_auto_ns_user,omitempty"`
	AutoNsMinSize       uint32            `json:"auto_userns_min_size,omitempty"`
	AutoNsMaxSize       uint32            `json:"auto_userns_max_size,omitempty"`
	PullOptions         map[string]string `toml:"pull_options"`
	DisableVolatile     bool              `json:"disable-volatile,omitempty"`
}
