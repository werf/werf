package instruction

type OnBuild struct {
	Instruction string
}

func NewOnBuild(instruction string) *OnBuild {
	return &OnBuild{Instruction: instruction}
}

func (i *OnBuild) Name() string {
	return "ONBUILD"
}
