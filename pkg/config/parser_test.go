package config

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const validConfigContent = `project: test-project
configVersion: 1
---
image: app
from: alpine:3.20
`

var _ = Describe("werf config source file path", func() {
	DescribeTable("is kept in the loaded config",
		func(ctx SpecContext, files map[string]string, customWerfConfigRelPath, expectedConfigPath string) {
			configPath, werfConfig, err := getWerfConfigFromProjectRepo(ctx, files, customWerfConfigRelPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(configPath).To(Equal(expectedConfigPath))
			Expect(werfConfig.SourceFilePath).To(Equal(expectedConfigPath))
		},
		Entry("default config", map[string]string{"werf.yaml": validConfigContent}, "", "werf.yaml"),
		Entry("custom config",
			map[string]string{"werf.yaml": validConfigContent, "another.werf.yaml": validConfigContent},
			"another.werf.yaml", "another.werf.yaml"),
	)

	DescribeTable("is reported in load errors",
		func(ctx SpecContext, files map[string]string, customWerfConfigRelPath, expectedConfigPath string) {
			_, _, err := getWerfConfigFromProjectRepo(ctx, files, customWerfConfigRelPath)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(HavePrefix(expectedConfigPath + ":"))
		},
		Entry("unknown field in the default config",
			map[string]string{"werf.yaml": validConfigContent + "unknownField: true\n"},
			"", "werf.yaml"),
		Entry("unknown field in the custom config",
			map[string]string{
				"werf.yaml":         validConfigContent,
				"another.werf.yaml": validConfigContent + "unknownField: true\n",
			},
			"another.werf.yaml", "another.werf.yaml"),
		Entry("invalid yaml in the custom config",
			map[string]string{
				"werf.yaml":         validConfigContent,
				"another.werf.yaml": validConfigContent + "  invalidIndent: true\n",
			},
			"another.werf.yaml", "another.werf.yaml"),
		Entry("template parse error in the custom config",
			map[string]string{
				"werf.yaml":         validConfigContent,
				"another.werf.yaml": validConfigContent + "{{ noSuchTemplateFunction }}\n",
			},
			"another.werf.yaml", "another.werf.yaml"),
		Entry("template execute error in the custom config",
			map[string]string{
				"werf.yaml":         validConfigContent,
				"another.werf.yaml": validConfigContent + "{{ .Env.NoSuchField }}\n",
			},
			"another.werf.yaml", "another.werf.yaml"),
	)

	It("keeps the rendered config file path and the failing lines in config errors", func(ctx SpecContext) {
		_, _, err := getWerfConfigFromProjectRepo(ctx, map[string]string{
			"werf.yaml":         validConfigContent,
			"another.werf.yaml": validConfigContent + "unknownField: true\n",
		}, "another.werf.yaml")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("config-render-"))
		Expect(err.Error()).To(ContainSubstring("unknownField"))
	})
})
