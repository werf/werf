package image

import "context"

func handleImageFromName(ctx context.Context, from string, fromLatest bool, image *Image) error {
	image.baseImageName = from

	if fromLatest {
		if _, err := image.getFromBaseImageIdFromRegistry(ctx, image.baseImageName); err != nil {
			return err
		}
	}

	return nil
}
