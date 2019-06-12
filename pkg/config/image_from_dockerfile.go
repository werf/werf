package config

type ImageFromDockerfile struct {
	Name       string
	Dockerfile string
	Context    string
	Target     string
	Args       map[string]interface{}
}
