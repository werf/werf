package common

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/telemetry"
	"github.com/werf/werf/pkg/util"
)

var telemetryIgnoreCommands = []string{
	"werf version",
	"werf synchronization",
	"werf completion",
}

func InitTelemetry(ctx context.Context) {
	if err := telemetry.Init(ctx, telemetry.TelemetryOptions{
		ErrorHandlerFunc: func(err error) {
			if err == nil {
				return
			}

			telemetry.LogF("error: %s", err)
		},
	}); err != nil {
		telemetry.LogF("error: %s", err)
	}
}

func ShutdownTelemetry(ctx context.Context, exitCode int) {
	telemetry.GetTelemetryWerfIO().CommandExited(ctx, exitCode)

	if err := telemetry.Shutdown(ctx); err != nil {
		telemetry.LogF("unable to shutdown: %s", err)
	}
}

func TelemetryPreRun(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	command := getTelemetryCommand(cmd)

	for _, c := range telemetryIgnoreCommands {
		if command == c {
			return nil
		}
	}

	InitTelemetry(ctx)
	telemetry.GetTelemetryWerfIO().SetCommand(ctx, command)

	var commandOptions []telemetry.CommandOption
	for _, fs := range []*flag.FlagSet{cmd.Flags(), cmd.PersistentFlags(), cmd.LocalFlags(), cmd.InheritedFlags()} {
		fs.VisitAll(func(f *flag.Flag) {
			if !f.Changed {
				return
			}

			for _, opt := range commandOptions {
				if opt.Name == f.Name {
					return
				}
			}

			commandOptions = append(commandOptions, telemetry.CommandOption{
				Name: f.Name,
			})
		})
	}
	telemetry.GetTelemetryWerfIO().SetCommandOptions(ctx, commandOptions)

	if userID, err := getTelemetryUserID(ctx); err != nil {
		telemetry.LogF("error: %s", err)
	} else {
		telemetry.GetTelemetryWerfIO().SetUserID(ctx, userID)
	}

	if projectID, err := getTelemetryProjectID(ctx); err != nil {
		telemetry.LogF("error: %s", err)
	} else {
		telemetry.GetTelemetryWerfIO().SetProjectID(ctx, projectID)
	}

	telemetry.GetTelemetryWerfIO().CommandStarted(ctx)

	return nil
}

func getTelemetryUserID(_ context.Context) (string, error) {
	macAddress, err := getMACAddress()
	if err != nil {
		return "", fmt.Errorf("unable to get mac address: %w", err)
	}

	return util.Sha256Hash(macAddress), nil
}

func getMACAddress() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, i := range interfaces {
		if i.Flags&net.FlagUp != 0 && bytes.Compare(i.HardwareAddr, nil) != 0 {
			// skip locally administered addresses
			if i.HardwareAddr[0]&2 == 2 {
				continue
			}

			return i.HardwareAddr.String(), nil
		}
	}

	return "", nil
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
		telemetry.LogF("unable to detect projectID: unable to open local repo: %s", err)
	} else {
		url, err := repo.RemoteOriginUrl(ctx)
		if err != nil {
			return "", fmt.Errorf("unable to get repo origin url: %w", err)
		}

		hash, err := hashOriginUrl(url)
		if err != nil {
			return "", fmt.Errorf("unable to hash origin url: %w", err)
		}

		telemetry.LogF("calculate projectID based on repo origin url")
		projectID = hash
	}

	return projectID, nil
}

func hashOriginUrl(url string) (string, error) {
	ep, err := transport.NewEndpoint(url)
	if err != nil {
		return "", fmt.Errorf("bad repo origin url %q: %w", url, err)
	}

	var formatPath string
	{
		formatPath = ep.Path

		paramsIndex := strings.Index(formatPath, "?")
		if paramsIndex > -1 {
			formatPath = formatPath[:paramsIndex]
		}

		formatPath = strings.TrimPrefix(formatPath, "/")
		formatPath = strings.TrimSuffix(formatPath, ".git")
	}

	hashParts := []string{ep.Host, formatPath}
	return util.Sha256Hash(hashParts...), nil
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
