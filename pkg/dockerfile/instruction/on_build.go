package instruction

type OnBuild struct {
	*Base

	Instruction string
}

func NewOnBuild(raw, instruction string) *OnBuild {
	return &OnBuild{Base: NewBase(raw), Instruction: instruction}
}

func (i *OnBuild) Name() string {
	return "ONBUILD"
}
