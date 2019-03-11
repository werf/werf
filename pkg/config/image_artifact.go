package config

import (
	"fmt"
	"strings"
)

type ImageArtifact struct {
	*ImageBase
}

func (c *ImageArtifact) IsArtifact() bool {
	return true
}

func (c *ImageArtifact) validateInfiniteLoop() error {
	if err, errImagesStack := validateImageInfiniteLoop(c, []ImageInterface{}); err != nil {
		return fmt.Errorf("%s: %s", err, strings.Join(errImagesStack, " -> "))
	}

	return nil
}

func (c *ImageArtifact) validate() error {
	if !oneOrNone([]bool{c.Shell != nil, c.Ansible != nil}) {
		return newDetailedConfigError("can not use shell and ansible builders at the same time!", nil, c.ImageBase.raw.doc)
	}

	return nil
}
