//go:build !embedstapel

package stapel

import (
	"context"
	"fmt"
)

func embeddedImageForPlatform(_ string) (embeddedImage, bool) {
	return embeddedImage{}, false
}

func loadEmbeddedImage(_ context.Context, targetPlatform string) error {
	return fmt.Errorf("no embedded stapel image for platform %q", targetPlatform)
}
