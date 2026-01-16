package container_backend

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/docker/docker/pkg/stringid"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/ssh_agent"
)

const (
	SSHHostAuthSockPath      = "/run/host-services/ssh-auth.sock"
	SSHContainerAuthSockPath = "/.werf/tmp/ssh-auth-sock"
	hostCleanupServiceImage  = "alpine:3.14"
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

	logboek.Context(ctx).Default().LogFDetails(logImageInfoFormat, "id", stringid.TruncateID(img.GetStageDesc().Info.ID))
	logboek.Context(ctx).Default().LogFDetails(logImageInfoFormat, "created", img.GetStageDesc().Info.GetCreatedAt())

	if prevStageImageSize == 0 {
		logboek.Context(ctx).Default().LogFDetails(logImageInfoFormat, "size", byteCountBinary(img.GetStageDesc().Info.Size))
	} else {
		logboek.Context(ctx).Default().LogFDetails(logImageInfoFormat, "size", fmt.Sprintf("%s (+%s)", byteCountBinary(img.GetStageDesc().Info.Size), byteCountBinary(img.GetStageDesc().Info.Size-prevStageImageSize)))
	}

	if commit, ok := img.GetStageDesc().Info.Labels[image.WerfProjectRepoCommitLabel]; ok && commit != "" {
		logboek.Context(ctx).Default().LogFDetails(logImageInfoFormat, "commit", commit)
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

func setSSHMountPoint(sshAuthSock string) (string, map[string]string) {
	var vol string
	env := make(map[string]string)
	if runtime.GOOS == "darwin" {
		// On MacOs Docker Desktop creates /run/host-services/ssh-auth.sock as a special bridge to your host system’s SSH_AUTH_SOCK.
		vol = fmt.Sprintf("%s:%s", SSHHostAuthSockPath, SSHHostAuthSockPath)
		env[ssh_agent.SSHAuthSockEnv] = SSHHostAuthSockPath
	} else {
		vol = fmt.Sprintf("%s:%s", sshAuthSock, SSHContainerAuthSockPath)
		env[ssh_agent.SSHAuthSockEnv] = SSHContainerAuthSockPath
	}
	return vol, env
}

func getHostCleanupServiceImage() string {
	imageName := hostCleanupServiceImage
	if v := os.Getenv("WERF_HOST_CLEANUP_SERVICE_IMAGE"); v != "" {
		imageName = v
	}

	return imageName
}
