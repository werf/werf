package docker_instruction_test

import (
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"

	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/integration/utils"
	utilsDocker "github.com/werf/werf/integration/utils/docker"
)

type entry struct {
	werfYaml     string
	inspectCheck func(inspect *types.ImageInspect)
}

var itBody = func(e entry) {
	testDirPath = utils.FixturePath("base")
	stubs.SetEnv("WERF_CONFIG", filepath.Join(testDirPath, e.werfYaml))

	utils.RunSucceedCommand(
		testDirPath,
		werfBinPath,
		"build",
	)

	resultImageName := utils.SucceedCommandOutputString(
		testDirPath,
		werfBinPath,
		"stage", "image",
	)

	inspect := utilsDocker.ImageInspect(strings.TrimSpace(resultImageName))

	e.inspectCheck(inspect)
}

var _ = DescribeTable("docker instructions", itBody,
	Entry("volume", entry{
		werfYaml: "werf_volume.yaml",
		inspectCheck: func(inspect *types.ImageInspect) {
			Ω(inspect.Config.Volumes).Should(HaveKey("/test"))
		},
	}),
	Entry("expose", entry{
		werfYaml: "werf_expose.yaml",
		inspectCheck: func(inspect *types.ImageInspect) {
			Ω(inspect.Config.ExposedPorts).Should(HaveKey(nat.Port("80/udp")))
		},
	}),
	Entry("env", entry{
		werfYaml: "werf_env.yaml",
		inspectCheck: func(inspect *types.ImageInspect) {
			Ω(inspect.Config.Env).Should(ContainElement("TEST_NAME=test_value"))
		},
	}),
	Entry("label", entry{
		werfYaml: "werf_label.yaml",
		inspectCheck: func(inspect *types.ImageInspect) {
			Ω(inspect.Config.Labels).Should(HaveKeyWithValue("test_name", "test_value"))
		},
	}),
	Entry("entrypoint", entry{
		werfYaml: "werf_entrypoint.yaml",
		inspectCheck: func(inspect *types.ImageInspect) {
			Ω([]string(inspect.Config.Entrypoint)).Should(Equal([]string{"command", "param1", "param2"}))
		},
	}),
	Entry("cmd", entry{
		werfYaml: "werf_cmd.yaml",
		inspectCheck: func(inspect *types.ImageInspect) {
			Ω([]string(inspect.Config.Cmd)).Should(Equal([]string{"command", "param1", "param2"}))
		},
	}),
	Entry("workdir", entry{
		werfYaml: "werf_workdir.yaml",
		inspectCheck: func(inspect *types.ImageInspect) {
			Ω(inspect.Config.WorkingDir).Should(Equal("/test"))
		},
	}),
	Entry("user", entry{
		werfYaml: "werf_user.yaml",
		inspectCheck: func(inspect *types.ImageInspect) {
			Ω(inspect.Config.User).Should(Equal("test"))
		},
	}),
	Entry("healthcheck", entry{
		werfYaml: "werf_healthcheck.yaml",
		inspectCheck: func(inspect *types.ImageInspect) {
			Ω(inspect.Config.Healthcheck.Test).Should(Equal([]string{"CMD-SHELL", "true"}))
		},
	}))
