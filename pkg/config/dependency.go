package config

type Dependency struct {
	ImageName string
	Imports   []*DependencyImport

	// TODO: raw *rawDependencies
}
