package convert

import (
	cdx "github.com/CycloneDX/cyclonedx-go"

	"github.com/werf/werf/v2/pkg/sbom/cyclonedxutil/gost"
)

type ImageSBOM struct {
	Name string
	BOM  *cdx.BOM
	GOST GOSTValues
}

type GOSTValues struct {
	AttackSurface    gost.GostValue
	SecurityFunction gost.GostValue
}

type ProductMeta struct {
	AppName      string
	AppVersion   string
	Manufacturer string
}
