package gost

const (
	PropertyAttackSurface    = "GOST:attack_surface"
	PropertySecurityFunction = "GOST:security_function"
)

type Config struct {
	AttackSurface    GostValue `json:"attackSurface"`
	SecurityFunction GostValue `json:"securityFunction"`
}

type GostValue string

func (v GostValue) String() string {
	return string(v)
}

const (
	GostValueYes       GostValue = "yes"
	GostValueNo        GostValue = "no"
	GostValueInherit   GostValue = "inherit"
	GostValueUndefined GostValue = ""
)

func DefaultConfig() Config {
	return Config{
		AttackSurface:    GostValueYes,
		SecurityFunction: GostValueYes,
	}
}

func (c Config) Merge(other Config) Config {
	res := c
	if !other.AttackSurface.IsUndefined() {
		res.AttackSurface = other.AttackSurface
	}
	if !other.SecurityFunction.IsUndefined() {
		res.SecurityFunction = other.SecurityFunction
	}
	return res
}

func IsValidGostValue(v string) bool {
	return v == GostValueYes.String() || v == GostValueNo.String() || v == GostValueInherit.String()
}

func (v GostValue) IsUndefined() bool {
	return v == GostValueUndefined
}
