package convert

import (
	cdx "github.com/CycloneDX/cyclonedx-go"

	"github.com/werf/werf/v2/pkg/sbom/cyclonedxutil/gost"
)

var gostPrecedence = map[gost.GostValue]int{
	gost.GostValueUndefined: 0,
	gost.GostValueNo:        1,
	gost.GostValueIndirect:  2,
	gost.GostValueYes:       3,
}

func aggregateGOST(components []cdx.Component) GOSTValues {
	var result GOSTValues
	for i := range components {
		cfg := gost.GetComponent(&components[i])
		result.AttackSurface = maxGOSTValue(result.AttackSurface, cfg.AttackSurface)
		result.SecurityFunction = maxGOSTValue(result.SecurityFunction, cfg.SecurityFunction)
	}
	return result
}

func aggregateImageGOST(images []*ImageSBOM) GOSTValues {
	var result GOSTValues
	for _, img := range images {
		result.AttackSurface = maxGOSTValue(result.AttackSurface, img.GOST.AttackSurface)
		result.SecurityFunction = maxGOSTValue(result.SecurityFunction, img.GOST.SecurityFunction)
	}
	return result
}

func maxGOSTValue(a, b gost.GostValue) gost.GostValue {
	if gostPrecedence[b] > gostPrecedence[a] {
		return b
	}
	return a
}

func setMissingGOSTOnComponent(comp *cdx.Component, values GOSTValues) {
	current := gost.GetComponent(comp)

	attack := current.AttackSurface
	if attack == gost.GostValueUndefined {
		attack = values.AttackSurface
	}

	security := current.SecurityFunction
	if security == gost.GostValueUndefined {
		security = values.SecurityFunction
	}

	gost.SetComponent(comp, gost.Config{
		AttackSurface:    attack,
		SecurityFunction: security,
	})
}
