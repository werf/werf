package host_cleaning

//go:generate enumer -type=containerBackendType -trimprefix=containerBackend

type containerBackendType uint8

const (
	containerBackendDocker containerBackendType = iota
	containerBackendBuildah
	containerBackendTest
)
