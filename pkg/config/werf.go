package config

type WerfConfig struct {
	Meta   *Meta
	Images []*Image
}

func (c *WerfConfig) HasImage(imageName string) bool {
	for _, image := range c.Images {
		if image.Name == imageName {
			return true
		}
	}

	return false
}
