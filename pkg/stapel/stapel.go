package stapel

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/containerd/containerd/platforms"
	cerrdefs "github.com/containerd/errdefs"
	"github.com/moby/moby/client"

	"github.com/werf/werf/v3/pkg/container_backend/thirdparty/platformutil"
	"github.com/werf/werf/v3/pkg/docker"
	"github.com/werf/werf/v3/pkg/image"
)

const (
	VERSION              = "0.7.2"
	IMAGE                = "registry.werf.io/werf/stapel"
	CONTAINER_MOUNT_ROOT = "/.werf"

	containerVolumeDestination = CONTAINER_MOUNT_ROOT + "/stapel"
)

func getVersion() string {
	version := VERSION
	if v := os.Getenv("WERF_STAPEL_IMAGE_VERSION"); v != "" {
		version = v
	}
	return version
}

func getImage() string {
	image := IMAGE
	if i := os.Getenv("WERF_STAPEL_IMAGE_NAME"); i != "" {
		image = i
	}
	return image
}

func isDefaultImageRef() bool {
	return os.Getenv("WERF_STAPEL_IMAGE_NAME") == "" && os.Getenv("WERF_STAPEL_IMAGE_VERSION") == ""
}

func ImageName() string {
	return fmt.Sprintf("%s:%s", getImage(), getVersion())
}

func containerName(version, targetPlatform string) string {
	suffix := version
	if targetPlatform != "" {
		suffix = fmt.Sprintf("%s_%s", suffix, strings.ReplaceAll(targetPlatform, "/", "_"))
	}

	return fmt.Sprintf("%s%s", image.AssemblingContainerNamePrefix, suffix)
}

func getContainer(targetPlatform string) container {
	return container{
		Name:      containerName(getVersion(), targetPlatform),
		ImageName: ImageName(),
		Volume:    containerVolumeDestination,
		Platform:  targetPlatform,
	}
}

func GetOrCreateContainer(ctx context.Context, targetPlatform string) (string, error) {
	container := getContainer(targetPlatform)

	if err := container.CreateIfNotExist(ctx); err != nil {
		return "", err
	} else {
		return container.Name, nil
	}
}

func isContainerNameOfVersion(name, version string) bool {
	base := containerName(version, "")
	if name == base {
		return true
	}

	// Custom versions may contain underscores, so match the full version prefix.
	suffix, ok := strings.CutPrefix(name, base+"_")
	if !ok {
		return false
	}

	spec, err := platformutil.ParsePlatform(strings.ReplaceAll(suffix, "_", "/"))
	if err != nil {
		return false
	}

	return containerName(version, platforms.Format(spec)) == name
}

