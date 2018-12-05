package config

type Docker struct {
	Volume     []string
	Expose     []string
	Env        map[string]string
	Label      map[string]string
	Cmd        []string
	Onbuild    []string
	Workdir    string
	User       string
	Entrypoint []string

	raw *rawDocker
}

func (c *Docker) validate() error {
	return nil
}
