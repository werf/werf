package stapel

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/werf/werf/pkg/docker"
)

const VERSION = "0.6.1"

func getVersion() string {
	version := VERSION
	if v := os.Getenv("WERF_STAPEL_IMAGE_VERSION"); v != "" {
		version = v
	}
	return version
}

func ImageName() string {
	return fmt.Sprintf("flant/werf-stapel:%s", getVersion())
}

func getContainer() container {
	return container{
		Name:      fmt.Sprintf("stapel_%s", getVersion()),
		ImageName: ImageName(),
		Volume:    "/.werf/stapel",
	}
}

func GetOrCreateContainer() (string, error) {
	container := getContainer()

	if err := container.CreateIfNotExist(); err != nil {
		return "", err
	} else {
		return container.Name, nil
	}
}

func Purge() error {
	container := getContainer()
	if err := container.RmIfExist(); err != nil {
		return err
	}

	if err := rmiIfExist(); err != nil {
		return err
	}

	return nil
}

func rmiIfExist() error {
	exist, err := docker.ImageExist(ImageName())
	if err != nil {
		return err
	}

	if exist {
		return docker.CliRmi(ImageName())
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

func RsyncBinPath() string {
	return embeddedBinPath("rsync")
}

func HeadBinPath() string {
	return embeddedBinPath("head")
}

func SudoBinPath() string {
	return embeddedBinPath("sudo")
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
		return fmt.Errorf("unable to create dir %s: %s", dirPath, err)
	}

	var scriptLines []string
	scriptLines = append(scriptLines, fmt.Sprintf("#!%s -e", BashBinPath()))
	scriptLines = append(scriptLines, "")
	scriptLines = append(scriptLines, commands...)
	scriptData := []byte(strings.Join(scriptLines, "\n") + "\n")

	return ioutil.WriteFile(path, scriptData, os.FileMode(0667))
}
