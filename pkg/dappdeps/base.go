package dappdeps

import (
	"fmt"
	"strconv"
)

const BASE_VERSION = "0.2.3"

func BaseContainer() (string, error) {
	container := &container{
		Name:      fmt.Sprintf("dappdeps_base_%s", BASE_VERSION),
		ImageName: fmt.Sprintf("dappdeps/base:%s", BASE_VERSION),
		Volume:    fmt.Sprintf("/.dapp/deps/base/%s", BASE_VERSION),
	}

	if err := container.CreateIfNotExist(); err != nil {
		return "", err
	} else {
		return container.Name, nil
	}
}

func RmBinPath() string {
	return BaseBinPath("rm")
}

func BaseBinPath(bin string) string {
	return fmt.Sprintf("/.dapp/deps/base/%s/embedded/bin/%s", BASE_VERSION, bin)
}

func BasePath() string {
	return fmt.Sprintf("/.dapp/deps/base/%[1]s/embedded/bin:/.dapp/deps/base/%[1]s/embedded/sbin", BASE_VERSION)
}

func SudoCommand(owner, group string) string {
	cmd := ""

	if owner != "" || group != "" {
		cmd += fmt.Sprintf("%s -E", BaseBinPath("sudo"))

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
