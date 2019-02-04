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

	WerfHome                                 Env = "WERF_HOME"
	WerfTmp                                  Env = "WERF_TMP"
	WerfAnsibleArgs                          Env = "WERF_ANSIBLE_ARGS"
	WerfDockerConfig                         Env = "WERF_DOCKER_CONFIG"
	WerfInsecureRegistry                     Env = "WERF_INSECURE_REGISTRY"
	WerfSecretKey                            Env = "WERF_SECRET_KEY"
	WerfCleanupImagesPassword                Env = "WERF_CLEANUP_IMAGES_PASSWORD"
	WerfDisableStagesCleanupDatePeriodPolicy Env = "WERF_DISABLE_STAGES_CLEANUP_DATE_PERIOD_POLICY"
	WerfGitTagsExpiryDatePeriodPolicy        Env = "WERF_GIT_TAGS_EXPIRY_DATE_PERIOD_POLICY"
	WerfGitTagsLimitPolicy                   Env = "WERF_GIT_TAGS_LIMIT_POLICY"
	WerfGitCommitsExpiryDatePeriodPolicy     Env = "WERF_GIT_COMMITS_EXPIRY_DATE_PERIOD_POLICY"
	WerfGitCommitsLimitPolicy                Env = "WERF_GIT_COMMITS_LIMIT_POLICY"
)

var envDescription = map[Env]string{
	WerfHome:                                 "",
	WerfTmp:                                  "",
	WerfAnsibleArgs:                          "",
	WerfDockerConfig:                         "",
	WerfInsecureRegistry:                     "",
	WerfSecretKey:                            "",
	WerfCleanupImagesPassword:                "",
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
