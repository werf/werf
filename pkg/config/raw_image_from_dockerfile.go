package config

import (
	"fmt"

	"github.com/werf/werf/v2/pkg/giterminism_manager"
	"github.com/werf/werf/v2/pkg/util/option"
)

type rawImageFromDockerfile struct {
	Images          []string               `yaml:"-"`
	Final           *bool                  `yaml:"final,omitempty"`
	Dockerfile      string                 `yaml:"dockerfile,omitempty"`
	CacheVersion    string                 `yaml:"cacheVersion,omitempty"`
	Context         string                 `yaml:"context,omitempty"`
	ContextAddFile  interface{}            `yaml:"contextAddFile,omitempty"`
	ContextAddFiles interface{}            `yaml:"contextAddFiles,omitempty"`
	Target          string                 `yaml:"target,omitempty"`
	Args            map[string]interface{} `yaml:"args,omitempty"`
	AddHost         interface{}            `yaml:"addHost,omitempty"`
	Network         string                 `yaml:"network,omitempty"`
	SSH             string                 `yaml:"ssh,omitempty"`
	RawDependencies []*rawDependency       `yaml:"dependencies,omitempty"`
	Staged          bool                   `yaml:"staged,omitempty"`
	Platform        []string               `yaml:"platform,omitempty"`
	RawSbom         *rawSbom               `yaml:"sbom,omitempty"`
	RawSecrets      []*rawSecret           `yaml:"secrets,omitempty"`
	RawImageSpec    *rawImageSpec          `yaml:"imageSpec,omitempty"`

	doc          *doc `yaml:"-"` // parent
	isFillStaged bool `yaml:"-"` // indicates whether 'staged' field is explicitly set in the image section

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

func (c *rawImageFromDockerfile) toImageFromDockerfileDirectives(giterminismManager giterminism_manager.Interface) (images []*ImageFromDockerfile, err error) {
	for _, imageName := range c.Images {
		if image, err := c.toImageFromDockerfileDirective(giterminismManager, imageName); err != nil {
			return nil, err
		} else {
			images = append(images, image)
		}
	}

	return images, nil
}

func (c *rawImageFromDockerfile) toImageFromDockerfileDirective(giterminismManager giterminism_manager.Interface, imageName string) (image *ImageFromDockerfile, err error) {
	image = &ImageFromDockerfile{}
	image.Name = imageName
	image.Dockerfile = c.Dockerfile
	image.Context = c.Context

	image.cacheVersion = c.CacheVersion
	image.final = option.PtrValueOrDefault(c.Final, true)

	contextAddFile, err := InterfaceToStringArray(c.ContextAddFile, nil, c.doc)
	if err != nil {
		return nil, err
	}
	contextAddFiles, err := InterfaceToStringArray(c.ContextAddFiles, nil, c.doc)
	if err != nil {
		return nil, err
	}

	switch {
	case len(contextAddFile) > 0 && len(contextAddFiles) > 0:
		return nil, newDetailedConfigError("only one out of contextAddFiles and contextAddFile directives can be used at a time, but both specified in werf.yaml. Move everything out of the contextAddFile: [] directive into the contextAddFiles: [] directive and remove contextAddFile: [] from werf.yaml", nil, c.doc)
	case len(contextAddFile) > 0:
		image.ContextAddFiles = contextAddFile
	default:
		image.ContextAddFiles = contextAddFiles
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

	for _, rawDep := range c.RawDependencies {
		dependencyDirective, err := rawDep.toDirective()
		if err != nil {
			return nil, err
		}

		image.Dependencies = append(image.Dependencies, dependencyDirective)
	}

	image.Staged = c.Staged
	image.platform = append([]string{}, c.Platform...)
	image.raw = c

	secrets, err := GetValidatedSecrets(c.RawSecrets, giterminismManager, c.doc)
	if err != nil {
		return nil, err
	}

	image.Secrets = secrets

	if c.RawImageSpec != nil {
		image.ImageSpec = c.RawImageSpec.toDirective()
	}

	if c.RawSbom != nil {
		image.sbom = c.RawSbom.toDirective()
	}

	if err := image.validate(giterminismManager); err != nil {
		return nil, err
	}

	return image, nil
}

func (r *rawImageFromDockerfile) getDoc() *doc {
	return r.doc
}
