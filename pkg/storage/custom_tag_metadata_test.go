package storage

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/image"
)

var _ = Describe("custom tag metadata", func() {
	const (
		metaRepo    = "registry.example.com/app/meta"
		installRepo = "registry.example.com/app/install"
	)

	DescribeTable("labels carry the stage id, the tag and the repo",
		func(tag CustomTag) {
			metadata := newCustomTagMetadata("ff0011-1748001122334", tag)
			restored := newCustomTagMetadataFromLabels(metadata.ToLabels())

			Expect(restored).To(Equal(metadata))
			Expect(restored.CustomTag()).To(Equal(tag))
		},
		Entry("primary repo", CustomTag{Tag: "main"}),
		Entry("other repo", CustomTag{Repo: installRepo, Tag: "main"}),
	)

	It("reads a record written before the repo label existed as a primary-repo tag", func() {
		restored := newCustomTagMetadataFromLabels(map[string]string{
			image.WerfCustomTagMetadataStageIDLabel: "ff0011-1748001122334",
			image.WerfCustomTagMetadataTag:          "main",
		})

		Expect(restored.CustomTag()).To(Equal(CustomTag{Tag: "main"}))
	})

	Describe("record id", func() {
		It("keeps the pre-repo id for primary-repo tags", func() {
			Expect(CustomTag{Tag: "main"}.metadataID()).To(Equal("main"))
		})

		DescribeTable("never lets two different tags share a record",
			func(a, b CustomTag) {
				Expect(a.metadataID()).NotTo(Equal(b.metadataID()))
			},
			Entry("same tag, primary vs other repo", CustomTag{Tag: "main"}, CustomTag{Repo: installRepo, Tag: "main"}),
			Entry("same tag, two other repos", CustomTag{Repo: installRepo, Tag: "main"}, CustomTag{Repo: "registry.example.com/app/tests", Tag: "main"}),
			Entry("repo and tag joined the same way", CustomTag{Repo: "registry.example.com/app-install", Tag: "main"}, CustomTag{Repo: "registry.example.com/app", Tag: "install-main"}),
		)

		It("is the same for equivalent spellings of the repo", func() {
			Expect(CustomTag{Repo: "docker.io/acme/install", Tag: "main"}.metadataID()).
				To(Equal(CustomTag{Repo: "index.docker.io/acme/install", Tag: "main"}.metadataID()))
		})

		It("stays a valid tag for a long repo and tag", func() {
			id := CustomTag{Repo: "registry.example.com/" + fmt.Sprintf("%0120d", 0), Tag: fmt.Sprintf("%0120d", 1)}.metadataID()
			Expect(len(id)).To(BeNumerically("<=", 48))
			Expect(fmt.Sprintf(RepoCustomTagMetadata_ImageNameFormat, metaRepo, id)).To(HaveSuffix(id))
		})
	})

	Describe("registration in the meta repo", func() {
		var (
			reg     *fakeRegistry
			storage *RepoStagesStorage
			primary = CustomTag{Tag: "main"}
			install = CustomTag{Repo: installRepo, Tag: "main"}
		)

		BeforeEach(func() {
			reg = newFakeRegistry()
			storage = &RepoStagesStorage{RepoAddress: metaRepo, DockerRegistry: reg}
		})

		It("writes one record per repo for the same tag", func() {
			stageDesc := &image.StageDesc{StageID: image.NewStageID("ff0011", 1748001122334)}

			Expect(storage.RegisterStageCustomTag(context.Background(), "proj", stageDesc, primary)).To(Succeed())
			Expect(storage.RegisterStageCustomTag(context.Background(), "proj", stageDesc, install)).To(Succeed())

			Expect(reg.pushCall).To(Equal(map[string]int{
				makeRepoCustomTagMetadataRecord(metaRepo, primary.metadataID()): 1,
				makeRepoCustomTagMetadataRecord(metaRepo, install.metadataID()): 1,
			}))
		})

		It("unregisters only the record of the given repo", func() {
			primaryRecord := makeRepoCustomTagMetadataRecord(metaRepo, primary.metadataID())
			installRecord := makeRepoCustomTagMetadataRecord(metaRepo, install.metadataID())
			reg.repoImages = map[string]*image.Info{
				primaryRecord: {Name: primaryRecord},
				installRecord: {Name: installRecord},
			}

			Expect(storage.UnregisterStageCustomTag(context.Background(), install)).To(Succeed())

			Expect(reg.deleteCall).To(Equal(map[string]int{installRecord: 1}))
		})
	})
})

var _ = Describe("SameRepoAddress", func() {
	DescribeTable("compares repositories canonically",
		func(a, b string, expected bool) {
			Expect(SameRepoAddress(a, b)).To(Equal(expected))
		},
		Entry("identical", "registry.example.com/app", "registry.example.com/app", true),
		Entry("docker hub spellings", "docker.io/acme/app", "index.docker.io/acme/app", true),
		Entry("different repos", "registry.example.com/app", "registry.example.com/app/install", false),
		Entry("empty vs set", "", "registry.example.com/app", false),
		Entry("both empty", "", "", true),
	)
})
