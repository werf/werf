package giterminism_inspector

var (
	LooseGiterminism bool
)

type InspectionOptions struct {
	LooseGiterminism bool
}

func Init(projectPath string, opts InspectionOptions) error {
	LooseGiterminism = opts.LooseGiterminism
	return nil
}
