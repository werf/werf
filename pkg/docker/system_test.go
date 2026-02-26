package docker

import (
	"errors"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("docker system", func() {
	Describe("readDaemonConfigFromFile", func() {
		var tmpDir string
		var oldHome string

		BeforeEach(func() {
			var err error
			tmpDir, err = os.MkdirTemp("", "docker-config-test")
			Expect(err).NotTo(HaveOccurred())

			oldHome = os.Getenv("HOME")
			os.Setenv("HOME", tmpDir)
		})

		AfterEach(func() {
			os.Setenv("HOME", oldHome)
			os.RemoveAll(tmpDir)
		})

		It("should return empty config when daemon.json is empty object", func() {
			// Create an empty but valid daemon.json so the function returns before checking /etc
			dockerDir := filepath.Join(tmpDir, ".docker")
			Expect(os.MkdirAll(dockerDir, 0755)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(dockerDir, "daemon.json"), []byte("{}"), 0644)).To(Succeed())

			cfg, err := readDaemonConfigFromFile()
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg).NotTo(BeNil())
			Expect(cfg.RegistryMirrors).To(BeEmpty())
			Expect(cfg.InsecureRegistries).To(BeEmpty())
		})

		It("should read valid daemon.json", func() {
			dockerDir := filepath.Join(tmpDir, ".docker")
			Expect(os.MkdirAll(dockerDir, 0755)).To(Succeed())

			configContent := `{
				"registry-mirrors": ["https://mirror1.example.com", "http://mirror2.example.com"],
				"insecure-registries": ["my-registry.local:5000", "10.0.0.0/8"]
			}`
			Expect(os.WriteFile(filepath.Join(dockerDir, "daemon.json"), []byte(configContent), 0644)).To(Succeed())

			cfg, err := readDaemonConfigFromFile()
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg).NotTo(BeNil())
			Expect(cfg.RegistryMirrors).To(HaveLen(2))
			Expect(cfg.InsecureRegistries).To(HaveLen(2))
		})

		It("should return error for invalid JSON", func() {
			dockerDir := filepath.Join(tmpDir, ".docker")
			Expect(os.MkdirAll(dockerDir, 0755)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(dockerDir, "daemon.json"), []byte("not valid json"), 0644)).To(Succeed())

			cfg, err := readDaemonConfigFromFile()
			Expect(err).To(HaveOccurred())
			Expect(cfg).To(BeNil())
		})
	})

	Describe("isDaemonUnavailableErr", func() {
		It("should return false for nil error", func() {
			Expect(isDaemonUnavailableErr(nil)).To(BeFalse())
		})

		It("should return true for connection refused", func() {
			err := errors.New("connect: connection refused")
			Expect(isDaemonUnavailableErr(err)).To(BeTrue())
		})

		It("should return true for no such file", func() {
			err := errors.New("connect: no such file or directory")
			Expect(isDaemonUnavailableErr(err)).To(BeTrue())
		})

		It("should return true for cannot connect", func() {
			err := errors.New("Cannot connect to the Docker daemon")
			Expect(isDaemonUnavailableErr(err)).To(BeTrue())
		})

		It("should return false for other errors", func() {
			err := errors.New("some other error")
			Expect(isDaemonUnavailableErr(err)).To(BeFalse())
		})
	})
})
