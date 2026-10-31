package bundles

import (
	"fmt"
	"strings"

	"github.com/werf/werf/v3/pkg/ref"
)

// ParseSourceAddr parses the address of a bundle to read, with "+" in the tag, invalid in a registry
// tag, encoded as "_" as Helm does. A registry address without a tag gets the latest tag.
func ParseSourceAddr(addr string) (*ref.Addr, error) {
	return ref.ParseAddr(encodeAddrTag(addr))
}

// ParseDestinationAddr parses the address of a bundle to write. The tag of a registry address is kept as
// given, since the chart version is derived from it, and is empty when not given: RemoteBundle pushes
// the chart with "+" in the tag encoded.
func ParseDestinationAddr(addr string) (*ref.Addr, error) {
	parsed, err := ParseSourceAddr(addr)
	if err != nil {
		return nil, err
	}

	if parsed.RegistryAddress != nil {
		_, tag, _ := splitAddrTag(addr)
		parsed.RegistryAddress.Tag = tag
	}

	return parsed, nil
}

func encodeAddrTag(addr string) string {
	rest, tag, ok := splitAddrTag(addr)
	if !ok {
		return addr
	}

	return fmt.Sprintf("%s:%s", rest, encodeTag(tag))
}

// splitAddrTag splits a registry address into the part before the tag and the tag. Archive and digest
// addresses have no tag.
func splitAddrTag(addr string) (string, string, bool) {
	if strings.HasPrefix(addr, ref.ArchiveSchema) || strings.Contains(addr, "@") {
		return addr, "", false
	}

	i := strings.LastIndex(addr, ":")
	if i < 0 || i < strings.LastIndex(addr, "/") {
		return addr, "", false
	}

	return addr[:i], addr[i+1:], true
}

func encodeTag(tag string) string {
	return strings.ReplaceAll(tag, "+", "_")
}
