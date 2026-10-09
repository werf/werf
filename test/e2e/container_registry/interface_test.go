package e2e_container_registry_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/docker_registry"
	"github.com/werf/werf/v3/pkg/docker_registry/container_registry_extensions"
	"github.com/werf/werf/v3/test/pkg/suite_init"
)

const labelName = "werf-test-label"

var _ = Describe("container registry implementation", func() {
	for _, iName := range suite_init.ContainerRegistryImplementationListToCheck(true) {
		implementationName := iName

		Context("["+implementationName+"]", func() {
			It("should push, list, inspect, tag and delete images", func(ctx SpecContext) {
				implData := SuiteData.ContainerRegistryPerImplementation[implementationName]
				repo := fmt.Sprintf("%s/%s", implData.RegistryAddress, SuiteData.ProjectName)

				SuiteData.SetupRepo(ctx, repo, implementationName, SuiteData.StubsData)

				registry, err := docker_registry.NewDockerRegistry(ctx, repo, implData.WerfImplementationName, implData.RegistryOptions)
				Expect(err).ShouldNot(HaveOccurred())

				if implData.WerfImplementationName != docker_registry.DefaultImplementationName {
					DeferCleanup(func(ctx SpecContext) {
						By("deleting the repository")
						if implData.WerfImplementationName == docker_registry.DockerHubImplementationName {
							ref, err := name.NewRepository(repo)
							Expect(err).NotTo(HaveOccurred())
							client := &http.Client{
								Timeout: 15 * time.Second,
								CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
									return http.ErrUseLastResponse
								},
							}
							token, err := dockerHubToken(ctx, client, implData.RegistryOptions)
							Expect(err).NotTo(HaveOccurred())
							endpoint := "https://hub.docker.com/v2/repositories/" + ref.RepositoryStr() + "/"
							Expect(dockerHubRepositoryExists(ctx, client, endpoint, token)).To(BeTrue())

							Expect(registry.DeleteRepo(ctx, repo)).To(Succeed())
							pollCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
							defer cancel()
							delay := 30 * time.Second
							Eventually(func(ctx context.Context) (bool, error) {
								exists, err := dockerHubRepositoryExists(ctx, client, endpoint, token)
								if err != nil {
									return false, StopTrying("check Docker Hub repository deletion").Wrap(err)
								}
								if exists {
									retry := TryAgainAfter(delay)
									delay = min(2*delay, 2*time.Minute)
									return true, retry
								}
								return false, nil
							}).WithContext(pollCtx).WithTimeout(10 * time.Minute).Should(BeFalse())
							return
						}
						Expect(registry.DeleteRepo(ctx, repo)).To(Succeed())
						Expect(registry.TryGetRepoImage(ctx, repo+":kept")).To(BeNil())
					})
				}

				By("pushing two images")
				Expect(registry.PushImage(ctx, repo+":kept", &docker_registry.PushImageOptions{
					Labels: map[string]string{labelName: "kept"},
				})).To(Succeed())
				Expect(registry.PushImage(ctx, repo+":deleted", &docker_registry.PushImageOptions{
					Labels: map[string]string{labelName: "deleted"},
				})).To(Succeed())

				By("listing the tags of both")
				Expect(registry.Tags(ctx, repo)).To(ContainElements("kept", "deleted"))

				By("inspecting the image to delete")
				imgToDelete, err := registry.GetRepoImage(ctx, repo+":deleted")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(imgToDelete.Tag).To(Equal("deleted"))
				Expect(imgToDelete.RepoDigest).ToNot(BeEmpty())
				Expect(imgToDelete.Labels).To(HaveKeyWithValue(labelName, "deleted"))

				By("adding an alias tag to it")
				Expect(registry.TagRepoImage(ctx, imgToDelete, "deleted-alias")).To(Succeed())
				Expect(registry.Tags(ctx, repo)).To(ContainElement("deleted-alias"))

				By("deleting it")
				Expect(registry.DeleteRepoImage(ctx, imgToDelete)).To(Succeed())
				Expect(registry.Tags(ctx, repo)).To(And(
					Not(ContainElement("deleted")),
					ContainElement("kept"),
				))

				By("reporting the deleted image as absent")
				Expect(registry.TryGetRepoImage(ctx, repo+":deleted")).To(BeNil())

				By("keeping the untouched image readable")
				imgKept, err := registry.GetRepoImage(ctx, repo+":kept")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(imgKept.Labels).To(HaveKeyWithValue(labelName, "kept"))

				By("accepting the OCI format the scratch stage is published in")
				Expect(registry.PushImage(ctx, repo+":oci", &docker_registry.PushImageOptions{
					Labels:         map[string]string{labelName: "oci"},
					ManifestFormat: container_registry_extensions.ManifestFormatOCI,
				})).To(Succeed())
				imgOCI, err := registry.GetRepoImage(ctx, repo+":oci")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(imgOCI.Labels).To(HaveKeyWithValue(labelName, "oci"))

				manifest := getManifest(ctx, repo+":oci", implData.RegistryOptions)
				Expect(manifest.MediaType).To(Equal(types.OCIManifestSchema1))
				Expect(manifest.Config.MediaType).To(Equal(types.OCIConfigJSON))
				Expect(manifest.Layers).To(HaveLen(1))
				Expect(manifest.Layers[0].MediaType).To(Equal(types.OCILayer))
			})
		})
	}
})

func getManifest(ctx context.Context, reference string, registryOptions docker_registry.DockerRegistryOptions) *v1.Manifest {
	GinkgoHelper()

	nameOptions := []name.Option{name.WeakValidation}
	if registryOptions.InsecureRegistry {
		nameOptions = append(nameOptions, name.Insecure)
	}
	ref, err := name.ParseReference(reference, nameOptions...)
	Expect(err).ShouldNot(HaveOccurred())

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: registryOptions.SkipTlsVerifyRegistry}
	desc, err := remote.Get(ref,
		remote.WithContext(ctx),
		remote.WithAuthFromKeychain(authn.DefaultKeychain),
		remote.WithTransport(transport),
	)
	Expect(err).ShouldNot(HaveOccurred())

	img, err := desc.Image()
	Expect(err).ShouldNot(HaveOccurred())

	manifest, err := img.Manifest()
	Expect(err).ShouldNot(HaveOccurred())

	return manifest
}

var _ = DescribeTable("Docker Hub repository status", func(ctx SpecContext, status int, exists, fails bool) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requests.Add(1)
		response.WriteHeader(status)
	}))
	defer server.Close()
	actual, err := dockerHubRepositoryExists(ctx, server.Client(), server.URL+"/v2/repositories/account/project/", "test-token")
	Expect(requests.Load()).To(Equal(int32(1)))
	if fails {
		Expect(err).To(HaveOccurred())
		return
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(actual).To(Equal(exists))
},
	Entry("existing or pending repository", http.StatusOK, true, false),
	Entry("deleted repository", http.StatusNotFound, false, false),
	Entry("unauthorized is not deletion", http.StatusUnauthorized, false, true),
	Entry("forbidden is not deletion", http.StatusForbidden, false, true),
	Entry("rate limit is not deletion or retried", http.StatusTooManyRequests, false, true),
	Entry("server error is not deletion", http.StatusInternalServerError, false, true),
)
