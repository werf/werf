package buildah

import (
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/common-go/pkg/util"
)

var _ = Describe("buildah", func() {
	DescribeTable("mapBackendOldFiltersToBuildahImageFilters",
		func(oldFilters []util.Pair[string, string], expectedFilters []string) {
			actual := mapBackendOldFiltersToBuildahImageFilters(oldFilters)
			Expect(actual).To(Equal(expectedFilters))
		},
		Entry(
			"should work with empty input",
			[]util.Pair[string, string]{},
			[]string{},
		),
		Entry(
			"should work with non-empty input",
			[]util.Pair[string, string]{
				util.NewPair("foo", "bar"),
				util.NewPair("key", "value"),
			},
			[]string{"foo=bar", "key=value"},
		),
	)

	Describe("generateRegistriesConfig", func() {
		It("should mark http:// mirrors as insecure", func() {
			config, err := generateRegistriesConfig(
				[]string{"http://mirror.example.com"},
				[]string{},
				[]string{},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(config).To(ContainSubstring(`prefix = "docker.io"`))
			Expect(config).To(ContainSubstring(`location = "docker.io"`))
			Expect(config).To(ContainSubstring(`location = "mirror.example.com"`))
			Expect(config).To(ContainSubstring(`insecure = true`))
		})

		It("should mark https mirrors in insecureRegistries as insecure", func() {
			config, err := generateRegistriesConfig(
				[]string{"https://secure-mirror.example.com"},
				[]string{"secure-mirror.example.com"},
				[]string{},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(config).To(ContainSubstring(`location = "secure-mirror.example.com"`))
			Expect(config).To(ContainSubstring(`insecure = true`))
		})

		It("should mark https mirrors not in insecureRegistries as secure", func() {
			config, err := generateRegistriesConfig(
				[]string{"https://secure-mirror.example.com"},
				[]string{},
				[]string{},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(config).To(ContainSubstring(`location = "secure-mirror.example.com"`))
			Expect(config).To(ContainSubstring(`insecure = false`))
		})

		It("should handle multiple mirrors with mixed security", func() {
			config, err := generateRegistriesConfig(
				[]string{
					"http://insecure-http.example.com",
					"https://insecure-https.example.com",
					"https://secure.example.com",
				},
				[]string{"insecure-https.example.com"},
				[]string{},
			)
			Expect(err).NotTo(HaveOccurred())

			// Count insecure = true occurrences (should be 2)
			insecureCount := strings.Count(config, "insecure = true")
			secureCount := strings.Count(config, "insecure = false")
			Expect(insecureCount).To(Equal(2))
			Expect(secureCount).To(Equal(1))
		})

		It("should generate standalone [[registry]] entry for insecure non-mirror registry", func() {
			config, err := generateRegistriesConfig(
				[]string{},
				[]string{"localhost:5000"},
				[]string{"localhost:5000"},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(config).To(ContainSubstring(`location = "localhost:5000"`))
			Expect(config).To(ContainSubstring(`insecure = true`))
			Expect(config).NotTo(ContainSubstring(`[[registry.mirror]]`))
		})

		It("should not generate standalone registry when insecure host is used only as mirror", func() {
			config, err := generateRegistriesConfig(
				[]string{"http://mirror.local:5000"},
				[]string{"mirror.local:5000"},
				[]string{},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(strings.Count(config, `location = "mirror.local:5000"`)).To(Equal(1))
			Expect(config).To(ContainSubstring(`[[registry.mirror]]`))
		})

		It("should keep explicit standalone insecure registry when mirror host matches it", func() {
			config, err := generateRegistriesConfig(
				[]string{"https://local-registry.test:32768"},
				[]string{"local-registry.test:32768"},
				[]string{"local-registry.test:32768"},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(strings.Count(config, `location = "local-registry.test:32768"`)).To(Equal(2))
			Expect(config).To(ContainSubstring(`[[registry.mirror]]`))
			Expect(config).To(ContainSubstring("[[registry]]\nlocation = \"local-registry.test:32768\"\ninsecure = true"))
		})

		It("should generate standalone entries for multiple insecure non-mirror registries", func() {
			config, err := generateRegistriesConfig(
				[]string{},
				[]string{"registry-a.local:5000", "registry-b.local:5001"},
				[]string{"registry-a.local:5000", "registry-b.local:5001"},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(config).To(ContainSubstring(`location = "registry-a.local:5000"`))
			Expect(config).To(ContainSubstring(`location = "registry-b.local:5001"`))
			insecureCount := strings.Count(config, "insecure = true")
			Expect(insecureCount).To(Equal(2))
		})

		It("should deduplicate identical mirrors from different sources", func() {
			config, err := generateRegistriesConfig(
				[]string{"https://local-registry.test:32768", "https://local-registry.test:32768", "http://dh-mirror.gitverse.ru"},
				[]string{"local-registry.test:32768", "dh-mirror.gitverse.ru"},
				[]string{"local-registry.test:32768"},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(strings.Count(config, `[[registry.mirror]]`)).To(Equal(2))
			Expect(strings.Count(config, `location = "local-registry.test:32768"`)).To(Equal(2))
		})
	})

	Describe("GetInsecureRegistriesFromConfig", func() {
		var tmpDir string
		var oldHome string

		BeforeEach(func() {
			resetRegistriesConfCache()

			var err error
			tmpDir, err = os.MkdirTemp("", "registries-test")
			Expect(err).NotTo(HaveOccurred())

			oldHome = os.Getenv("HOME")
			os.Setenv("HOME", tmpDir)

			DeferCleanup(func() {
				os.Setenv("HOME", oldHome)
				os.RemoveAll(tmpDir)
			})

			configDir := tmpDir + "/.config/containers"
			Expect(os.MkdirAll(configDir, 0o755)).To(Succeed())
		})

		It("should return empty list when config has no insecure registries", func() {
			configPath := tmpDir + "/.config/containers/registries.conf"
			content := `# Empty registries config
`
			Expect(os.WriteFile(configPath, []byte(content), 0o644)).To(Succeed())

			result, err := GetInsecureRegistriesFromConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeEmpty())
		})

		It("should parse insecure registry", func() {
			configPath := tmpDir + "/.config/containers/registries.conf"
			content := `
[[registry]]
location = "localhost:5000"
insecure = true
`
			Expect(os.WriteFile(configPath, []byte(content), 0o644)).To(Succeed())

			result, err := GetInsecureRegistriesFromConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(ContainElement("localhost:5000"))
		})

		It("should not treat insecure mirror as standalone insecure registry", func() {
			configPath := tmpDir + "/.config/containers/registries.conf"
			content := `
[[registry]]
location = "docker.io"

[[registry.mirror]]
location = "mirror.example.com"
insecure = true
`
			Expect(os.WriteFile(configPath, []byte(content), 0o644)).To(Succeed())

			result, err := GetInsecureRegistriesFromConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeEmpty())
		})

		It("should not treat insecure relocation-style docker.io mirror as standalone insecure registry", func() {
			configPath := tmpDir + "/.config/containers/registries.conf"
			content := `
[[registry]]
prefix = "docker.io"
location = "dh-mirror.gitverse.ru"
insecure = true
`
			Expect(os.WriteFile(configPath, []byte(content), 0o644)).To(Succeed())

			result, err := GetInsecureRegistriesFromConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeEmpty())
		})
	})

	Describe("GetRegistryMirrorsFromConfig", func() {
		var tmpDir string
		var oldHome string

		BeforeEach(func() {
			resetRegistriesConfCache()

			var err error
			tmpDir, err = os.MkdirTemp("", "registries-mirrors-test")
			Expect(err).NotTo(HaveOccurred())

			oldHome = os.Getenv("HOME")
			os.Setenv("HOME", tmpDir)

			DeferCleanup(func() {
				os.Setenv("HOME", oldHome)
				os.RemoveAll(tmpDir)
			})

			configDir := tmpDir + "/.config/containers"
			Expect(os.MkdirAll(configDir, 0o755)).To(Succeed())
		})

		It("should return empty list when config has no mirrors", func() {
			configPath := tmpDir + "/.config/containers/registries.conf"
			content := `# Empty registries config
`
			Expect(os.WriteFile(configPath, []byte(content), 0o644)).To(Succeed())

			result, err := GetRegistryMirrorsFromConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeEmpty())
		})

		It("should parse docker.io mirrors", func() {
			configPath := tmpDir + "/.config/containers/registries.conf"
			content := `
[[registry]]
location = "docker.io"

[[registry.mirror]]
location = "mirror.example.com"

[[registry.mirror]]
location = "mirror2.example.com"
`
			Expect(os.WriteFile(configPath, []byte(content), 0o644)).To(Succeed())

			result, err := GetRegistryMirrorsFromConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(2))
			Expect(result).To(ContainElement("https://mirror.example.com"))
			Expect(result).To(ContainElement("https://mirror2.example.com"))
		})

		It("should ignore mirrors for non-docker.io registries", func() {
			configPath := tmpDir + "/.config/containers/registries.conf"
			content := `
[[registry]]
location = "quay.io"

[[registry.mirror]]
location = "quay-mirror.example.com"
`
			Expect(os.WriteFile(configPath, []byte(content), 0o644)).To(Succeed())

			result, err := GetRegistryMirrorsFromConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeEmpty())
		})

		It("should deduplicate mirrors", func() {
			configPath := tmpDir + "/.config/containers/registries.conf"
			content := `
[[registry]]
location = "docker.io"

[[registry.mirror]]
location = "mirror.example.com"

[[registry.mirror]]
location = "mirror.example.com"
`
			Expect(os.WriteFile(configPath, []byte(content), 0o644)).To(Succeed())

			result, err := GetRegistryMirrorsFromConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(1))
			Expect(result).To(ContainElement("https://mirror.example.com"))
		})

		It("should return nil when no config file exists", func() {
			result, err := GetRegistryMirrorsFromConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeNil())
		})

		It("should treat relocation as mirror (prefix=docker.io, location=mirror-host)", func() {
			configPath := tmpDir + "/.config/containers/registries.conf"
			content := `
[[registry]]
prefix = "docker.io"
location = "dh-mirror.gitverse.ru"
insecure = true
`
			Expect(os.WriteFile(configPath, []byte(content), 0o644)).To(Succeed())

			result, err := GetRegistryMirrorsFromConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(1))
			Expect(result).To(ContainElement("http://dh-mirror.gitverse.ru"))
		})

		It("should not treat docker.io location as mirror", func() {
			configPath := tmpDir + "/.config/containers/registries.conf"
			content := `
[[registry]]
prefix = "docker.io"
location = "docker.io"

[[registry.mirror]]
location = "mirror.example.com"
`
			Expect(os.WriteFile(configPath, []byte(content), 0o644)).To(Succeed())

			result, err := GetRegistryMirrorsFromConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(1))
			Expect(result).To(ContainElement("https://mirror.example.com"))
		})

		It("should keep insecure mirror as http scheme", func() {
			configPath := tmpDir + "/.config/containers/registries.conf"
			content := `
[[registry]]
location = "docker.io"

[[registry.mirror]]
location = "mirror.example.com"
insecure = true
`
			Expect(os.WriteFile(configPath, []byte(content), 0o644)).To(Succeed())

			result, err := GetRegistryMirrorsFromConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(1))
			Expect(result).To(ContainElement("http://mirror.example.com"))
		})

		It("should keep insecure relocation mirror as http scheme", func() {
			configPath := tmpDir + "/.config/containers/registries.conf"
			content := `
[[registry]]
prefix = "docker.io"
location = "dh-mirror.gitverse.ru"
insecure = true
`
			Expect(os.WriteFile(configPath, []byte(content), 0o644)).To(Succeed())

			result, err := GetRegistryMirrorsFromConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(1))
			Expect(result).To(ContainElement("http://dh-mirror.gitverse.ru"))
		})
	})

	Describe("registries.conf real-world docker.io mirror with standalone insecure registry", func() {
		var tmpDir string
		var oldHome string

		BeforeEach(func() {
			resetRegistriesConfCache()

			var err error
			tmpDir, err = os.MkdirTemp("", "registries-realworld-test")
			Expect(err).NotTo(HaveOccurred())

			oldHome = os.Getenv("HOME")
			os.Setenv("HOME", tmpDir)

			DeferCleanup(func() {
				os.Setenv("HOME", oldHome)
				os.RemoveAll(tmpDir)
			})

			configDir := tmpDir + "/.config/containers"
			Expect(os.MkdirAll(configDir, 0o755)).To(Succeed())
		})

		It("should parse docker.io mirror and standalone insecure registry from the same config", func() {
			configPath := tmpDir + "/.config/containers/registries.conf"
			content := `
[[registry]]
prefix = "docker.io"
location = "docker.io"

[[registry.mirror]]
location = "local-registry.test:32768"
insecure = true

[[registry]]
location = "local-registry.test:32768"
insecure = true
`
			Expect(os.WriteFile(configPath, []byte(content), 0o644)).To(Succeed())

			mirrors, err := GetRegistryMirrorsFromConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(mirrors).To(ContainElement("http://local-registry.test:32768"))

			insecureRegs, err := GetInsecureRegistriesFromConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(insecureRegs).To(ContainElement("local-registry.test:32768"))
		})
	})

	Describe("registries.conf path precedence", func() {
		var tmpDir string
		var oldHome string
		var oldContainersRegistriesConf string

		BeforeEach(func() {
			resetRegistriesConfCache()

			var err error
			tmpDir, err = os.MkdirTemp("", "registries-paths-test")
			Expect(err).NotTo(HaveOccurred())

			oldHome = os.Getenv("HOME")
			oldContainersRegistriesConf = os.Getenv("CONTAINERS_REGISTRIES_CONF")

			os.Setenv("HOME", tmpDir)
			os.Unsetenv("CONTAINERS_REGISTRIES_CONF")

			DeferCleanup(func() {
				os.Setenv("HOME", oldHome)
				os.Setenv("CONTAINERS_REGISTRIES_CONF", oldContainersRegistriesConf)
				os.RemoveAll(tmpDir)
			})

			Expect(os.MkdirAll(tmpDir+"/.config/containers", 0o755)).To(Succeed())
		})

		It("should read HOME config by default", func() {
			configPath := tmpDir + "/.config/containers/registries.conf"
			content := `
[[registry]]
location = "docker.io"

[[registry.mirror]]
location = "home-mirror.example.com"
`
			Expect(os.WriteFile(configPath, []byte(content), 0o644)).To(Succeed())

			result, err := GetRegistryMirrorsFromConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(ContainElement("https://home-mirror.example.com"))
		})

		It("should prefer CONTAINERS_REGISTRIES_CONF over HOME config", func() {
			homeConfigPath := tmpDir + "/.config/containers/registries.conf"
			envConfigPath := tmpDir + "/custom-registries.conf"
			Expect(os.WriteFile(homeConfigPath, []byte(`
[[registry]]
location = "docker.io"

[[registry.mirror]]
location = "home-mirror.example.com"
`), 0o644)).To(Succeed())
			Expect(os.WriteFile(envConfigPath, []byte(`
[[registry]]
location = "docker.io"

[[registry.mirror]]
location = "env-mirror.example.com"
`), 0o644)).To(Succeed())
			os.Setenv("CONTAINERS_REGISTRIES_CONF", envConfigPath)

			result, err := GetRegistryMirrorsFromConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(ContainElement("https://env-mirror.example.com"))
			Expect(result).NotTo(ContainElement("https://home-mirror.example.com"))
		})

		It("should use only CONTAINERS_REGISTRIES_CONF and its neighboring drop-in dir", func() {
			homeConfigPath := tmpDir + "/.config/containers/registries.conf"
			envConfigPath := tmpDir + "/custom-registries.conf"
			homeDropInDir := tmpDir + "/.config/containers/registries.conf.d"
			envDropInDir := envConfigPath + ".d"

			Expect(os.WriteFile(homeConfigPath, []byte(`# base home config is ignored when env config is set
`), 0o644)).To(Succeed())
			Expect(os.WriteFile(envConfigPath, []byte(`
[[registry]]
location = "docker.io"

[[registry.mirror]]
location = "env-mirror.example.com"
`), 0o644)).To(Succeed())
			Expect(os.MkdirAll(homeDropInDir, 0o755)).To(Succeed())
			Expect(os.WriteFile(homeDropInDir+"/10-insecure.conf", []byte(`
[[registry]]
location = "home-dropin-registry.example.com"
insecure = true
`), 0o644)).To(Succeed())
			Expect(os.MkdirAll(envDropInDir, 0o755)).To(Succeed())
			Expect(os.WriteFile(envDropInDir+"/10-insecure.conf", []byte(`
[[registry]]
location = "env-dropin-registry.example.com"
insecure = true
`), 0o644)).To(Succeed())
			os.Setenv("CONTAINERS_REGISTRIES_CONF", envConfigPath)

			mirrors, err := GetRegistryMirrorsFromConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(mirrors).To(ContainElement("https://env-mirror.example.com"))

			insecureRegs, err := GetInsecureRegistriesFromConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(insecureRegs).To(ContainElement("env-dropin-registry.example.com"))
			Expect(insecureRegs).NotTo(ContainElement("home-dropin-registry.example.com"))
		})

		It("should read mirrors from registries.conf.d drop-ins", func() {
			configPath := tmpDir + "/.config/containers/registries.conf"
			dropInDir := configPath + ".d"
			Expect(os.WriteFile(configPath, []byte(`
[[registry]]
location = "docker.io"
`), 0o644)).To(Succeed())
			Expect(os.MkdirAll(dropInDir, 0o755)).To(Succeed())
			Expect(os.WriteFile(dropInDir+"/10-mirror.conf", []byte(`
[[registry]]
location = "docker.io"

[[registry.mirror]]
location = "dropin-mirror.example.com"
insecure = true
`), 0o644)).To(Succeed())

			result, err := GetRegistryMirrorsFromConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(ContainElement("http://dropin-mirror.example.com"))
		})

		It("should read standalone insecure registries from registries.conf.d drop-ins", func() {
			configPath := tmpDir + "/.config/containers/registries.conf"
			dropInDir := configPath + ".d"
			Expect(os.WriteFile(configPath, []byte("# base\n"), 0o644)).To(Succeed())
			Expect(os.MkdirAll(dropInDir, 0o755)).To(Succeed())
			Expect(os.WriteFile(dropInDir+"/20-insecure.conf", []byte(`
[[registry]]
location = "dropin-registry.example.com"
insecure = true
`), 0o644)).To(Succeed())

			result, err := GetInsecureRegistriesFromConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(ContainElement("dropin-registry.example.com"))
		})

		It("should merge multiple registries.conf.d drop-ins in lexicographic order", func() {
			configPath := tmpDir + "/.config/containers/registries.conf"
			dropInDir := configPath + ".d"
			Expect(os.WriteFile(configPath, []byte(`
[[registry]]
location = "docker.io"
`), 0o644)).To(Succeed())
			Expect(os.MkdirAll(dropInDir, 0o755)).To(Succeed())
			Expect(os.WriteFile(dropInDir+"/10-first.conf", []byte(`
[[registry]]
location = "docker.io"

[[registry.mirror]]
location = "first-mirror.example.com"
`), 0o644)).To(Succeed())
			Expect(os.WriteFile(dropInDir+"/20-second.conf", []byte(`
[[registry]]
location = "docker.io"

[[registry.mirror]]
location = "second-mirror.example.com"
`), 0o644)).To(Succeed())

			result, err := GetRegistryMirrorsFromConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal([]string{"https://first-mirror.example.com", "https://second-mirror.example.com"}))
		})

		It("should deduplicate mirrors between base config and drop-ins", func() {
			configPath := tmpDir + "/.config/containers/registries.conf"
			dropInDir := configPath + ".d"
			Expect(os.WriteFile(configPath, []byte(`
[[registry]]
location = "docker.io"

[[registry.mirror]]
location = "same-mirror.example.com"
`), 0o644)).To(Succeed())
			Expect(os.MkdirAll(dropInDir, 0o755)).To(Succeed())
			Expect(os.WriteFile(dropInDir+"/10-duplicate.conf", []byte(`
[[registry]]
location = "docker.io"

[[registry.mirror]]
location = "same-mirror.example.com"
insecure = true
`), 0o644)).To(Succeed())

			result, err := GetRegistryMirrorsFromConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal([]string{"https://same-mirror.example.com"}))
		})

		It("should deduplicate insecure registries between base config and drop-ins", func() {
			configPath := tmpDir + "/.config/containers/registries.conf"
			dropInDir := configPath + ".d"
			Expect(os.WriteFile(configPath, []byte(`
[[registry]]
location = "same-registry.example.com"
insecure = true
`), 0o644)).To(Succeed())
			Expect(os.MkdirAll(dropInDir, 0o755)).To(Succeed())
			Expect(os.WriteFile(dropInDir+"/10-duplicate.conf", []byte(`
[[registry]]
location = "same-registry.example.com"
insecure = true
`), 0o644)).To(Succeed())

			result, err := GetInsecureRegistriesFromConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal([]string{"same-registry.example.com"}))
		})

		It("should fallback to registries.conf.d on invalid base registries.conf", func() {
			configPath := tmpDir + "/.config/containers/registries.conf"
			dropInDir := configPath + ".d"
			Expect(os.WriteFile(configPath, []byte(`
invalid config
`), 0o644)).To(Succeed())
			Expect(os.MkdirAll(dropInDir, 0o755)).To(Succeed())
			Expect(os.WriteFile(dropInDir+"/10-valid.conf", []byte(`
[[registry]]
location = "docker.io"

[[registry.mirror]]
location = "dropin-mirror.example.com"
`), 0o644)).To(Succeed())

			result, err := GetRegistryMirrorsFromConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(ContainElement("https://dropin-mirror.example.com"))
		})
	})
})
