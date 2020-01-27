package stages_storage

type MemoryCache struct {
	ImagesInspects map[string]*ImageInspect
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{ImagesInspects: make(map[string]*ImageInspect)}
}

func (cache *MemoryCache) GetImageInspect(imageName string) (*ImageInspect, error) {
	return cache.ImagesInspects[imageName], nil
}

func (cache *MemoryCache) SetImageInspect(imageName string, inspect *ImageInspect) error {
	cache.ImagesInspects[imageName] = inspect
	return nil
}
