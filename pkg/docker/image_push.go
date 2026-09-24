package docker

import (
	"context"

	"github.com/docker/docker/api/types/versions"
)

// The daemon API version that introduced the --platform flag of docker push,
// taken from the flag annotation in docker/cli.
const pushPlatformAPIVersion = "1.46"

type PushImageOptions struct {
	// TargetPlatform, when set and supported by the daemon, limits the push to that platform's
	// manifest instead of the whole index the image may be tagged from.
	TargetPlatform string
}

func PushImage(ctx context.Context, ref string, opts PushImageOptions) error {
	return CliPushWithRetries(ctx, pushArgs(ref, opts.TargetPlatform, daemonSupports(ctx, pushPlatformAPIVersion))...)
}

func pushArgs(ref, targetPlatform string, platformSupported bool) []string {
	if targetPlatform == "" || !platformSupported {
		return []string{ref}
	}
	return []string{"--platform", targetPlatform, ref}
}

func daemonSupports(ctx context.Context, apiVersion string) bool {
	return versions.GreaterThanOrEqualTo(cli(ctx).CurrentVersion(), apiVersion)
}
