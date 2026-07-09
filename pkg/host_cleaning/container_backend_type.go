package host_cleaning

//go:generate enumer -type=containerBackendType -trimprefix=containerBackend

type containerBackendType uint8

const (
	containerBackendTest containerBackendType = iota
)