func Purge(ctx context.Context) error {
	containers, err := docker.Containers(ctx, client.ContainerListOptions{
		All:     true,
		Filters: make(client.Filters).Add("name", image.AssemblingContainerNamePrefix),
	})
	if err != nil {
		return fmt.Errorf("list stapel containers: %w", err)
	}

	// A failed removal must not hide the remaining ones: report every failure and
	// leave the host with as little stapel garbage as the daemon allows.
	var errs []error

	for _, c := range containers {
		if err := rmContainerWithVolumes(ctx, c.ID); err != nil {
			errs = append(errs, err)
		}
	}

	if err := rmiIfExist(ctx); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func rmContainerWithVolumes(ctx context.Context, id string) error {
	inspect, err := docker.ContainerInspect(ctx, id)
	if err != nil {
		if cerrdefs.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("inspect container %s: %w", id, err)
	}

	if inspect.Config == nil {
		return nil
	}

	version, ok := strings.CutPrefix(inspect.Config.Image, getImage()+":")
	if !ok || version == "" {
		return nil
	}

	if !isContainerNameOfVersion(strings.TrimPrefix(inspect.Name, "/"), version) {
		return nil
	}

	// Only one mount can sit at the stapel mount point.
	var volumeName string
	for _, m := range inspect.Mounts {
		if m.Type == "volume" && m.Destination == containerVolumeDestination {
			volumeName = m.Name
			break
		}
	}

	// The container is the only record of its volume name: Purge finds volumes
	// through the mounts of the containers it owns, and nothing prunes volumes
	// afterwards. Removing the container while another one still holds the volume
	// would orphan the volume for good, so keep the container as the anchor and
	// let a later purge take both.
	if volumeName != "" {
		holders, err := volumeHolders(ctx, volumeName, id)
		if err != nil {
			return err
		}
		if len(holders) > 0 {
			return fmt.Errorf("keep container %s: volume %s is in use by %s", id, volumeName, strings.Join(holders, ", "))
		}
	}

	if err := docker.ContainerRemove(ctx, id, client.ContainerRemoveOptions{}); err != nil {
		if cerrdefs.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("remove container %s: %w", id, err)
	}

	if volumeName == "" {
		return nil
	}

	// A conflict here is a container that appeared after the check above.
	if err := docker.VolumeRm(ctx, volumeName, false); err != nil && !cerrdefs.IsNotFound(err) {
		return fmt.Errorf("remove volume %s: %w", volumeName, err)
	}

	return nil
}

// volumeHolders returns the ids of the containers other than exceptID that
// reference the volume.
func volumeHolders(ctx context.Context, volumeName, exceptID string) ([]string, error) {
	containers, err := docker.Containers(ctx, client.ContainerListOptions{
		All:     true,
		Filters: make(client.Filters).Add("volume", volumeName),
	})
	if err != nil {
		return nil, fmt.Errorf("list containers using volume %s: %w", volumeName, err)
	}

	var holders []string
	for _, c := range containers {
		if c.ID != exceptID {
			holders = append(holders, c.ID)
		}
	}

	return holders, nil
}

func rmiIfExist(ctx context.Context) error {
	exist, err := docker.ImageExist(ctx, ImageName())
	if err != nil {
		return err
	}

	if exist {
		return docker.CliRmi(ctx, ImageName())
	}

	return nil
}

func TrueBinPath() string {
	return embeddedBinPath("true")
}

func LsBinPath() string {
	return embeddedBinPath("ls")
}

func RmBinPath() string {
	return embeddedBinPath("rm")
}

func InstallBinPath() string {
	return embeddedBinPath("install")
}

func ChownBinPath(_ context.Context) string {
	return embeddedBinPath("chown")
}

func XargsBinPath() string {
	return embeddedBinPath("xargs")
}

func TarBinPath() string {
	return embeddedBinPath("tar")
}

func MkdirBinPath() string {
	return embeddedBinPath("mkdir")
}

func BashBinPath() string {
	return embeddedBinPath("bash")
}

func RsyncBinPath() string {
	return embeddedBinPath("rsync")
}

func HeadBinPath() string {
	return embeddedBinPath("head")
}

func embeddedBinPath(bin string) string {
	return path.Join(CONTAINER_MOUNT_ROOT, "stapel/embedded/bin", bin)
}

func CreateScript(path string, lines []string) error {
	dirPath := filepath.Dir(path)
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return fmt.Errorf("unable to create dir %s: %w", dirPath, err)
	}

	var scriptLines []string
	scriptLines = append(scriptLines, fmt.Sprintf("#!%s -e", BashBinPath()))
	scriptLines = append(scriptLines, "")
	scriptLines = append(scriptLines, lines...)
	scriptData := []byte(strings.Join(scriptLines, "\n") + "\n")

	if err := os.WriteFile(path, scriptData, 0o755); err != nil {
		return fmt.Errorf("write script %s: %w", path, err)
	}

	// os.WriteFile applies the process umask, which can leave the script without any
	// executable bit, and then even root cannot run it.
	if err := os.Chmod(path, 0o755); err != nil {
		return fmt.Errorf("chmod script %s: %w", path, err)
	}

	return nil
}
