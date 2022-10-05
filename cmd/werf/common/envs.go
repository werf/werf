package common

import (
	"fmt"
	"strings"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/types"
)

type Env string

const (
	CmdEnvAnno                  string = "environment"
	DisableOptionsInUseLineAnno string = "disableOptionsInUseLine"

	DocsLongMD string = "docsLongMD"

	WerfDebugAnsibleArgs Env = "WERF_DEBUG_ANSIBLE_ARGS"
	WerfSecretKey        Env = "WERF_SECRET_KEY"
	WerfOldSecretKey     Env = "WERF_OLD_SECRET_KEY"
)

var envDescription = map[Env]string{
	WerfDebugAnsibleArgs: "Pass specified cli args to ansible ($ANSIBLE_ARGS)",
	WerfSecretKey: `Use specified secret key to extract secrets for the deploy. Recommended way to set secret key in CI-system.

Secret key also can be defined in files:
* ~/.werf/global_secret_key (globally),
* .werf_secret_key (per project)`,
	WerfOldSecretKey: "Use specified old secret key to rotate secrets",
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
		fitTextOptions := types.FitTextOptions{MaxWidth: 100, ExtraIndentWidth: leftPartLength + len(space)}
		rightPart := strings.TrimLeft(logboek.FitText(envDescription[env], fitTextOptions), " ")
		lines = append(lines, fmt.Sprintf("%s%s%s", leftPart, space, rightPart))
	}

	return strings.Join(lines, "\n")
}
