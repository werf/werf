package e2e_build_test

import (
	. "github.com/onsi/ginkgo/v2"

	"github.com/werf/werf/v3/test/pkg/suite_init"
	"github.com/werf/werf/v3/test/pkg/werf"
)

type setupEnvOptions struct {
	ContainerBackendMode        string
	WithLocalRepo               bool
	WithFinalRepo               bool
	WithStagedDockerfileBuilder bool
	State                       string
}

func (opts setupEnvOptions) env() setupEnvOptions {
	return opts
}

// envCarrier is implemented by every per-table options struct, so that a single
// entry constructor can label entries by the resources they need.
type envCarrier interface {
	env() setupEnvOptions
}

// backendEntry is Entry labeled with the external resources its backend and
// storage need.
func backendEntry(description string, opts envCarrier, args ...interface{}) TableEntry {
	return Entry(description, append([]interface{}{opts, entryLabels(opts.env())}, args...)...)
}

func entryLabels(opts setupEnvOptions) Labels {
	labels := Labels{opts.ContainerBackendMode}

	if opts.ContainerBackendMode != "docker" {
		labels = append(labels, suite_init.LabelNeedsBuildah)
	}

	if opts.WithLocalRepo || opts.WithFinalRepo {
		labels = append(labels, suite_init.LabelNeedsRegistry)
	}

	return labels
}

func setupEnv(opts setupEnvOptions) {
	SuiteData.Stubs.SetEnv("WERF_BUILDAH_MODE", opts.ContainerBackendMode)

	if opts.WithLocalRepo {
		SuiteData.Stubs.SetEnv(
			"WERF_REPO",
			suite_init.TestRepo(SuiteData.ProjectName),
		)
	} else {
		SuiteData.Stubs.UnsetEnv("WERF_REPO")
	}

	if opts.WithFinalRepo {
		SuiteData.Stubs.SetEnv(
			"WERF_FINAL_REPO",
			suite_init.TestRepo(SuiteData.ProjectName+"-final"),
		)
	} else {
		SuiteData.Stubs.UnsetEnv("WERF_FINAL_REPO")
	}

	if opts.WithLocalRepo || opts.WithFinalRepo {
		SuiteData.Stubs.SetEnv("WERF_INSECURE_REGISTRY", "1")
		SuiteData.Stubs.SetEnv("WERF_SKIP_TLS_VERIFY_REGISTRY", "1")
	} else {
		SuiteData.Stubs.UnsetEnv("WERF_INSECURE_REGISTRY")
		SuiteData.Stubs.UnsetEnv("WERF_SKIP_TLS_VERIFY_REGISTRY")
	}

	if opts.WithStagedDockerfileBuilder {
		SuiteData.Stubs.SetEnv("WERF_FORCE_STAGED_DOCKERFILE", "1")
	} else {
		SuiteData.Stubs.UnsetEnv("WERF_FORCE_STAGED_DOCKERFILE")
	}

	SuiteData.Stubs.SetEnv("ENV_SECRET", "WERF_BUILD_SECRET")
}

func newWerfProject(repoDirname string) *werf.Project {
	return werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
}
