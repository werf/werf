package e2e_build_test

import (
	"fmt"
	"slices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v3/test/pkg/contback"
	"github.com/werf/werf/v3/test/pkg/report"
	"github.com/werf/werf/v3/test/pkg/thirdparty/contruntime/strslice"
)

var _ = Describe("CMD and ENTRYPOINT combinations", Label("e2e", "build", "extra"), func() {
	type expectation struct {
		imageName  string
		entrypoint strslice.StrSlice
		cmd        strslice.StrSlice
	}

	commonExpectations := []expectation{
		{"dockerfile_shell_entrypoint", strslice.StrSlice{"/bin/sh", "-c", "echo \"ENTRYPOINT (shell)\""}, nil},
		{"dockerfile_exec_entrypoint", strslice.StrSlice{"echo \"ENTRYPOINT (exec)\""}, nil},
		{"dockerfile_shell_cmd", nil, strslice.StrSlice{"/bin/sh", "-c", "echo \"CMD (shell)\""}},
		{"dockerfile_exec_cmd", nil, strslice.StrSlice{"echo \"CMD (exec)\""}},
		{"dockerfile_no_cmd_no_entrypoint", nil, nil},
		{"dockerfile_entrypoint_reset_cmd", strslice.StrSlice{"/bin/sh", "-c", "echo \"ENTRYPOINT (shell)\""}, nil},
		{"dockerfile_entrypoint_cmd", strslice.StrSlice{"/bin/sh", "-c", "echo \"ENTRYPOINT (shell)\""}, strslice.StrSlice{"/bin/sh", "-c", "echo \"CMD (shell)\""}},
		{"dockerfile_base_image_cmd", nil, strslice.StrSlice{"/bin/sh", "-c", "echo \"CMD (shell, base image)\""}},
	}

	// Pure Buildah ignores the Dockerfile SHELL instruction, so the custom shell
	// never reaches the image config: https://github.com/containers/buildah/issues/2959.
	customShellExpectation := func(shell string) expectation {
		return expectation{
			imageName:  "dockerfile_custom_shell_exec_cmd_and_entrypoint",
			entrypoint: strslice.StrSlice{shell, "-c", "echo \"ENTRYPOINT (shell)\""},
			cmd:        strslice.StrSlice{shell, "-c", "echo \"CMD (shell)\""},
		}
	}

	backends := []struct {
		name         string
		options      setupEnvOptions
		expectations []expectation
	}{
		{
			name: "docker",
			options: setupEnvOptions{
				ContainerBackendMode:        "docker",
				WithLocalRepo:               false,
				WithStagedDockerfileBuilder: false,
			},
			expectations: slices.Concat(commonExpectations, []expectation{customShellExpectation("/bin/bash")}),
		},
		{
			name: "native-rootless",
			options: setupEnvOptions{
				ContainerBackendMode:        "native-rootless",
				WithLocalRepo:               true,
				WithStagedDockerfileBuilder: false,
			},
			expectations: slices.Concat(commonExpectations, []expectation{customShellExpectation("/bin/sh")}),
		},
		{
			name: "native-rootless-staged",
			options: setupEnvOptions{
				ContainerBackendMode:        "native-rootless",
				WithLocalRepo:               true,
				WithStagedDockerfileBuilder: true,
			},
			expectations: slices.Concat(commonExpectations, []expectation{customShellExpectation("/bin/bash")}),
		},
	}

	for _, backend := range backends {
		It(fmt.Sprintf("should produce expected image configurations with %s backend", backend.name), entryLabels(backend.options), func(ctx SpecContext) {
			repoDirname := "repo"

			By("initializing")
			setupEnv(backend.options)
			contRuntime := contback.NewContainerBackend(backend.options.ContainerBackendMode)

			By("preparing test repo")
			SuiteData.InitTestRepo(ctx, repoDirname, "cmd_entrypoint")

			By("building images")
			reportProject := report.NewProjectWithReport(newWerfProject(repoDirname))
			buildOut, buildReport := reportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath("report.json"), nil)
			Expect(buildOut).To(ContainSubstring("Building stage"))

			for _, e := range backend.expectations {
				By(fmt.Sprintf("checking metadata of %s", e.imageName))

				builtImage, found := buildReport.Images[e.imageName]
				Expect(found).To(BeTrue(), "%s is missing from the build report", e.imageName)

				imgCfg := contRuntime.GetImageInspect(ctx, builtImage.DockerImageName).Config

				if e.entrypoint == nil {
					Expect(imgCfg.Entrypoint).To(BeEmpty(), "%s: expected an empty Entrypoint", e.imageName)
				} else {
					Expect(imgCfg.Entrypoint).To(Equal(e.entrypoint), "%s: unexpected Entrypoint", e.imageName)
				}

				if e.cmd == nil {
					Expect(imgCfg.Cmd).To(BeEmpty(), "%s: expected an empty Cmd", e.imageName)
				} else {
					Expect(imgCfg.Cmd).To(Equal(e.cmd), "%s: unexpected Cmd", e.imageName)
				}
			}
		})
	}
})
