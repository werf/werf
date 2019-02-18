package common

import (
	"fmt"
	"strings"

	"github.com/flant/werf/pkg/logger"
)

type Env string

const (
	CmdEnvAnno                  string = "environment"
	DisableOptionsInUseLineAnno string = "disableOptionsInUseLine"

	WerfDebugAnsibleArgs Env = "WERF_DEBUG_ANSIBLE_ARGS"
	WerfSecretKey        Env = "WERF_SECRET_KEY"
)

var envDescription = map[Env]string{
	WerfDebugAnsibleArgs: "Pass specified cli args to ansible (ANSIBLE_ARGS)",
	WerfSecretKey:        "Use specified secret key to extract secrets for the deploy; recommended way to set secret key in CI-system",
}

func EnvsDescription(envs ...Env) string {
	var lines []string

	var envNameWidth int
	for _, env := range envs {
		if len(env) > envNameWidth {
			envNameWidth = len(env)
		}
	}

	for _, env := range envs {
		leftPart := strings.Join([]string{"  ", "$", string(env), strings.Repeat(" ", envNameWidth-len(env))}, "")
		leftPartLength := len(leftPart)
		space := "  "
		fitTextOptions := logger.FitTextOptions{MaxWidth: 100, ExtraIndentWidth: leftPartLength + len(space)}
		rightPart := strings.TrimLeft(logger.FitText(envDescription[env], fitTextOptions), " ")
		lines = append(lines, fmt.Sprintf("%s%s%s", leftPart, space, rightPart))
	}

	return strings.Join(lines, "\n")
}
