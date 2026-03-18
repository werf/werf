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
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(config).To(ContainSubstring(`location = "secure-mirror.example.com"`))
			Expect(config).To(ContainSubstring(`insecure = true`))
		})

		It("should mark https mirrors not in insecureRegistries as secure", func() {
			config, err := generateRegistriesConfig(
				[]string{"https://secure-mirror.example.com"},
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
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(config).To(ContainSubstring(`location = "localhost:5000"`))
			Expect(config).To(ContainSubstring(`insecure = true`))
			Expect(config).NotTo(ContainSubstring(`[[registry.mirror]]`))
		})

		It("should not duplicate registry when insecure registry is also a mirror target", func() {
			config, err := generateRegistriesConfig(
				[]string{"http://mirror.local:5000"},
				[]string{"mirror.local:5000"},
			)
			Expect(err).NotTo(HaveOccurred())
			// mirror is covered by [[registry.mirror]] entry — no standalone [[registry]] for it
			mirrorCount := strings.Count(config, `location = "mirror.local:5000"`)
			Expect(mirrorCount).To(Equal(1))
		})

		It("should generate standalone entries for multiple insecure non-mirror registries", func() {
			config, err := generateRegistriesConfig(
				[]string{},
				[]string{"registry-a.local:5000", "registry-b.local:5001"},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(config).To(ContainSubstring(`location = "registry-a.local:5000"`))
			Expect(config).To(ContainSubstring(`location = "registry-b.local:5001"`))
			insecureCount := strings.Count(config, "insecure = true")
			Expect(insecureCount).To(Equal(2))
		})
	})

	Describe("GetInsecureRegistriesFromConfig", func() {
		var tmpDir string
		var oldHome string

		BeforeEach(func() {
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

		It("should parse insecure mirror within registry", func() {
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
			Expect(result).To(ContainElement("mirror.example.com"))
		})
	})

	Describe("GetRegistryMirrorsFromConfig", func() {
		var tmpDir string
		var oldHome string

		BeforeEach(func() {
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
			Expect(result).To(ContainElement("https://dh-mirror.gitverse.ru"))
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
	})
})
