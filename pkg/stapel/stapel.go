package stapel

import (
	"fmt"
	"strconv"

	"github.com/flant/werf/pkg/docker"
)

const VERSION = "0.1.2"

func ImageName() string {
	return fmt.Sprintf("flant/werf-stapel:%s", VERSION)
}

func getContainer() container {
	return container{
		Name:      fmt.Sprintf("stapel_%s", VERSION),
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

func FindBinPath() string {
	return embeddedBinPath("find")
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

func SudoBinPath() string {
	return embeddedBinPath("sudo")
}

func AnsiblePlaybookBinPath() string {
	return embeddedBinPath("ansible-playbook")
}

func SystemPATH() string {
	return fmt.Sprintf("/.werf/stapel/embedded/bin:/.werf/stapel/embedded/sbin")
}

func SudoCommand(owner, group string) string {
	cmd := ""

	if owner != "" || group != "" {
		cmd += fmt.Sprintf("%s -E", embeddedBinPath("sudo"))

		if owner != "" {
			cmd += fmt.Sprintf(" -u %s", sudoFormatUser(owner))
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
