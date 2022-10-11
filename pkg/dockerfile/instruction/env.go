package instruction

type Env struct {
	Envs map[string]string
}

func NewEnv(envs map[string]string) *Env {
	return &Env{Envs: envs}
}

func (i *Env) Name() string {
	return "ENV"
}
