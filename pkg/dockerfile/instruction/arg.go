package instruction

import "github.com/moby/buildkit/frontend/dockerfile/instructions"

type Arg struct {
	*Base

	Args []instructions.KeyValuePairOptional
}

func NewArg(raw string, args []instructions.KeyValuePairOptional) *Arg {
	return &Arg{Base: NewBase(raw), Args: args}
}

func (i *Arg) Name() string {
	return "ARG"
}
