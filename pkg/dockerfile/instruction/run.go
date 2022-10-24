package instruction

import (
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
)

type NetworkType string

const (
	NetworkDefault NetworkType = "default"
	NetworkNone    NetworkType = "none"
	NetworkHost    NetworkType = "host"
)

func NewNetworkType(network string) NetworkType {
	v := NetworkType(network)
	switch v {
	case NetworkDefault, NetworkHost, NetworkNone:
		return v
	default:
		panic(fmt.Sprintf("unknown network type %q", network))
	}
}

type SecurityType string

const (
	SecurityInsecure SecurityType = "insecure"
	SecuritySandbox  SecurityType = "sandbox"
)

func NewSecurityType(security string) SecurityType {
	v := SecurityType(security)
	switch v {
	case SecurityInsecure, SecuritySandbox:
		return v
	default:
		panic(fmt.Sprintf("unknown security type %q", security))
	}
}

type Run struct {
	*Base

	Command      []string
	PrependShell bool
	Mounts       []*instructions.Mount
	Network      NetworkType
	Security     SecurityType
}

func NewRun(raw string, command []string, prependShell bool, mounts []*instructions.Mount, network NetworkType, security SecurityType) *Run {
	return &Run{Base: NewBase(raw), Command: command, PrependShell: prependShell, Mounts: mounts, Network: network}
}

func (i *Run) Name() string {
	return "RUN"
}
