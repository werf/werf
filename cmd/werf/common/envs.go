package common

import (
	"fmt"
	"strings"

	"github.com/flant/werf/pkg/logger"
)

type Env string

const (
	CmdEnvAnno string = "environment"

	WerfHome                                   Env = "WERF_HOME"
	WerfTmp                                    Env = "WERF_TMP"
	WerfAnsibleArgs                            Env = "WERF_ANSIBLE_ARGS"
	WerfDockerConfig                           Env = "WERF_DOCKER_CONFIG"
	WerfIgnoreCIDockerAutologin                Env = "WERF_IGNORE_CI_DOCKER_AUTOLOGIN"
	WerfInsecureRegistry                       Env = "WERF_INSECURE_REGISTRY"
	WerfSecretKey                              Env = "WERF_SECRET_KEY"
	WerfCleanupRegistryPassword                Env = "WERF_CLEANUP_REGISTRY_PASSWORD"
	WerfDisableSyncLocalStagesDatePeriodPolicy Env = "WERF_DISABLE_SYNC_LOCAL_STAGES_DATE_PERIOD_POLICY"
	WerfGitTagsExpiryDatePeriodPolicy          Env = "WERF_GIT_TAGS_EXPIRY_DATE_PERIOD_POLICY"
	WerfGitTagsLimitPolicy                     Env = "WERF_GIT_TAGS_LIMIT_POLICY"
	WerfGitCommitsExpiryDatePeriodPolicy       Env = "WERF_GIT_COMMITS_EXPIRY_DATE_PERIOD_POLICY"
	WerfGitCommitsLimitPolicy                  Env = "WERF_GIT_COMMITS_LIMIT_POLICY"
)

var envDescription = map[Env]string{
	WerfHome:                    "",
	WerfTmp:                     "",
	WerfAnsibleArgs:             "",
	WerfDockerConfig:            "",
	WerfIgnoreCIDockerAutologin: "",
	WerfInsecureRegistry:        "",
	WerfSecretKey:               "",
	WerfCleanupRegistryPassword: "",
	WerfDisableSyncLocalStagesDatePeriodPolicy: "",
	WerfGitTagsExpiryDatePeriodPolicy:          "",
	WerfGitTagsLimitPolicy:                     "",
	WerfGitCommitsExpiryDatePeriodPolicy:       "",
	WerfGitCommitsLimitPolicy:                  "",
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
		rightPart := fitDescription(envDescription[env], leftPartLength+len(space))
		lines = append(lines, fmt.Sprintf("%s%s%s", leftPart, space, rightPart))
	}

	return strings.Join(lines, "\n")
}

func fitDescription(desc string, indentLength int) string {
	tw := logger.TerminalWidth()
	descLines := strings.Split(desc, "\n")
	if len(descLines[0]) > tw-indentLength {
		descWords := strings.Split(desc, " ")
		var cursor int
		var firstLine string
		for ind, word := range descWords {
			isLastWord := ind == len(descWords)
			toAdd := word
			if !isLastWord {
				toAdd += " "
			}

			if cursor+len(toAdd) > tw-indentLength {
				break
			}

			firstLine += toAdd
			cursor += len(toAdd)
		}

		fittedText := logger.FitText(desc[cursor:], indentLength)
		return firstLine + "\n" + fittedText
	} else {
		firstLine := descLines[0]
		if len(descLines) > 1 {
			otherText := strings.Join(descLines[1:], "\n")
			fittedText := logger.FitText(otherText, indentLength)
			return firstLine + "\n" + fittedText
		} else {
			return firstLine
		}
	}
}
