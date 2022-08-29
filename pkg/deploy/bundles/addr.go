package bundles

import (
	"fmt"
	"strings"

	"github.com/werf/werf/pkg/deploy/bundles/registry"
)

const (
	ArchiveSchema  = "archive:"
	RegistrySchema = "docker://"
)

type Addr struct {
	*ArchiveAddress
	*RegistryAddress
}

func (addr *Addr) String() string {
	if addr.RegistryAddress != nil {
		return addr.RegistryAddress.FullName()
	}
	if addr.ArchiveAddress != nil {
		return addr.ArchiveAddress.Path
	}
	return ""
}

type ArchiveAddress struct {
	Path string
}

type RegistryAddress struct {
	*registry.Reference
}

func ParseAddr(addr string) (*Addr, error) {
	switch {
	case strings.HasPrefix(addr, ArchiveSchema):
		return &Addr{ArchiveAddress: parseArchiveAddress(addr)}, nil
	case strings.HasPrefix(addr, RegistrySchema):
		if regAddr, err := parseRegistryAddress(addr); err != nil {
			return nil, fmt.Errorf("unable to parse registry address %q: %w", addr, err)
		} else {
			return &Addr{RegistryAddress: regAddr}, nil
		}
	default:
		if regAddr, err := parseRegistryAddress(addr); err != nil {
			return nil, fmt.Errorf("unable to parse registry address %q: %w", addr, err)
		} else {
			return &Addr{RegistryAddress: regAddr}, nil
		}
	}
}

func parseRegistryAddress(addr string) (*RegistryAddress, error) {
	cleanAddr := strings.TrimPrefix(addr, RegistrySchema)

	ref, err := registry.ParseReference(cleanAddr)
	if err != nil {
		return nil, err
	}

	if ref.Tag == "" {
		ref.Tag = "latest"
	}

	return &RegistryAddress{Reference: ref}, nil
}

func parseArchiveAddress(addr string) *ArchiveAddress {
	path := strings.TrimPrefix(addr, ArchiveSchema)
	return &ArchiveAddress{Path: path}
}
