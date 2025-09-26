package stapel

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/werf/werf/v2/pkg/docker"
	"github.com/werf/werf/v2/pkg/image"
)

const (
	VERSION              = "0.6.2"
	IMAGE                = "registry.werf.io/werf/stapel"
	CONTAINER_MOUNT_ROOT = "/.werf"
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

func ImageName() string {
	return fmt.Sprintf("%s:%s", getImage(), getVersion())
}

func getContainer() container {
	return container{
		Name:      fmt.Sprintf("%s%s", image.AssemblingContainerNamePrefix, getVersion()),
		ImageName: ImageName(),
		Volume:    filepath.Join(CONTAINER_MOUNT_ROOT, "stapel"),
	}
}

func GetOrCreateContainer(ctx context.Context) (string, error) {
	container := getContainer()

	if err := container.CreateIfNotExist(ctx); err != nil {
		return "", err
	} else {
		return container.Name, nil
	}
}

func Purge(ctx context.Context) error {
	container := getContainer()
	if err := container.RmIfExist(ctx); err != nil {
		return err
	}

	if err := rmiIfExist(ctx); err != nil {
		return err
	}

	return nil
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

func Base64BinPath() string {
	return embeddedBinPath("base64")
}

func LsBinPath() string {
	return embeddedBinPath("ls")
}

func RmBinPath() string {
	return embeddedBinPath("rm")
}

func GitBinPath() string {
	return embeddedBinPath("git")
}

func PythonBinPath() string {
	return embeddedBinPath("python")
}

func InstallBinPath() string {
	return embeddedBinPath("install")
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

func CutBinPath() string {
	return embeddedBinPath("cut")
}

func RsyncBinPath() string {
	return embeddedBinPath("rsync")
}

func HeadBinPath() string {
	return embeddedBinPath("head")
}

func StatBinPath() string {
	return embeddedBinPath("stat")
}

func SudoBinPath() string {
	return embeddedBinPath("sudo")
}

func FindBinPath() string {
	return embeddedBinPath("find")
}

func SortBinPath() string {
	return embeddedBinPath("sort")
}

func Md5sumBinPath() string {
	return embeddedBinPath("md5sum")
}

func AnsiblePlaybookBinPath() string {
	return embeddedBinPath("ansible-playbook")
}

/*
 * Ansible tools and libs overlay path is like /usr/local which has more priority than /usr.
 * Ansible tools and libs overlay path used to force ansible to use tools directly from stapel rather than find it in the base system.
 *
 * Use case is "unarchive" module which does not work with alpine busybox "tar" util (which is installed by default
 * and takes precedence over other utils). For this case we put tar into ansible tools overlay path.
 */

func AnsibleToolsOverlayPATH() string {
	return filepath.Join(CONTAINER_MOUNT_ROOT, "stapel/ansible_tools_overlay/bin")
}

func AnsibleLibsOverlayLDPATH() string {
	return filepath.Join(CONTAINER_MOUNT_ROOT, "stapel/ansible_tools_overlay/lib")
}

func SystemPATH() string {
	return strings.Join([]string{
		filepath.Join(CONTAINER_MOUNT_ROOT, "stapel/sbin"),
		filepath.Join(CONTAINER_MOUNT_ROOT, "stapel/embedded/sbin"),
		filepath.Join(CONTAINER_MOUNT_ROOT, "stapel/bin"),
		filepath.Join(CONTAINER_MOUNT_ROOT, "stapel/embedded/bin"),
	}, ":")
}

func embeddedBinPath(bin string) string {
	return filepath.Join(CONTAINER_MOUNT_ROOT, "stapel/embedded/bin", bin)
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

	return os.WriteFile(path, scriptData, os.FileMode(0o667))
}
