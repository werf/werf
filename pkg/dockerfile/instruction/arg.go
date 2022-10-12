package instruction

import "github.com/moby/buildkit/frontend/dockerfile/instructions"

type Arg struct {
	Args []instructions.KeyValuePairOptional
}

func NewArg(args []instructions.KeyValuePairOptional) *Arg {
	return &Arg{Args: args}
}

func (i *Arg) Name() string {
	return "ARG"
}
