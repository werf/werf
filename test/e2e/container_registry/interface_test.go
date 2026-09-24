package e2e_container_registry_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/docker_registry"
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
			})
		})
	}
})
