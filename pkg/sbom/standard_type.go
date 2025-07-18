package sbom

//go:generate enumer -type=StandardType -trimprefix=StandardType -linecomment -transform=lower -output=./standard_type_enumer.go
type StandardType uint

const (
	StandardTypeCycloneDX16 StandardType = iota // CycloneDX@1.6
	StandardTypeCycloneDX15                     // CycloneDX@1.5
	StandardTypeSPDX23                          // SPDX@2.3
	StandardTypeSPDX22                          // SPDX@2.2
)
