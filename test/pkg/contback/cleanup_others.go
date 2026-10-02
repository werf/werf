//go:build !linux

package contback

import (
	"context"
	"fmt"
)

func cleanupBuildahProject(context.Context, string, []string) error {
	return fmt.Errorf("buildah test cleanup requires Linux")
}
