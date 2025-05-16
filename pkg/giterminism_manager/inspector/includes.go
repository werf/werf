package inspector

func (i Inspector) InspectIncludesAllowUpdate() bool {
	return i.giterminismConfig.IsUpdateIncludesAccepted()
}
