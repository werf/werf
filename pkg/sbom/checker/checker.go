package checker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/docker"
)

const (
	Image         = "registry.werf.io/sbom-toolkit/3p-ispras-sbom-checker:master"
	containerPath = "/sbom/input.json"
	errorPrefix   = "ERROR:"
	warningPrefix = "WARNING:"
)

type RunOptions struct {
	CheckVCS bool
}

func Run(ctx context.Context, paths []string, isprasFormat IsprasFormat, opts RunOptions) error {
	if err := checkFilesExisting(paths); err != nil {
		return err
	}

	header := fmt.Sprintf("Validating %d SBOM file(s) as %q", len(paths), isprasFormat)
	if opts.CheckVCS {
		header += " with VCS check"
	}

	return logboek.Context(ctx).Default().LogProcess(header).DoError(func() error {
		logboek.Context(ctx).Debug().LogF("Using checker image: %s\n", Image)

		var failures []string
		total := len(paths)

		for i, p := range paths {
			args, err := buildDockerArgs(p, isprasFormat, opts.CheckVCS)
			if err != nil {
				return fmt.Errorf("build docker args for %q: %w", p, err)
			}

			fileName := filepath.Base(p)

			out, err := docker.CliRun_RecordedOutput(ctx, args...)
			if err != nil && out == "" {
				return fmt.Errorf("run sbom-checker container for %s: %w", fileName, err)
			}

			if err := parseResult(ctx, out, fileName, i+1, total); err != nil {
				failures = append(failures, err.Error())
			}
		}

		passed := total - len(failures)
		logboek.Context(ctx).Default().LogF("Result: %d passed, %d failed\n", passed, len(failures))

		if len(failures) > 0 {
			return fmt.Errorf("%s", strings.Join(failures, "\n"))
		}

		return nil
	})
}

func checkFilesExisting(paths []string) error {
	for _, p := range paths {
		if _, err := os.Stat(p); err != nil {
			return fmt.Errorf("unable to access sbom file %q: %w", p, err)
		}
	}

	return nil
}

func buildDockerArgs(path string, isprasFormat IsprasFormat, checkVCS bool) ([]string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve absolute path for %q: %w", path, err)
	}

	args := []string{
		"--rm",
		"-v", absPath + ":" + containerPath + ":ro",
		Image,
		"--format", isprasFormat.String(),
		"--errors", "0",
	}

	if checkVCS {
		args = append(args, "--check-vcs")
	}

	return append(args, containerPath), nil
}

func parseResult(ctx context.Context, out, fileName string, index, total int) error {
	errs := extractPrefixedLines(out, errorPrefix)
	warnings := extractPrefixedLines(out, warningPrefix)

	if len(errs) == 0 && len(warnings) == 0 {
		logboek.Context(ctx).Default().LogF("(%d/%d) %s... OK\n", index, total, fileName)
		return nil
	}

	logboek.Context(ctx).Default().LogF("(%d/%d) %s... FAILED\n", index, total, fileName)
	for _, e := range errs {
		logboek.Context(ctx).Default().LogF("  %s\n", e)
	}
	for _, w := range warnings {
		logboek.Context(ctx).Default().LogF("  %s\n", w)
	}

	var details []string
	details = append(details, errs...)
	details = append(details, warnings...)

	return fmt.Errorf("validation failed for %s:\n%s", fileName, strings.Join(details, "\n"))
}

func extractPrefixedLines(text, prefix string) []string {
	var result []string
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, prefix) {
			result = append(result, trimmed)
		}
	}

	return result
}
