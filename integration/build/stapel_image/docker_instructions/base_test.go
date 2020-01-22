package docker_instruction_test

import (
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"

	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/flant/werf/pkg/testing/utils"
	utilsDocker "github.com/flant/werf/pkg/testing/utils/docker"
)

type entry struct {
	envs         map[string]string
	inspectCheck func(inspect *types.ImageInspect)
}

var itBody = func(e entry) {
	testDirPath = utils.FixturePath("base")

	for envName, envValue := range e.envs {
		stubs.SetEnv(envName, envValue)
	}

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
		envs: map[string]string{
			"DOCKER_VOLUME": "/test",
		},
		inspectCheck: func(inspect *types.ImageInspect) {
			Ω(inspect.Config.Volumes).Should(HaveKey("/test"))
		},
	}),
	Entry("expose", entry{
		envs: map[string]string{
			"DOCKER_EXPOSE": "80/udp",
		},
		inspectCheck: func(inspect *types.ImageInspect) {
			Ω(inspect.Config.ExposedPorts).Should(HaveKey(nat.Port("80/udp")))
		},
	}),
	Entry("env", entry{
		envs: map[string]string{
			"DOCKER_ENV_NAME":  "TEST_NAME",
			"DOCKER_ENV_VALUE": "test_value",
		},
		inspectCheck: func(inspect *types.ImageInspect) {
			Ω(inspect.Config.Env).Should(ContainElement("TEST_NAME=test_value"))
		},
	}),
	Entry("label", entry{
		envs: map[string]string{
			"DOCKER_LABEL_NAME":  "test_name",
			"DOCKER_LABEL_VALUE": "test_value",
		},
		inspectCheck: func(inspect *types.ImageInspect) {
			Ω(inspect.Config.Labels).Should(HaveKeyWithValue("test_name", "test_value"))
		},
	}),
	Entry("entrypoint", entry{
		envs: map[string]string{
			"DOCKER_ENTRYPOINT": "[\"command\", \"param1\", \"param2\"]",
		},
		inspectCheck: func(inspect *types.ImageInspect) {
			Ω([]string(inspect.Config.Entrypoint)).Should(Equal([]string{"command", "param1", "param2"}))
		},
	}),
	Entry("cmd", entry{
		envs: map[string]string{
			"DOCKER_CMD": "[\"command\", \"param1\", \"param2\"]",
		},
		inspectCheck: func(inspect *types.ImageInspect) {
			Ω([]string(inspect.Config.Cmd)).Should(Equal([]string{"command", "param1", "param2"}))
		},
	}),
	Entry("workdir", entry{
		envs: map[string]string{
			"DOCKER_WORKDIR": "/test",
		},
		inspectCheck: func(inspect *types.ImageInspect) {
			Ω(inspect.Config.WorkingDir).Should(Equal("/test"))
		},
	}),
	Entry("user", entry{
		envs: map[string]string{
			"DOCKER_USER": "test",
		},
		inspectCheck: func(inspect *types.ImageInspect) {
			Ω(inspect.Config.User).Should(Equal("test"))
		},
	}),
	Entry("healthcheck", entry{
		envs: map[string]string{
			"DOCKER_HEALTHCHECK": "CMD true",
		},
		inspectCheck: func(inspect *types.ImageInspect) {
			Ω(inspect.Config.Healthcheck.Test).Should(Equal([]string{"CMD-SHELL", "true"}))
		},
	}))
