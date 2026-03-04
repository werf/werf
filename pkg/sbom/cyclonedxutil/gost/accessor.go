package gost

import (
	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/samber/lo"
)

type accessor struct {
	comp *cdx.Component
}

func newAccessor(comp *cdx.Component) *accessor {
	return &accessor{comp: comp}
}

func (a *accessor) GetAttackSurface() (GostValue, bool) {
	return a.getProperty(PropertyAttackSurface)
}

func (a *accessor) SetAttackSurface(val GostValue) {
	a.setProperty(PropertyAttackSurface, val)
}

func (a *accessor) GetSecurityFunction() (GostValue, bool) {
	return a.getProperty(PropertySecurityFunction)
}

func (a *accessor) SetSecurityFunction(val GostValue) {
	a.setProperty(PropertySecurityFunction, val)
}

func (a *accessor) getProperty(name string) (GostValue, bool) {
	for _, prop := range lo.FromPtr(a.comp.Properties) {
		if prop.Name == name {
			return GostValue(prop.Value), true
		}
	}
	return GostValueUndefined, false
}

func (a *accessor) setProperty(name string, val GostValue) {
	if val.IsUndefined() {
		return
	}

	// update case
	for i, prop := range lo.FromPtr(a.comp.Properties) {
		if prop.Name == name {
			(*a.comp.Properties)[i].Value = val.String()
			return
		}
	}

	if a.comp.Properties == nil {
		a.comp.Properties = lo.ToPtr([]cdx.Property{})
	}

	// insert case
	*a.comp.Properties = append(*a.comp.Properties, cdx.Property{
		Name:  name,
		Value: val.String(),
	})
}
