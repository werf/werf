package e2e_container_registry_test

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
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
							ref, err := name.NewTag(repo + ":kept")
							Expect(err).NotTo(HaveOccurred())
							puller, err := remote.NewPuller(
								remote.WithAuthFromKeychain(authn.DefaultKeychain),
								remote.WithRetryBackoff(remote.Backoff{Steps: 1}),
							)
							Expect(err).NotTo(HaveOccurred())
							_, err = puller.Head(ctx, ref)
							Expect(err).NotTo(HaveOccurred())

							Expect(registry.DeleteRepo(ctx, repo)).To(Succeed())
							pollCtx, cancel := context.WithTimeout(ctx, time.Minute)
							defer cancel()
							registryHost := ref.Context().RegistryStr()
							if registryHost == name.DefaultRegistry {
								registryHost = "registry-1.docker.io"
							}
							delay := 5 * time.Second
							Eventually(func(ctx context.Context) error {
								_, err := puller.Head(ctx, ref)
								if err == nil {
									retry := TryAgainAfter(delay)
									delay = min(2*delay, 20*time.Second)
									return retry
								}
								var registryErr *transport.Error
								if errors.As(err, &registryErr) && registryErr.StatusCode == http.StatusNotFound &&
									registryErr.Request != nil && registryErr.Request.Method == http.MethodHead &&
									registryErr.Request.URL.Host == registryHost &&
									registryErr.Request.URL.Path == "/v2/"+ref.Context().RepositoryStr()+"/manifests/"+ref.Identifier() {
									return nil
								}
								return StopTrying("check Docker Hub manifest deletion").Wrap(err)
							}).WithContext(pollCtx).WithTimeout(time.Minute).Should(Succeed())
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
