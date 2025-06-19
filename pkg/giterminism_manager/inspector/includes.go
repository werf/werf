package inspector

func (i Inspector) InspectIncludesAllowUpdate() error {
	if i.giterminismConfig.IsUpdateIncludesAccepted() {
		return nil
	}

	return NewExternalDependencyFoundError(`using --allow-includes-update option is not allowed by giterminism.

The use of --allow-includes-update option might make builds and deployments unreproducible.
	`)
}
