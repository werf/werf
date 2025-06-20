package scanner

//go:generate enumer -type=Type -trimprefix=Type -output=./type_enumer.go
type Type uint

const (
	TypeSyft Type = iota
	TypeTrivy
)
