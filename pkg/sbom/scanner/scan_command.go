package scanner

import (
	"fmt"
	"strings"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v2/pkg/sbom"
)

type ScanCommand struct {
	scannerType     Type
	scannerExecPath string
	SourceType      SourceType
	SourcePath      string
	OutputStandard  sbom.StandardType
	OutputPath      string
	outputFormat    string
}

func (c ScanCommand) output() string {
	switch c.scannerType {
	case TypeSyft:
		var out strings.Builder

		switch c.OutputStandard {
		case sbom.StandardTypeCycloneDX16:
			out.WriteString(fmt.Sprintf("cyclonedx-%s@1.6", c.outputFormat))
		case sbom.StandardTypeCycloneDX15:
			out.WriteString(fmt.Sprintf("cyclonedx-%s@1.5", c.outputFormat))
		case sbom.StandardTypeSPDX23:
			out.WriteString(fmt.Sprintf("spdx-%s@2.3", c.outputFormat))
		case sbom.StandardTypeSPDX22:
			out.WriteString(fmt.Sprintf("spdx-%s@2.2", c.outputFormat))
		default:
			panic(fmt.Sprintf("unsupported %s", c.OutputStandard))
		}

		if len(c.OutputPath) > 0 {
			out.WriteString(fmt.Sprintf("=%s", c.OutputPath))
		}

		return out.String()
	default:
		panic(fmt.Sprintf("unsupported scanner type %s", c.scannerType))
	}
}

func (c ScanCommand) String() string {
	switch c.scannerType {
	case TypeSyft:
		var out strings.Builder

		if len(c.scannerExecPath) > 0 {
			out.WriteString(fmt.Sprintf("%s ", c.scannerExecPath))
		}

		out.WriteString(fmt.Sprintf("scan %s:%s --output=%s", c.SourceType, c.SourcePath, c.output()))

		return out.String()
	default:
		panic(fmt.Sprintf("unsupported scanner type %s", c.scannerType))
	}
}

func (c ScanCommand) Checksum() string {
	args := []string{
		"scanner_type", c.scannerType.String(),
		"source_type", c.SourceType.String(),
		"output_standard", c.OutputStandard.String(),
	}
	return util.Sha256Hash(args...)
}

func NewSyftScanCommand() ScanCommand {
	return ScanCommand{
		scannerType:     TypeSyft,
		scannerExecPath: "/syft",
		SourceType:      SourceTypeDocker,
		SourcePath:      "",
		OutputStandard:  sbom.StandardTypeCycloneDX16,
		OutputPath:      "",
		outputFormat:    "json",
	}
}
