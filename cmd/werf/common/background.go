package common

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/werf"
)

func WithContext(f func(ctx context.Context) error) error {
	if !IsBackgroundModeEnabled() {
		ctx := logboek.NewContext(context.Background(), logboek.DefaultLogger())
		return f(ctx)
	}

	fileName := filepath.Join(werf.GetServiceDir(), fmt.Sprintf("background_output_%d.log", os.Getpid()))

	out, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return fmt.Errorf("unable to open background output file %q: %w", fileName, err)
	}
	defer out.Close()

	ctx := logboek.NewContext(context.Background(), logboek.NewLogger(out, out))
	return f(ctx)
}

func IsBackgroundModeEnabled() bool {
	return os.Getenv("_WERF_BACKGROUND_MODE_ENABLED") == "1"
}
