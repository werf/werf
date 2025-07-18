package scanner

//go:generate enumer -type=SourceType -trimprefix=SourceType -transform=kebab -output=./scan_command_source_type_enumer.go
type SourceType uint

const (
	SourceTypeDocker SourceType = iota
	SourceTypeRegistry
	SourceTypeDir
)
