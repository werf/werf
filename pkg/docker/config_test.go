package docker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/credentials"
	"github.com/docker/cli/cli/config/types"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("registry credentials", func() {
	var configDir string
	ginkgo.BeforeEach(func() {
		configDir = ginkgo.GinkgoT().TempDir()
		ginkgo.GinkgoT().Setenv("DOCKER_CONFIG", configDir)
		gomega.Expect(os.WriteFile(filepath.Join(configDir, "config.json"), []byte(`{"auths":{}}`), 0o600)).To(gomega.Succeed())
	})

	ginkgo.DescribeTable("stores credentials usable by the registry client and removes them on logout",
		func(address, repository, expectedKey string) {
			gomega.Expect(StoreCredentials(configDir, types.AuthConfig{
				ServerAddress: address, Username: "test-user", Password: "test-password",
			})).To(gomega.Succeed())
			conf, err := config.Load(configDir)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(conf.AuthConfigs).To(gomega.HaveKey(expectedKey))
			gomega.Expect(conf.AuthConfigs).To(gomega.HaveLen(1))
			ref, err := name.NewRepository(repository)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			authenticator, err := authn.DefaultKeychain.Resolve(ref)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			resolved, err := authenticator.Authorization()
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(resolved.Username).To(gomega.Equal("test-user"))
			gomega.Expect(resolved.Password).To(gomega.Equal("test-password"))

			gomega.Expect(EraseCredentials(configDir, types.AuthConfig{ServerAddress: address})).To(gomega.Succeed())
			authenticator, err = authn.DefaultKeychain.Resolve(ref)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(authenticator).To(gomega.Equal(authn.Anonymous))
			gomega.Expect(EraseCredentials(configDir, types.AuthConfig{ServerAddress: address})).To(gomega.Succeed())
		},
		ginkgo.Entry("Docker Hub index", "index.docker.io", "index.docker.io/test/image", "https://index.docker.io/v1/"),
		ginkgo.Entry("Docker Hub alias", "docker.io", "docker.io/test/image", "https://index.docker.io/v1/"),
		ginkgo.Entry("Docker Hub canonical key", "https://index.docker.io/v1/", "index.docker.io/test/image", "https://index.docker.io/v1/"),
		ginkgo.Entry("another registry", "registry.example.com:5000", "registry.example.com:5000/test/image", "registry.example.com:5000"),
	)

	ginkgo.DescribeTable("removes legacy Docker Hub credentials without touching other registries", func(address string) {
		conf, err := config.Load(configDir)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		for _, key := range []string{"https://index.docker.io/v1/", "index.docker.io", "docker.io", "registry.example.com"} {
			gomega.Expect(credentials.NewFileStore(conf).Store(types.AuthConfig{ServerAddress: key, Username: "test-user", Password: "test-password"})).To(gomega.Succeed())
		}
		gomega.Expect(EraseCredentials(configDir, types.AuthConfig{ServerAddress: address})).To(gomega.Succeed())
		conf, err = config.Load(configDir)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(conf.AuthConfigs).To(gomega.HaveLen(1))
		gomega.Expect(conf.AuthConfigs).To(gomega.HaveKey("registry.example.com"))
	},
		ginkgo.Entry("index", "index.docker.io"),
		ginkgo.Entry("alias", "docker.io"),
		ginkgo.Entry("canonical key", "https://index.docker.io/v1/"),
	)

	ginkgo.DescribeTable("preserves the Docker Hub credential helper", func(helperKey string) {
		helperDir := ginkgo.GinkgoT().TempDir()
		ginkgo.GinkgoT().Setenv("PATH", helperDir+string(os.PathListSeparator)+os.Getenv("PATH"))
		ginkgo.GinkgoT().Setenv("WERF_TEST_CREDENTIAL_FILE", filepath.Join(helperDir, "credentials.json"))
		helper := `#!/bin/sh
case "$1" in
 store) cat > "$WERF_TEST_CREDENTIAL_FILE" ;;
 get)
  read -r server
  file="$WERF_TEST_CREDENTIAL_FILE"
  if [ "$server" != "https://index.docker.io/v1/" ]; then file="$file.$server"; fi
  if [ -f "$file" ]; then
   cat "$file"
  else
   echo 'credentials not found in native keychain'
   exit 1
  fi ;;
 erase)
  read -r server
  file="$WERF_TEST_CREDENTIAL_FILE"
  if [ "$server" != "https://index.docker.io/v1/" ]; then file="$file.$server"; fi
  rm "$file" ;;
 *) exit 1 ;;
esac
`
		gomega.Expect(os.WriteFile(filepath.Join(helperDir, "docker-credential-werf-test"), []byte(helper), 0o700)).To(gomega.Succeed())
		gomega.Expect(os.WriteFile(filepath.Join(configDir, "config.json"), []byte(fmt.Sprintf(`{"auths":{},"credHelpers":{%q:"werf-test"}}`, helperKey)), 0o600)).To(gomega.Succeed())
		legacyFile := os.Getenv("WERF_TEST_CREDENTIAL_FILE") + "." + helperKey
		if helperKey == "mixed" {
			legacyHelper := strings.Replace(helper, "#!/bin/sh", "#!/bin/sh\nWERF_TEST_CREDENTIAL_FILE=\"$WERF_TEST_CREDENTIAL_FILE.legacy\"", 1)
			gomega.Expect(os.WriteFile(filepath.Join(helperDir, "docker-credential-werf-legacy-test"), []byte(legacyHelper), 0o700)).To(gomega.Succeed())
			gomega.Expect(os.WriteFile(filepath.Join(configDir, "config.json"), []byte(`{"auths":{},"credHelpers":{"https://index.docker.io/v1/":"werf-test","index.docker.io":"werf-legacy-test"}}`), 0o600)).To(gomega.Succeed())
			legacyFile = os.Getenv("WERF_TEST_CREDENTIAL_FILE") + ".legacy.index.docker.io"
		}
		if helperKey != "https://index.docker.io/v1/" {
			gomega.Expect(os.WriteFile(legacyFile, []byte(`{"Username":"old-user","Secret":"old-password"}`), 0o600)).To(gomega.Succeed())
		}
		gomega.Expect(StoreCredentials(configDir, types.AuthConfig{ServerAddress: "index.docker.io", Username: "test-user", Password: "test-password"})).To(gomega.Succeed())
		data, err := os.ReadFile(os.Getenv("WERF_TEST_CREDENTIAL_FILE"))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		var stored struct{ ServerURL, Username, Secret string }
		gomega.Expect(json.Unmarshal(data, &stored)).To(gomega.Succeed())
		gomega.Expect(stored.ServerURL).To(gomega.Equal("https://index.docker.io/v1/"))
		gomega.Expect(stored.Username).To(gomega.Equal("test-user"))
		gomega.Expect(stored.Secret).To(gomega.Equal("test-password"))
		conf, err := config.Load(configDir)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(conf.CredentialHelpers["https://index.docker.io/v1/"]).To(gomega.Equal("werf-test"))
		for _, credentials := range conf.AuthConfigs {
			gomega.Expect(credentials.Username).To(gomega.BeEmpty())
			gomega.Expect(credentials.Password).To(gomega.BeEmpty())
			gomega.Expect(credentials.Auth).To(gomega.BeEmpty())
		}
		ref, err := name.NewRepository("index.docker.io/test/image")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		authenticator, err := authn.DefaultKeychain.Resolve(ref)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		resolved, err := authenticator.Authorization()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(resolved.Username).To(gomega.Equal("test-user"))
		gomega.Expect(resolved.Password).To(gomega.Equal("test-password"))
		for _, address := range []string{"index.docker.io", "docker.io"} {
			gomega.Expect(credentials.NewFileStore(conf).Store(types.AuthConfig{ServerAddress: address, Username: "old-user", Password: "old-password"})).To(gomega.Succeed())
		}

		gomega.Expect(EraseCredentials(configDir, types.AuthConfig{ServerAddress: "docker.io"})).To(gomega.Succeed())
		conf, err = config.Load(configDir)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(conf.AuthConfigs).To(gomega.BeEmpty())
		_, err = os.Stat(os.Getenv("WERF_TEST_CREDENTIAL_FILE"))
		gomega.Expect(os.IsNotExist(err)).To(gomega.BeTrue())
		if helperKey != "https://index.docker.io/v1/" {
			_, err = os.Stat(legacyFile)
			gomega.Expect(os.IsNotExist(err)).To(gomega.BeTrue())
		}
	},
		ginkgo.Entry("canonical helper", "https://index.docker.io/v1/"),
		ginkgo.Entry("legacy index helper", "index.docker.io"),
		ginkgo.Entry("legacy alias helper", "docker.io"),
		ginkgo.Entry("distinct canonical and legacy helpers", "mixed"),
	)
})
