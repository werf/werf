package platformutil

import (
	"fmt"

	"github.com/containerd/containerd/platforms"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

func NormalizeUserParams(platformParams []string) ([]string, error) {
	specs, err := Parse(platformParams)
	if err != nil {
		return nil, fmt.Errorf("unable to parse platform specs: %w", err)
	}

	return Format(Dedupe(specs)), nil
}

func GetPlatformMetaArgsMap(targetPlatform string) (map[string]string, error) {
	var bp, tp specs.Platform
	{
		bp = platforms.DefaultSpec()
		if targetPlatform != "" {
			p, err := platforms.Parse(targetPlatform)
			if err != nil {
				return nil, err
			}

			tp = p
		} else {
			tp = bp
		}
	}

	return map[string]string{
		"BUILDPLATFORM":  platforms.Format(bp),
		"BUILDOS":        bp.OS,
		"BUILDARCH":      bp.Architecture,
		"BUILDVARIANT":   bp.Variant,
		"TARGETPLATFORM": platforms.Format(tp),
		"TARGETOS":       tp.OS,
		"TARGETARCH":     tp.Architecture,
		"TARGETVARIANT":  tp.Variant,
	}, nil
}
