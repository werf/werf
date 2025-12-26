package docker_instruction_test

import (
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/utils"
	utilsDocker "github.com/werf/werf/v2/test/pkg/utils/docker"
)

var _ = AfterEach(func(ctx SpecContext) {
	utils.RunSucceedCommand(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "host", "purge", "--force")
})

type entry struct {
	werfYaml     string
	inspectCheck func(inspect *types.ImageInspect)
}

var itBody = func(ctx SpecContext, e entry) {
	SuiteData.CommitProjectWorktree(ctx, SuiteData.ProjectName, utils.FixturePath("base"), "initial commit")

	SuiteData.Stubs.SetEnv("WERF_CONFIG", filepath.Join(SuiteData.GetProjectWorktree(SuiteData.ProjectName), e.werfYaml))

	utils.RunSucceedCommand(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "build")

	resultImageName := utils.SucceedCommandOutputString(
		ctx,
		SuiteData.GetProjectWorktree(SuiteData.ProjectName),
		SuiteData.WerfBinPath,
		"stage", "image",
	)

	inspect := utilsDocker.ImageInspect(ctx, strings.TrimSpace(resultImageName))

	e.inspectCheck(inspect)
}

var _ = DescribeTable("docker instructions", itBody,
	Entry("volume", entry{
		werfYaml: "werf_volume.yaml",
		inspectCheck: func(inspect *types.ImageInspect) {
			Expect(inspect.Config.Volumes).Should(HaveKey("/test"))
		},
	}),
	Entry("expose", entry{
		werfYaml: "werf_expose.yaml",
		inspectCheck: func(inspect *types.ImageInspect) {
			Expect(inspect.Config.ExposedPorts).Should(HaveKey(nat.Port("80/udp")))
		},
	}),
	Entry("env", entry{
		werfYaml: "werf_env.yaml",
		inspectCheck: func(inspect *types.ImageInspect) {
			Expect(inspect.Config.Env).Should(ContainElement("TEST_NAME=test_value"))
		},
	}),
	Entry("label", entry{
		werfYaml: "werf_label.yaml",
		inspectCheck: func(inspect *types.ImageInspect) {
			Expect(inspect.Config.Labels).Should(HaveKeyWithValue("test_name", "test_value"))
		},
	}),
	Entry("entrypoint", entry{
		werfYaml: "werf_entrypoint.yaml",
		inspectCheck: func(inspect *types.ImageInspect) {
			Expect([]string(inspect.Config.Entrypoint)).Should(Equal([]string{"command", "param1", "param2"}))
		},
	}),
	Entry("cmd", entry{
		werfYaml: "werf_cmd.yaml",
		inspectCheck: func(inspect *types.ImageInspect) {
			Expect([]string(inspect.Config.Cmd)).Should(Equal([]string{"command", "param1", "param2"}))
		},
	}),
	Entry("workdir", entry{
		werfYaml: "werf_workdir.yaml",
		inspectCheck: func(inspect *types.ImageInspect) {
			Expect(inspect.Config.WorkingDir).Should(Equal("/test"))
		},
	}),
	Entry("user", entry{
		werfYaml: "werf_user.yaml",
		inspectCheck: func(inspect *types.ImageInspect) {
			Expect(inspect.Config.User).Should(Equal("test"))
		},
	}),
	Entry("healthcheck", entry{
		werfYaml: "werf_healthcheck.yaml",
		inspectCheck: func(inspect *types.ImageInspect) {
			Expect(inspect.Config.Healthcheck.Test).Should(Equal([]string{"CMD-SHELL", "true"}))
		},
	}))
