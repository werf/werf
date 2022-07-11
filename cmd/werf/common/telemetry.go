package common

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/spf13/cobra"

	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/telemetry"
	"github.com/werf/werf/pkg/util"
)

func InitTelemetry(ctx context.Context) {
	if err := telemetry.Init(ctx, telemetry.TelemetryOptions{
		ErrorHandlerFunc: func(err error) {
			if err == nil {
				return
			}
			logTelemetryError(err.Error())
		},
	}); err != nil {
		logTelemetryError(fmt.Sprintf("unable to init: %s", err))
	}
}

func ShutdownTelemetry(ctx context.Context, exitCode int) {
	if err := telemetry.Shutdown(ctx); err != nil {
		logTelemetryError(fmt.Sprintf("unable to shutdown: %s", err))
	}
}

func logTelemetryError(msg string) {
	if !telemetry.IsLogsEnabled() {
		return
	}
	fmt.Fprintf(os.Stderr, "Telemetry error: %s\n", msg)
}

func TelemetryPreRun(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	telemetry.GetTelemetryWerfIO().SetCommand(ctx, getTelemetryCommand(cmd))

	if projectID, err := getTelemetryProjectID(ctx); err != nil {
		if telemetry.IsLogsEnabled() {
			fmt.Fprintf(os.Stderr, "Telemetry error: %s\n", err)
		}
	} else {
		telemetry.GetTelemetryWerfIO().SetProjectID(ctx, projectID)
	}

	telemetry.GetTelemetryWerfIO().CommandStarted(ctx)

	return nil
}

func getTelemetryProjectID(ctx context.Context) (string, error) {
	var projectID string

	var workingDir, gitWorkTree string

	if commonCmdData := GetCmdDataFromContext(ctx); commonCmdData != nil {
		if commonCmdData.GitWorkTree != nil && commonCmdData.Dir != nil {
			workingDir = GetWorkingDir(commonCmdData)

			var err error
			gitWorkTree, err = GetGitWorkTree(ctx, commonCmdData, workingDir)
			if err != nil {
				return "", fmt.Errorf("unable to get git work tree: %w", err)
			}
		}
	} else {
		workingDir = util.GetAbsoluteFilepath(".")

		if res, err := LookupGitWorkTree(ctx, workingDir); err != nil {
			return "", fmt.Errorf("unable to lookup git work tree from wd %q: %w", workingDir, err)
		} else {
			gitWorkTree = res
		}
	}

	if repo, err := getTelemetryLocalRepo(ctx, workingDir, gitWorkTree); err != nil {
		if telemetry.IsLogsEnabled() {
			fmt.Fprintf(os.Stderr, "Telemetry: unable to open local repo: %s\n", err)
		}
	} else {
		url, err := repo.RemoteOriginUrl(ctx)
		if err != nil {
			return "", fmt.Errorf("unable to get repo origin url: %w", err)
		}

		ep, err := transport.NewEndpoint(url)
		if err != nil {
			return "", fmt.Errorf("bad repo origin url %q: %w", url, err)
		}

		hashParts := []string{ep.Protocol, ep.Host, fmt.Sprintf("%d", ep.Port), ep.Path}

		if telemetry.IsLogsEnabled() {
			fmt.Fprintf(os.Stderr, "Telemetry: calculate projectID based on repo origin url\n")
		}
		projectID = util.Sha256Hash(hashParts...)
	}

	return projectID, nil
}

func getTelemetryLocalRepo(ctx context.Context, workingDir, gitWorkTree string) (*git_repo.Local, error) {
	isWorkingDirInsideGitWorkTree := util.IsSubpathOfBasePath(gitWorkTree, workingDir)
	areWorkingDirAndGitWorkTreeTheSame := gitWorkTree == workingDir
	if !(isWorkingDirInsideGitWorkTree || areWorkingDirAndGitWorkTreeTheSame) {
		return nil, fmt.Errorf("werf requires project dir — the current working directory or directory specified with --dir option (or WERF_DIR env var) — to be located inside the git work tree: %q is located outside of the git work tree %q", gitWorkTree, workingDir)
	}

	return git_repo.OpenLocalRepo(ctx, "own", gitWorkTree, git_repo.OpenLocalRepoOptions{})
}

func getTelemetryCommand(cmd *cobra.Command) string {
	commandParts := []string{cmd.Name()}
	c := cmd
	for {
		p := c.Parent()
		if p == nil {
			break
		}
		commandParts = append(commandParts, p.Name())
		c = p
	}

	var p []string
	for i := 0; i < len(commandParts); i++ {
		p = append(p, commandParts[len(commandParts)-i-1])
	}

	return strings.Join(p, " ")
}
