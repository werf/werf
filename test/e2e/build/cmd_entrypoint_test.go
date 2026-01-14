package e2e_build_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/contback"
	"github.com/werf/werf/v2/test/pkg/report"
	"github.com/werf/werf/v2/test/pkg/thirdparty/contruntime/strslice"
	"github.com/werf/werf/v2/test/pkg/werf"
)

type cmdEntrypointTestOptions struct {
	setupEnvOptions
}

var _ = Describe("CMD and ENTRYPOINT combinations", Label("e2e", "build", "extra"), func() {
	checkFunc := func(ctx SpecContext, testOpts cmdEntrypointTestOptions, imageName string, expectedEntrypoint, expectedCmd strslice.StrSlice) {
		SuiteData.Stubs.SetEnv("WERF_STAGED_DOCKERFILE_VERSION", "v2")

		repoDirname := "repo"
		fixtureRelPath := "cmd_entrypoint"
		buildReportName := fmt.Sprintf("report-%s.json", imageName)

		By("initializing")
		setupEnv(testOpts.setupEnvOptions)
		contRuntime, err := contback.NewContainerBackend(testOpts.ContainerBackendMode)
		if err == contback.ErrRuntimeUnavailable {
			Skip(err.Error())
		} else if err != nil {
			Fail(err.Error())
		}

		By("preparing test repo")
		SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)

		By("building images")
		werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
		reportProject := report.NewProjectWithReport(werfProject)
		buildOut, buildReport := reportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath(buildReportName), &werf.WithReportOptions{CommonOptions: werf.CommonOptions{
			ExtraArgs: []string{imageName},
		}})
		Expect(buildOut).To(ContainSubstring("Building stage"))

		By("getting built images metadata")
		inspectOfImage := contRuntime.GetImageInspect(ctx, buildReport.Images[imageName].DockerImageName)
		imgCfg := inspectOfImage.Config

		By("checking image metadata")
		if expectedEntrypoint == nil {
			Expect(imgCfg.Entrypoint == nil || len(imgCfg.Entrypoint) == 0).To(BeTrue(), "Expected Entrypoint to be nil or empty")
		} else {
			Expect(imgCfg.Entrypoint).To(Equal(expectedEntrypoint))
		}

		if expectedCmd == nil {
			Expect(imgCfg.Cmd == nil || len(imgCfg.Cmd) == 0).To(BeTrue(), "Expected Cmd to be nil or empty")
		} else {
			Expect(imgCfg.Cmd).To(Equal(expectedCmd))
		}
	}

	backends := []struct {
		name    string
		options setupEnvOptions
	}{
		{
			name: "vanilla-docker",
			options: setupEnvOptions{
				ContainerBackendMode:        "vanilla-docker",
				WithLocalRepo:               false,
				WithStagedDockerfileBuilder: false,
			},
		},
		{
			name: "native-rootless",
			options: setupEnvOptions{
				ContainerBackendMode:        "native-rootless",
				WithLocalRepo:               true,
				WithStagedDockerfileBuilder: false,
			},
		},
		{
			name: "native-rootless-staged",
			options: setupEnvOptions{
				ContainerBackendMode:        "native-rootless",
				WithLocalRepo:               true,
				WithStagedDockerfileBuilder: true,
			},
		},
	}

	for _, backend := range backends {
		// Prevent closure over loop variable.
		backend = backend

		Describe(fmt.Sprintf("dockerfile with %s backend", backend.name), func() {
			DescribeTable("should produce expected image configurations", checkFunc,
				Entry("Shell form ENTRYPOINT", cmdEntrypointTestOptions{setupEnvOptions: backend.options}, "dockerfile_shell_entrypoint", strslice.StrSlice{"/bin/sh", "-c", "echo \"ENTRYPOINT (shell)\""}, nil),
				Entry("Exec form ENTRYPOINT", cmdEntrypointTestOptions{setupEnvOptions: backend.options}, "dockerfile_exec_entrypoint", strslice.StrSlice{"echo \"ENTRYPOINT (exec)\""}, nil),
				Entry("Shell form CMD", cmdEntrypointTestOptions{setupEnvOptions: backend.options}, "dockerfile_shell_cmd", nil, strslice.StrSlice{"/bin/sh", "-c", "echo \"CMD (shell)\""}),
				Entry("Exec form CMD", cmdEntrypointTestOptions{setupEnvOptions: backend.options}, "dockerfile_exec_cmd", nil, strslice.StrSlice{"echo \"CMD (exec)\""}),
				Entry("No CMD, No ENTRYPOINT", cmdEntrypointTestOptions{setupEnvOptions: backend.options}, "dockerfile_no_cmd_no_entrypoint", nil, nil),
				Entry("Shell form ENTRYPOINT, reset CMD", cmdEntrypointTestOptions{setupEnvOptions: backend.options}, "dockerfile_entrypoint_reset_cmd", strslice.StrSlice{"/bin/sh", "-c", "echo \"ENTRYPOINT (shell)\""}, nil),
				Entry("Shell form ENTRYPOINT, Shell form CMD", cmdEntrypointTestOptions{setupEnvOptions: backend.options}, "dockerfile_entrypoint_cmd", strslice.StrSlice{"/bin/sh", "-c", "echo \"ENTRYPOINT (shell)\""}, strslice.StrSlice{"/bin/sh", "-c", "echo \"CMD (shell)\""}),
				Entry("Base image CMD", cmdEntrypointTestOptions{setupEnvOptions: backend.options}, "dockerfile_base_image_cmd", nil, strslice.StrSlice{"/bin/sh", "-c", "echo \"CMD (shell, base image)\""}),
			)

			if backend.name == "native-rootless" {
				// Dockerfile SHELL instruction is ignored by pure Buildah.
				// rel https://github.com/containers/buildah/issues/2959.
				DescribeTable("should produce expected image configurations", checkFunc,
					Entry("Custom shell", cmdEntrypointTestOptions{setupEnvOptions: backend.options}, "dockerfile_custom_shell_exec_cmd_and_entrypoint", strslice.StrSlice{"/bin/sh", "-c", "echo \"ENTRYPOINT (shell)\""}, strslice.StrSlice{"/bin/sh", "-c", "echo \"CMD (shell)\""}),
				)
			} else {
				DescribeTable("should produce expected image configurations", checkFunc,
					Entry("Custom shell", cmdEntrypointTestOptions{setupEnvOptions: backend.options}, "dockerfile_custom_shell_exec_cmd_and_entrypoint", strslice.StrSlice{"/bin/bash", "-c", "echo \"ENTRYPOINT (shell)\""}, strslice.StrSlice{"/bin/bash", "-c", "echo \"CMD (shell)\""}),
				)
			}
		})
	}
})
