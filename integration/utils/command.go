package utils

import (
	"encoding/base64"
	"fmt"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func RunCommand(dir, command string, args ...string) ([]byte, error) {
	cmd := exec.Command(command, args...)

	if dir != "" {
		cmd.Dir = dir
	}

	res, err := cmd.CombinedOutput()

	_, _ = GinkgoWriter.Write(res)

	return res, err
}

func RunSucceedCommand(dir, command string, args ...string) {
	cmd := exec.Command(command, args...)

	if dir != "" {
		cmd.Dir = dir
	}

	cmd.Stdout = GinkgoWriter
	cmd.Stderr = GinkgoWriter

	errorDesc := fmt.Sprintf("%[2]s %[3]s (dir: %[1]s)", dir, command, strings.Join(args, " "))
	Ω(cmd.Run()).ShouldNot(HaveOccurred(), errorDesc)
}

func SucceedCommandOutput(dir, command string, args ...string) string {
	cmd := exec.Command(command, args...)

	if dir != "" {
		cmd.Dir = dir
	}

	errorDesc := fmt.Sprintf("%[2]s %[3]s (dir: %[1]s)", dir, command, strings.Join(args, " "))
	res, err := cmd.CombinedOutput()

	_, _ = GinkgoWriter.Write(res)

	Ω(err).ShouldNot(HaveOccurred(), errorDesc)
	return string(res)
}

func ShelloutPack(command string) string {
	return fmt.Sprintf("eval $(echo %s | base64 --decode)", base64.StdEncoding.EncodeToString([]byte(command)))
}
