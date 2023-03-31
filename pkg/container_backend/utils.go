package container_backend

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/docker/docker/pkg/stringid"

	"github.com/werf/logboek"
)

var (
	logImageInfoLeftPartWidth = 9
	logImageInfoFormat        = fmt.Sprintf("%%%ds: %%s\n", logImageInfoLeftPartWidth)
)

func Debug() bool {
	return os.Getenv("WERF_CONTAINER_RUNTIME_DEBUG") == "1"
}

func LogImageName(ctx context.Context, name string) {
	logboek.Context(ctx).Default().LogFDetails(logImageInfoFormat, "name", name)
}

func LogImageInfo(ctx context.Context, img LegacyImageInterface, prevStageImageSize int64, withPlatform bool) {
	LogImageName(ctx, img.Name())

	logboek.Context(ctx).Default().LogFDetails(logImageInfoFormat, "id", stringid.TruncateID(img.GetStageDescription().Info.ID))
	logboek.Context(ctx).Default().LogFDetails(logImageInfoFormat, "created", img.GetStageDescription().Info.GetCreatedAt())

	if prevStageImageSize == 0 {
		logboek.Context(ctx).Default().LogFDetails(logImageInfoFormat, "size", byteCountBinary(img.GetStageDescription().Info.Size))
	} else {
		logboek.Context(ctx).Default().LogFDetails(logImageInfoFormat, "size", fmt.Sprintf("%s (+%s)", byteCountBinary(img.GetStageDescription().Info.Size), byteCountBinary(img.GetStageDescription().Info.Size-prevStageImageSize)))
	}

	if withPlatform {
		logboek.Context(ctx).Default().LogFDetails(logImageInfoFormat, "platform", img.GetTargetPlatform())
	}
}

func LogMultiplatformImageInfo(ctx context.Context, platforms []string) {
	logboek.Context(ctx).Default().LogFDetails(logImageInfoFormat, "platform", strings.Join(platforms, ","))
}

func byteCountBinary(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
