package config

type Docker struct {
	Volume      []string
	Expose      []string
	Env         map[string]string
	Label       map[string]string
	Cmd         string
	Workdir     string
	User        string
	Entrypoint  string
	HealthCheck string

	ExactValues bool

	raw *rawDocker
}

func (c *Docker) validate() error {
	return nil
}
