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

	WerfDebugAnsibleArgs                     Env = "WERF_DEBUG_ANSIBLE_ARGS"
	WerfSecretKey                            Env = "WERF_SECRET_KEY"
	WerfDisableStagesCleanupDatePeriodPolicy Env = "WERF_DISABLE_STAGES_CLEANUP_DATE_PERIOD_POLICY"
	WerfGitTagsExpiryDatePeriodPolicy        Env = "WERF_GIT_TAGS_EXPIRY_DATE_PERIOD_POLICY"
	WerfGitTagsLimitPolicy                   Env = "WERF_GIT_TAGS_LIMIT_POLICY"
	WerfGitCommitsExpiryDatePeriodPolicy     Env = "WERF_GIT_COMMITS_EXPIRY_DATE_PERIOD_POLICY"
	WerfGitCommitsLimitPolicy                Env = "WERF_GIT_COMMITS_LIMIT_POLICY"
)

var envDescription = map[Env]string{
	WerfDebugAnsibleArgs:                     "Pass specified cli args to ansible (ANSIBLE_ARGS)",
	WerfSecretKey:                            "Use specified secret key to extract secrets for the deploy; recommended way to set secret key in CI-system",
	WerfDisableStagesCleanupDatePeriodPolicy: "Redefine default ",
	WerfGitTagsExpiryDatePeriodPolicy:        "Redefine default tags expiry date period policy: keep images built for git tags, that are no older than 30 days since build time",
	WerfGitTagsLimitPolicy:                   "Redefine default tags limit policy: keep no more than 10 images built for git tags",
	WerfGitCommitsExpiryDatePeriodPolicy:     "Redefine default commits expiry date period policy: keep images built for git commits, that are no older than 30 days since build time",
	WerfGitCommitsLimitPolicy:                "Redefine default commits limit policy: keep no more than 50 images built for git commits",
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
