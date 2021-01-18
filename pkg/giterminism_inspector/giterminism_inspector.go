package giterminism_inspector

import (
	"context"
	"fmt"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/giterminism_inspector/config"
)

const giterminismDocPageURL = "https://werf.io/v1.2-alpha/documentation/advanced/configuration/giterminism.html"

var (
	LooseGiterminism         bool
	NonStrict                bool
	DevMode                  bool
	ReportedUncommittedPaths []string
	ReportedUntrackedPaths   []string

	giterminismConfig config.GiterminismConfig
)

type InspectionOptions struct {
	LooseGiterminism bool
	NonStrict        bool
	DevMode          bool
}

func Init(projectPath string, opts InspectionOptions) error {
	LooseGiterminism = opts.LooseGiterminism
	NonStrict = opts.NonStrict
	DevMode = opts.DevMode

	if c, err := config.PrepareConfig(projectPath); err != nil {
		return err
	} else {
		giterminismConfig = c
	}

	return nil
}

func IsUncommittedConfigGoTemplateRenderingFileAccepted(path string) (bool, error) {
	return giterminismConfig.Config.GoTemplateRendering.IsUncommittedFileAccepted(path)
}

func ReportUntrackedFile(ctx context.Context, path string) error {
	for _, p := range ReportedUntrackedPaths {
		if p == path {
			return nil
		}
	}
	ReportedUntrackedPaths = append(ReportedUntrackedPaths, path)

	if NonStrict {
		logboek.Context(ctx).Warn().LogF("WARNING: Untracked file %s was not taken into account (more info %s)\n", path, giterminismDocPageURL)
		return nil
	} else {
		return fmt.Errorf("restricted usage of untracked file %s (more info %s)", path, giterminismDocPageURL)
	}
}

func ReportUncommittedFile(ctx context.Context, path string) error {
	for _, p := range ReportedUncommittedPaths {
		if p == path {
			return nil
		}
	}
	ReportedUncommittedPaths = append(ReportedUncommittedPaths, path)

	if NonStrict {
		logboek.Context(ctx).Warn().LogF("WARNING: Uncommitted file %s was not taken into account (more info %s)\n", path, giterminismDocPageURL)
		return nil
	} else {
		return fmt.Errorf("restricted usage of uncommitted file %s (more info %s)", path, giterminismDocPageURL)
	}
}

func ReportUntrackedConfigGoTemplateRenderingFile(ctx context.Context, path string) error {
	return ReportUntrackedFile(ctx, path)
}

func ReportConfigGoTemplateRenderingEnv(_ context.Context, envName string) error {
	if isAccepted, err := giterminismConfig.Config.GoTemplateRendering.IsEnvNameAccepted(envName); err != nil {
		return err
	} else if isAccepted {
		return nil
	}

	return fmt.Errorf("env name %s is forbidden due to enabled giterminism mode (more info %s)", envName, giterminismDocPageURL)
}

func PrintInspectionDebrief(ctx context.Context) {
	headerPrinted := false
	printHeader := func() {
		if headerPrinted {
			return
		}
		logboek.Context(ctx).Warn().LogLn()
		logboek.Context(ctx).Warn().LogF("### Giterminism inspection debrief ###\n")
		logboek.Context(ctx).Warn().LogLn()
		headerPrinted = true
	}

	defer func() {
		if headerPrinted {
			logboek.Context(ctx).Warn().LogF("More info about giterminism in the werf is available at %s\n", giterminismDocPageURL)
			logboek.Context(ctx).Warn().LogLn()
		}
	}()

	if NonStrict {
		if len(ReportedUncommittedPaths) > 0 || len(ReportedUntrackedPaths) > 0 {
			printHeader()

			if len(ReportedUncommittedPaths) > 0 {
				logboek.Context(ctx).Warn().LogF("Following uncommitted files were not taken into account:\n")
				for _, path := range ReportedUncommittedPaths {
					logboek.Context(ctx).Warn().LogF(" - %s\n", path)
				}
				logboek.Context(ctx).Warn().LogLn()
			}

			if len(ReportedUntrackedPaths) > 0 {
				logboek.Context(ctx).Warn().LogF("Following untracked files were not taken into account:\n")
				for _, path := range ReportedUntrackedPaths {
					logboek.Context(ctx).Warn().LogF(" - %s\n", path)
				}
				logboek.Context(ctx).Warn().LogLn()
			}

		}
	}

	if LooseGiterminism {
		printHeader()

		logboek.Context(ctx).Warn().LogF("--loose-giterminism option (and WERF_LOOSE_GITERMINISM env variable) is forbidden and will be removed soon!\n")
		logboek.Context(ctx).Warn().LogLn()
		logboek.Context(ctx).Warn().LogF("Please use werf-giterminism.yaml config instead to loosen giterminism restrictions if needed.\n")
		logboek.Context(ctx).Warn().LogF("Description of werf-giterminsim.yaml configuration is available at %s\n", giterminismDocPageURL)
		logboek.Context(ctx).Warn().LogLn()
	}
}
