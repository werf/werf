package common

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/werf/werf/v3/pkg/buildah"
)

var _ = ginkgo.Describe("List environment options", func() {
	ginkgo.BeforeEach(clearListEnv)
	ginkgo.It("keeps bare and suffixed values intact in lexical order", func() {
		for key, value := range map[string]string{"WERF_VALUES": "bare,a=b", "WERF_VALUES_2": "two", "WERF_VALUES_10": "ten", "WERF_VALUES_NAMED": "named", "WERF_VALUES_1": "", "WERF_VALUES_OTHEROPTION": "valid-named", "WERF_VALUESOTHER": "unrelated"} {
			ginkgo.GinkgoT().Setenv(key, value)
		}
		gomega.Expect(GetEnvValues(context.Background(), EnvValuesOptions{Name: "WERF_VALUES"})).To(gomega.Equal([]string{"bare,a=b", "ten", "", "two", "named", "valid-named"}))
	})
	ginkgo.It("ignores an empty bare value and excludes exact option names and their suffixes", func() {
		for key, value := range map[string]string{"WERF_EXPORT_ADD_LABEL": "", "WERF_EXPORT_ADD_LABEL_1": "key=value", "WERF_EXPORT_ADD_LABEL_SEPARATOR": ",", "WERF_EXPORT_ADD_LABEL_SEPARATOR_1": ";", "WERF_EXPORT_ADD_LABEL_SEPARATORLIKE": "named=value"} {
			ginkgo.GinkgoT().Setenv(key, value)
		}
		gomega.Expect(GetEnvValues(context.Background(), EnvValuesOptions{Name: "WERF_EXPORT_ADD_LABEL", Exclude: []string{"WERF_EXPORT_ADD_LABEL_SEPARATOR"}})).To(gomega.Equal([]string{"key=value", "named=value"}))
	})
	ginkgo.DescribeTable("routes deploy values only to their own option", func(name string, field func(*CmdData) *[]string) {
		ginkgo.GinkgoT().Setenv("WERF_SET", "ordinary=bare")
		ginkgo.GinkgoT().Setenv("WERF_SET_1", "ordinary=suffixed")
		ginkgo.GinkgoT().Setenv("WERF_SET_DOCKER_CONFIG_JSON_VALUE", "true")
		ginkgo.GinkgoT().Setenv(name, "bare")
		ginkgo.GinkgoT().Setenv(name+"_1", "suffixed")
		data := &CmdData{}
		*field(data) = []string{"cli"}
		gomega.Expect(data.processFlags()).To(gomega.Succeed())
		gomega.Expect(*field(data)).To(gomega.Equal([]string{"bare", "suffixed", "cli"}))
		gomega.Expect(data.ValuesSet).To(gomega.Equal([]string{"ordinary=bare", "ordinary=suffixed"}))
	},
		ginkgo.Entry("WERF_KUBE_IMPERSONATE_GROUP", "WERF_KUBE_IMPERSONATE_GROUP", func(data *CmdData) *[]string { return &data.KubeImpersonateGroups }),
		ginkgo.Entry("WERF_SET_STRING", "WERF_SET_STRING", func(data *CmdData) *[]string { return &data.ValuesSetString }),
		ginkgo.Entry("WERF_SET_FILE", "WERF_SET_FILE", func(data *CmdData) *[]string { return &data.ValuesSetFile }),
		ginkgo.Entry("WERF_SET_ROOT_JSON", "WERF_SET_ROOT_JSON", func(data *CmdData) *[]string { return &data.RootSetJSON }),
		ginkgo.Entry("WERF_SET_JSON", "WERF_SET_JSON", func(data *CmdData) *[]string { return &data.ValuesSetJSON }),
		ginkgo.Entry("WERF_SET_LITERAL", "WERF_SET_LITERAL", func(data *CmdData) *[]string { return &data.ValuesSetLiteral }),
		ginkgo.Entry("WERF_VALUES", "WERF_VALUES", func(data *CmdData) *[]string { return &data.ValuesFiles }),
		ginkgo.Entry("WERF_SECRET_VALUES", "WERF_SECRET_VALUES", func(data *CmdData) *[]string { return &data.SecretValuesFiles }),
		ginkgo.Entry("WERF_PATCHES", "WERF_PATCHES", func(data *CmdData) *[]string { return &data.PatchesFiles }),
		ginkgo.Entry("WERF_RESOURCE_VALIDATION_SKIP", "WERF_RESOURCE_VALIDATION_SKIP", func(data *CmdData) *[]string { return &data.ValidationSkip }),
		ginkgo.Entry("WERF_RESOURCE_VALIDATION_EXTRA_SCHEMA", "WERF_RESOURCE_VALIDATION_EXTRA_SCHEMA", func(data *CmdData) *[]string { return &data.ValidationExtraSchemas }),
		ginkgo.Entry("WERF_EXTRA_APIVERSIONS", "WERF_EXTRA_APIVERSIONS", func(data *CmdData) *[]string { return &data.ExtraAPIVersions }),
		ginkgo.Entry("WERF_LOOKUP_RESOURCES", "WERF_LOOKUP_RESOURCES", func(data *CmdData) *[]string { return &data.LocalLookupResourcesPaths }),
	)
	ginkgo.It("combines registry mirrors from environment, flags and the backend", func() {
		ginkgo.GinkgoT().Setenv("WERF_CONTAINER_REGISTRY_MIRROR", "bare.example")
		ginkgo.GinkgoT().Setenv("WERF_CONTAINER_REGISTRY_MIRROR_1", "https://suffix.example")
		data := &CmdData{}
		command := &cobra.Command{}
		SetupContainerRegistryMirror(data, command)
		gomega.Expect(command.ParseFlags([]string{"--container-registry-mirror=cli.example", "--container-registry-mirror=bare.example"})).To(gomega.Succeed())
		mirrors, err := GetContainerRegistryMirror(registryMirrorContext(), data, buildah.ModeDisabled)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(mirrors).To(gomega.Equal([]string{"https://bare.example", "https://suffix.example", "https://cli.example", "https://backend.example"}))
	})
	ginkgo.DescribeTable("combines build options with environment values", func(name string, get func(*CmdData) []string) {
		ginkgo.GinkgoT().Setenv(name, "bare")
		ginkgo.GinkgoT().Setenv(name+"_1", "suffixed")
		cli := []string{"cli"}
		data := &CmdData{KubeScanNamespaces: &cli, CacheStagesStorage: &cli, SecondaryStagesStorage: &cli, AddCustomTag: &cli, SSHKeys: &cli, DevIgnore: &cli, ExtraLabels: cli, ExtraAnnotations: cli, ReleaseLabels: cli}
		gomega.Expect(get(data)).To(gomega.Equal([]string{"bare", "suffixed", "cli"}))
	},
		ginkgo.Entry("namespace scan", "WERF_KUBE_SCAN_NAMESPACES", GetKubeScanNamespaces),
		ginkgo.Entry("cache repo", "WERF_CACHE_REPO", GetCacheStagesStorage),
		ginkgo.Entry("secondary repo", "WERF_SECONDARY_REPO", GetSecondaryStagesStorage),
		ginkgo.Entry("custom tag", "WERF_ADD_CUSTOM_TAG", getAddCustomTag),
		ginkgo.Entry("label", "WERF_ADD_LABEL", GetAddLabels),
		ginkgo.Entry("annotation", "WERF_ADD_ANNOTATION", GetAddAnnotations),
		ginkgo.Entry("release label", "WERF_RELEASE_LABEL", GetReleaseLabel),
		ginkgo.Entry("ssh key", "WERF_SSH_KEY", GetSSHKey),
		ginkgo.Entry("dev ignore", "WERF_DEV_IGNORE", GetDevIgnore),
	)
})
