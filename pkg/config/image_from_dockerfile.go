package config

type ImageFromDockerfile struct {
	Name           string
	Dockerfile     string
	Context        string
	ContextAddFile []string
	Target         string
	Args           map[string]interface{}
	AddHost        []string
	Network        string
	SSH            string

	raw *rawImageFromDockerfile
}

func (c *ImageFromDockerfile) validate() error {
	if c.Dockerfile == "" || !isRelativePath(c.Dockerfile) {
		return newDetailedConfigError("`dockerfile: PATH` required and should be relative path!", nil, c.raw.doc)
	} else if !allRelativePaths(c.ContextAddFile) {
		return newDetailedConfigError("`contextAddFile: [PATH, ...]|PATH` should be relative paths!", nil, c.raw.doc)
	}
	return nil
}

func (c *ImageFromDockerfile) GetName() string {
	return c.Name
}
