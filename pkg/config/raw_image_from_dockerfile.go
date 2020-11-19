package config

import (
	"fmt"
	"path/filepath"
)

type rawImageFromDockerfile struct {
	Images         []string               `yaml:"-"`
	Dockerfile     string                 `yaml:"dockerfile,omitempty"`
	Context        string                 `yaml:"context,omitempty"`
	ContextAddFile []string               `yaml:"contextAddFile,omitempty"`
	Target         string                 `yaml:"target,omitempty"`
	Args           map[string]interface{} `yaml:"args,omitempty"`
	AddHost        interface{}            `yaml:"addHost,omitempty"`
	Network        string                 `yaml:"network,omitempty"`
	SSH            string                 `yaml:"ssh,omitempty"`

	doc *doc `yaml:"-"` // parent

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *rawImageFromDockerfile) setAndValidateImage() error {
	value, ok := c.UnsupportedAttributes["image"]
	if ok {
		delete(c.UnsupportedAttributes, "image")

		switch t := value.(type) {
		case []interface{}:
			if images, err := InterfaceToStringArray(value, nil, c.doc); err != nil {
				return err
			} else {
				c.Images = images
			}
		case string:
			c.Images = []string{value.(string)}
		case nil:
			c.Images = []string{""}
		default:
			return newDetailedConfigError(fmt.Sprintf("invalid image name `%v`!", t), nil, c.doc)
		}
	}

	return nil
}

func (c *rawImageFromDockerfile) UnmarshalYAML(unmarshal func(interface{}) error) error {
	parentStack.Push(c)
	type plain rawImageFromDockerfile
	err := unmarshal((*plain)(c))
	parentStack.Pop()
	if err != nil {
		return err
	}

	if err := c.setAndValidateImage(); err != nil {
		return err
	}

	if err := checkOverflow(c.UnsupportedAttributes, nil, c.doc); err != nil {
		return err
	}

	return nil
}

func (c *rawImageFromDockerfile) toImageFromDockerfileDirectives() (images []*ImageFromDockerfile, err error) {
	for _, imageName := range c.Images {
		if image, err := c.toImageFromDockerfileDirective(imageName); err != nil {
			return nil, err
		} else {
			images = append(images, image)
		}
	}

	return images, nil
}

func (c *rawImageFromDockerfile) toImageFromDockerfileDirective(imageName string) (image *ImageFromDockerfile, err error) {
	image = &ImageFromDockerfile{}
	image.Name = imageName
	image.Dockerfile = filepath.FromSlash(c.Dockerfile)
	image.Context = filepath.FromSlash(c.Context)

	for _, path := range c.ContextAddFile {
		image.ContextAddFile = append(image.ContextAddFile, filepath.FromSlash(path))
	}

	image.Target = c.Target
	image.Args = c.Args

	if addHost, err := InterfaceToStringArray(c.AddHost, c, c.doc); err != nil {
		return nil, err
	} else {
		image.AddHost = addHost
	}

	image.Network = c.Network
	image.SSH = c.SSH

	image.raw = c

	return image, nil
}
