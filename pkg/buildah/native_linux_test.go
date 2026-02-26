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

			// Create config directory
			configDir := tmpDir + "/.config/containers"
			Expect(os.MkdirAll(configDir, 0755)).To(Succeed())
		})

		AfterEach(func() {
			os.Setenv("HOME", oldHome)
			os.RemoveAll(tmpDir)
		})

		It("should return empty list when config has no insecure registries", func() {
			// Create an empty but valid registries.conf so the function returns before checking /etc
			configPath := tmpDir + "/.config/containers/registries.conf"
			content := `# Empty registries config
`
			Expect(os.WriteFile(configPath, []byte(content), 0644)).To(Succeed())

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
			Expect(os.WriteFile(configPath, []byte(content), 0644)).To(Succeed())

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
			Expect(os.WriteFile(configPath, []byte(content), 0644)).To(Succeed())

			result, err := GetInsecureRegistriesFromConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(ContainElement("mirror.example.com"))
		})
	})
})
