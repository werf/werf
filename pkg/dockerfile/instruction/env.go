package instruction

type Env struct {
	*Base

	Envs map[string]string
}

func NewEnv(raw string, envs map[string]string) *Env {
	return &Env{Base: NewBase(raw), Envs: envs}
}

func (i *Env) Name() string {
	return "ENV"
}
