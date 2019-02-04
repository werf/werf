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

	WerfAnsibleArgs                          Env = "WERF_ANSIBLE_ARGS"
	WerfDockerConfig                         Env = "WERF_DOCKER_CONFIG"
	WerfInsecureRepo                         Env = "WERF_INSECURE_REPO"
	WerfSecretKey                            Env = "WERF_SECRET_KEY"
	WerfDisableStagesCleanupDatePeriodPolicy Env = "WERF_DISABLE_STAGES_CLEANUP_DATE_PERIOD_POLICY"
	WerfGitTagsExpiryDatePeriodPolicy        Env = "WERF_GIT_TAGS_EXPIRY_DATE_PERIOD_POLICY"
	WerfGitTagsLimitPolicy                   Env = "WERF_GIT_TAGS_LIMIT_POLICY"
	WerfGitCommitsExpiryDatePeriodPolicy     Env = "WERF_GIT_COMMITS_EXPIRY_DATE_PERIOD_POLICY"
	WerfGitCommitsLimitPolicy                Env = "WERF_GIT_COMMITS_LIMIT_POLICY"
)

var envDescription = map[Env]string{
	WerfAnsibleArgs:                          "",
	WerfDockerConfig:                         "",
	WerfInsecureRepo:                         "",
	WerfSecretKey:                            "",
	WerfDisableStagesCleanupDatePeriodPolicy: "",
	WerfGitTagsExpiryDatePeriodPolicy:        "",
	WerfGitTagsLimitPolicy:                   "",
	WerfGitCommitsExpiryDatePeriodPolicy:     "",
	WerfGitCommitsLimitPolicy:                "",
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
		rightPart := strings.TrimLeft(logger.FitTextWithIndentWithWidthMaxLimit(envDescription[env], leftPartLength+len(space), 100), " ")
		lines = append(lines, fmt.Sprintf("%s%s%s", leftPart, space, rightPart))
	}

	return strings.Join(lines, "\n")
}
