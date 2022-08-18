package stapel

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/werf/werf/pkg/docker"
)

const (
	VERSION = "0.6.2"
	IMAGE   = "registry.werf.io/werf/stapel"
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
		Name:      fmt.Sprintf("stapel_%s", getVersion()),
		ImageName: ImageName(),
		Volume:    "/.werf/stapel",
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
	return "/.werf/stapel/ansible_tools_overlay/bin"
}

func AnsibleLibsOverlayLDPATH() string {
	return "/.werf/stapel/ansible_tools_overlay/lib"
}

func SystemPATH() string {
	return "/.werf/stapel/sbin:/.werf/stapel/embedded/sbin:/.werf/stapel/bin:/.werf/stapel/embedded/bin"
}

func OptionalSudoCommand(user, group string) string {
	cmd := ""

	if user != "" || group != "" {
		cmd += fmt.Sprintf("%s -E", embeddedBinPath("sudo"))

		if user != "" {
			cmd += fmt.Sprintf(" -u %s -H", sudoFormatUser(user))
		}

		if group != "" {
			cmd += fmt.Sprintf(" -g %s", sudoFormatUser(group))
		}
	}

	return cmd
}

func sudoFormatUser(user string) string {
	var userStr string
	userInt, err := strconv.Atoi(user)
	if err == nil {
		userStr = strconv.Itoa(userInt)
	}

	if user == userStr {
		return fmt.Sprintf("\\#%s", user)
	} else {
		return user
	}
}

func embeddedBinPath(bin string) string {
	return fmt.Sprintf("/.werf/stapel/embedded/bin/%s", bin)
}

func CreateScript(path string, commands []string) error {
	dirPath := filepath.Dir(path)
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return fmt.Errorf("unable to create dir %s: %w", dirPath, err)
	}

	var scriptLines []string
	scriptLines = append(scriptLines, fmt.Sprintf("#!%s -e", BashBinPath()))
	scriptLines = append(scriptLines, "")
	scriptLines = append(scriptLines, commands...)
	scriptData := []byte(strings.Join(scriptLines, "\n") + "\n")

	return ioutil.WriteFile(path, scriptData, os.FileMode(0o667))
}
