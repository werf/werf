package config

import (
	"fmt"
	"strings"
)

type Image struct {
	*ImageBase
	Docker *Docker
}

func (c *Image) ImageTree() (tree []ImageInterface) {
	return imageTree(c)
}

func (c *Image) validateInfiniteLoop() error {
	if err, errImagesStack := validateImageInfiniteLoop(c, []ImageInterface{}); err != nil {
		return fmt.Errorf("%s: %s", err, strings.Join(errImagesStack, " -> "))
	}

	return nil
}

func (c *Image) validate() error {
	if !oneOrNone([]bool{c.Shell != nil, c.Ansible != nil}) {
		return newDetailedConfigError("can not use shell and ansible builders at the same time!", nil, c.ImageBase.raw.doc)
	}

	return nil
}
