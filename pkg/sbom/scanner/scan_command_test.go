package scanner

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/matchers"
	"github.com/onsi/gomega/types"

	"github.com/werf/werf/v2/pkg/sbom"
)

var _ = Describe("ScanCommand", func() {
	Describe("NewSyftScanCommand()", func() {
		It("should work", func() {
			Expect(NewSyftScanCommand()).To(Equal(ScanCommand{
				scannerType:     TypeSyft,
				scannerExecPath: "/syft",
				SourceType:      SourceTypeDocker,
				SourcePath:      "",
				OutputStandard:  sbom.StandardTypeCycloneDX16,
				OutputPath:      "",
				outputFormat:    "json",
			}))
		})
	})

	DescribeTable("output()",
		func(scanCmd ScanCommand, expected types.GomegaMatcher) {
			switch expected.(type) {
			case *matchers.PanicMatcher:
				Expect(func() { scanCmd.output() }).To(expected)
			default:
				Expect(scanCmd.output()).To(expected)
			}
		},
		Entry(
			"should panic for scanner=trivy",
			ScanCommand{
				scannerType: TypeTrivy,
			},
			Panic(),
		),
		Entry(
			"should work for scanner=syft, standard=CycloneDX@1.6, format=json, path=",
			ScanCommand{
				scannerType:    TypeSyft,
				OutputStandard: sbom.StandardTypeCycloneDX16,
				OutputPath:     "",
				outputFormat:   "json",
			},
			Equal("cyclonedx-json@1.6"),
		),
		Entry(
			"should work for scanner=syft, standard=CycloneDX@1.6, format=json, path=file",
			ScanCommand{
				scannerType:    TypeSyft,
				OutputStandard: sbom.StandardTypeCycloneDX16,
				OutputPath:     "file.json",
				outputFormat:   "json",
			},
			Equal("cyclonedx-json@1.6=file.json"),
		),
		Entry(
			"should work for scanner=syft, standard=CycloneDX@1.5, format=json, path=",
			ScanCommand{
				scannerType:    TypeSyft,
				OutputStandard: sbom.StandardTypeCycloneDX15,
				OutputPath:     "",
				outputFormat:   "json",
			},
			Equal("cyclonedx-json@1.5"),
		),
		Entry(
			"should work for scanner=syft, standard=CycloneDX@1.5, format=json, path=file",
			ScanCommand{
				scannerType:    TypeSyft,
				OutputStandard: sbom.StandardTypeCycloneDX15,
				OutputPath:     "file.json",
				outputFormat:   "json",
			},
			Equal("cyclonedx-json@1.5=file.json"),
		),
		Entry(
			"should work for scanner=syft, standard=SPDX@2.3, format=json, path=",
			ScanCommand{
				scannerType:    TypeSyft,
				OutputStandard: sbom.StandardTypeSPDX23,
				OutputPath:     "",
				outputFormat:   "json",
			},
			Equal("spdx-json@2.3"),
		),
		Entry(
			"should work for scanner=syft, standard=SPDX@2.3, format=json, path=file",
			ScanCommand{
				scannerType:    TypeSyft,
				OutputStandard: sbom.StandardTypeSPDX23,
				OutputPath:     "file.json",
				outputFormat:   "json",
			},
			Equal("spdx-json@2.3=file.json"),
		),
		Entry(
			"should work for scanner=syft, standard=SPDX@2.2, format=json, path=",
			ScanCommand{
				scannerType:    TypeSyft,
				OutputStandard: sbom.StandardTypeSPDX22,
				OutputPath:     "",
				outputFormat:   "json",
			},
			Equal("spdx-json@2.2"),
		),
		Entry(
			"should work for scanner=syft, standard=SPDX@2.2, format=json, path=file",
			ScanCommand{
				scannerType:    TypeSyft,
				OutputStandard: sbom.StandardTypeSPDX22,
				OutputPath:     "file.json",
				outputFormat:   "json",
			},
			Equal("spdx-json@2.2=file.json"),
		),
	)

	DescribeTable("String()",
		func(scanCmd ScanCommand, expected types.GomegaMatcher) {
			switch expected.(type) {
			case *matchers.PanicMatcher:
				Expect(func() { _ = scanCmd.String() }).To(expected)
			default:
				Expect(scanCmd.String()).To(expected)
			}
		},
		Entry(
			"should panic for scanner=trivy",
			ScanCommand{
				scannerType: TypeTrivy,
			},
			Panic(),
		),
		Entry(
			"should work for scanner_type=syft, scanner_exec_path=/syft, source_type=dir, source_path=some, output_standard=CycloneDX@1.6, output_path=some; output_format=json",
			ScanCommand{
				scannerType:     TypeSyft,
				scannerExecPath: "/syft",
				SourceType:      SourceTypeDir,
				SourcePath:      "/some/dir",
				OutputStandard:  sbom.StandardTypeCycloneDX16,
				OutputPath:      "file.json",
				outputFormat:    "json",
			},
			Equal("/syft scan dir:/some/dir --output=cyclonedx-json@1.6=file.json"),
		),
		Entry(
			"should work for scanner_type=syft, scanner_exec_path=, source_type=docker, source_path=some, output_standard=CycloneDX@1.6, output_path=; output_format=json",
			ScanCommand{
				scannerType:     TypeSyft,
				scannerExecPath: "",
				SourceType:      SourceTypeDocker,
				SourcePath:      "alpine:3.18",
				OutputStandard:  sbom.StandardTypeCycloneDX16,
				OutputPath:      "",
				outputFormat:    "json",
			},
			Equal("scan docker:alpine:3.18 --output=cyclonedx-json@1.6"),
		),
	)

	DescribeTable("Checksum()",
		func(scanCmd ScanCommand, expected types.GomegaMatcher) {
			Expect(scanCmd.Checksum()).To(expected)
		},
		Entry(
			"should work for scanner_type=syft, source_type=docker, source_path=some, output_standard=CycloneDX@1.6",
			ScanCommand{
				scannerType:    TypeSyft,
				SourceType:     SourceTypeDir,
				SourcePath:     "/some/dir",
				OutputStandard: sbom.StandardTypeCycloneDX16,
			},
			Equal("71b089eff92427ff29ba2986ef6d5c025865ce37d186f6bde2d1737708d7539b"),
		),
	)
})
