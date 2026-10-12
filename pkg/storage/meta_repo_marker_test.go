package storage

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/image"
)

// markerGuardedMethods write project metadata, so metaRepoMarkerStorage must
// override each one to plant the marker before delegating.
var markerGuardedMethods = []string{
	"AddManagedImage",
	"PostLastCleanupRecord",
	"PutImageMetadata",
	"RegisterStageCustomTag",
	"RmImageMetadata",
	"RmManagedImage",
	"UnregisterStageCustomTag",
}

// markerUnguardedMethods either read, or act on stages rather than on the
// project metadata routed through the meta-repo.
var markerUnguardedMethods = []string{
	"AddStageCustomTag",
	"Address",
	"CheckStageCustomTag",
	"ConstructStageImageName",
	"CopyFromStorage",
	"CreateRepo",
	"DeleteRejectedStageImage",
	"DeleteRejectedStageRecord",
	"DeleteRepo",
	"DeleteStage",
	"DeleteStageCustomTag",
	"ExportStage",
	"FetchImage",
	"FilterStageDescSetAndProcessRelatedData",
	"GetAllAndGroupImageMetadataByImageName",
	"GetClientIDRecords",
	"GetImportMetadata",
	"GetImportMetadataIDs",
	"PutImportMetadata",
	"RmImportMetadata",
	"GetSyncServerRecords",
	"GetLastCleanupRecord",
	"GetManagedImages",
	"GetRejectedStageIDs",
	"GetStageCustomTagMetadata",
	"GetStageCustomTagMetadataIDs",
	"GetStageDesc",
	"GetStagesIDs",
	"GetStagesIDsByDigest",
	"IsImageMetadataExist",
	"IsManagedImageExist",
	"MutateAndPushImage",
	"PostClientIDRecord",
	"PostSyncServerRecord",
	"PostManifest",
	"PostMultiplatformImage",
	"RejectStage",
	"ShouldFetchImage",
	"StoreImage",
	"String",
}

var _ = ginkgo.Describe("meta-repo marker", func() {
	var reg *markerRegistry
	var stages *RepoStagesStorage

	ginkgo.BeforeEach(func() {
		reg = newMarkerRegistry()
		stages = newRepoStorage(reg, stagesRepo)
	})

	markerRef := func(project string) string {
		return fmt.Sprintf("%s:%s%s", stagesRepo, RepoMetaRepoMarker_ImageTagPrefix, getMetaRepoMarkerID(project))
	}

	ginkgo.Describe("CanonicalRepoAddress", func() {
		ginkgo.DescribeTable("resolves equivalent spellings to one address",
			func(address, expected string) {
				got, err := CanonicalRepoAddress(context.Background(), address)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(got).To(gomega.Equal(expected))
			},
			ginkgo.Entry("drops the tag", "registry.example/meta:v1", "registry.example/meta"),
			ginkgo.Entry("keeps a tagless address", "registry.example/meta", "registry.example/meta"),
			ginkgo.Entry("expands a docker hub shorthand", "myproject", "index.docker.io/library/myproject"),
			ginkgo.Entry("normalizes the docker hub host", "docker.io/myuser/myproject", "index.docker.io/myuser/myproject"),
			ginkgo.Entry("keeps a port", "localhost:5000/meta:v1", "localhost:5000/meta"),
		)

		ginkgo.It("rejects an unparseable address", func() {
			_, err := CanonicalRepoAddress(context.Background(), "registry.example/UPPERCASE")
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("parse repo address"))
		})

		ginkgo.It("rejects an address that is not a container registry address", func() {
			_, err := CanonicalRepoAddress(context.Background(), "archive:/tmp/repo.tar")
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("is not a container registry address"))
		})
	})

	ginkgo.Describe("marker record", func() {
		ginkgo.It("round-trips address and found flag", func(ctx ginkgo.SpecContext) {
			addr, found, err := stages.GetMetaRepoMarker(ctx, proj)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(found).To(gomega.BeFalse())
			gomega.Expect(addr).To(gomega.BeEmpty())

			gomega.Expect(stages.PutMetaRepoMarker(ctx, proj, "registry.example/meta")).To(gomega.Succeed())

			addr, found, err = stages.GetMetaRepoMarker(ctx, proj)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(found).To(gomega.BeTrue())
			gomega.Expect(addr).To(gomega.Equal("registry.example/meta"))
			gomega.Expect(reg.has(markerRef(proj))).To(gomega.BeTrue())
		})

		ginkgo.It("reports a malformed marker as found with empty address", func(ctx ginkgo.SpecContext) {
			reg.put(markerRef(proj), map[string]string{image.WerfLabel: proj})
			addr, found, err := stages.GetMetaRepoMarker(ctx, proj)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(found).To(gomega.BeTrue())
			gomega.Expect(addr).To(gomega.BeEmpty())
		})

		ginkgo.It("deletes idempotently", func(ctx ginkgo.SpecContext) {
			gomega.Expect(stages.RmMetaRepoMarker(ctx, proj)).To(gomega.Succeed())
			gomega.Expect(stages.PutMetaRepoMarker(ctx, proj, metaRepo)).To(gomega.Succeed())
			gomega.Expect(stages.RmMetaRepoMarker(ctx, proj)).To(gomega.Succeed())
			gomega.Expect(reg.has(markerRef(proj))).To(gomega.BeFalse())
		})

		ginkgo.It("refuses to delete a stage image occupying the marker tag", func(ctx ginkgo.SpecContext) {
			reg.putIndex(markerRef(proj))
			gomega.Expect(stages.RmMetaRepoMarker(ctx, proj)).NotTo(gomega.Succeed())
			gomega.Expect(reg.has(markerRef(proj))).To(gomega.BeTrue(), "a stage image at the marker tag must survive marker removal")
		})

		ginkgo.It("encodes the project with the managed-image slug (identity for tag-safe names)", func() {
			gomega.Expect(getMetaRepoMarkerID(proj)).To(gomega.Equal(proj))
		})
	})

	ginkgo.Describe("marker isolation from other namespaces", func() {
		ginkgo.It("is not classified as image-metadata by groupImageMetadataTagsByImageName", func(ctx ginkgo.SpecContext) {
			tags := []string{RepoMetaRepoMarker_ImageTagPrefix + getMetaRepoMarkerID(proj)}
			managed, notManaged, err := groupImageMetadataTagsByImageName(ctx, []string{"app"}, tags, RepoImageMetadataByCommitRecord_ImageTagPrefix)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(managed).To(gomega.BeEmpty())
			gomega.Expect(notManaged).To(gomega.BeEmpty())
		})

		ginkgo.It("is not a metadata candidate tag", func() {
			gomega.Expect(isMetadataCandidateTag(RepoMetaRepoMarker_ImageTagPrefix + getMetaRepoMarkerID(proj))).To(gomega.BeFalse())
		})

		ginkgo.It("gates metadata detection to exact record formats", func() {
			gomega.Expect(isMetadataCandidateTag("cleanup")).To(gomega.BeTrue())
			gomega.Expect(isMetadataCandidateTag("cleanup-foo")).To(gomega.BeFalse())
			gomega.Expect(isMetadataCandidateTag("meta-abc_commit_stage")).To(gomega.BeTrue())
			gomega.Expect(isMetadataCandidateTag("meta-foo")).To(gomega.BeFalse())
		})

		ginkgo.It("with a -rejected-suffixed project name is not returned by GetRejectedStageIDs", func(ctx ginkgo.SpecContext) {
			rejectedProj := "myproj-rejected"
			gomega.Expect(stages.PutMetaRepoMarker(ctx, rejectedProj, metaRepo)).To(gomega.Succeed())
			ids, err := stages.GetRejectedStageIDs(ctx)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(ids).To(gomega.BeEmpty())
		})
	})

	ginkgo.Describe("collectMetadataRecords", func() {
		collectedTags := func(ctx context.Context) []string {
			records, err := stages.collectMetadataRecords(ctx)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			tags := make([]string, 0, len(records))
			for _, rec := range records {
				tags = append(tags, rec.tag)
			}
			return tags
		}

		ginkgo.It("collects each family and skips the rejected-stage and marker tags", func(ctx ginkgo.SpecContext) {
			reg.put(stagesRepo+":managed-image-app", map[string]string{image.WerfLabel: proj})
			reg.put(stagesRepo+":meta-abc_commit_stage", map[string]string{image.WerfLabel: proj})
			reg.put(stagesRepo+":custom-tag-meta-v1", map[string]string{image.WerfLabel: proj})
			reg.put(stagesRepo+":cleanup", map[string]string{image.WerfLabel: proj, RepoCleanUpRecord_LabelTimestamp: "123"})
			reg.put(stagesRepo+":abc-123-rejected", map[string]string{image.WerfLabel: proj})
			reg.put(markerRef(proj), map[string]string{image.WerfLabel: proj, RepoMetaRepoMarker_LabelMetaRepo: metaRepo})

			gomega.Expect(collectedTags(ctx)).To(gomega.ConsistOf("managed-image-app", "meta-abc_commit_stage", "custom-tag-meta-v1", "cleanup"))
		})

		ginkgo.It("ignores a cleanup record without the timestamp label", func(ctx ginkgo.SpecContext) {
			reg.put(stagesRepo+":cleanup", map[string]string{image.WerfLabel: proj})
			gomega.Expect(collectedTags(ctx)).To(gomega.BeEmpty())
		})

		ginkgo.DescribeTable("does not consult the werf label, since a repository holds one project",
			func(ctx ginkgo.SpecContext, tag string, labels map[string]string) {
				reg.put(stagesRepo+":"+tag, labels)
				gomega.Expect(collectedTags(ctx)).To(gomega.ConsistOf(tag))
			},
			ginkgo.Entry("unlabeled image-metadata", "meta-abc_commit_stage", map[string]string{}),
			ginkgo.Entry("unlabeled managed-image", "managed-image-app", map[string]string{}),
			ginkgo.Entry("empty werf label", "managed-image-app", map[string]string{image.WerfLabel: ""}),
			ginkgo.Entry("foreign werf label", "meta-abc_commit_stage", map[string]string{image.WerfLabel: "otherproject"}),
		)
	})

	ginkgo.Describe("decorator", func() {
		var meta *RepoStagesStorage

		newDecorator := func(cleanupDisabled bool) *metaRepoMarkerStorage {
			meta = newRepoStorage(reg, metaRepo)
			return &metaRepoMarkerStorage{
				PrimaryStagesStorage: meta,
				markerStore:          stages,
				projectName:          proj,
				metaRepoAddress:      metaRepo,
				cleanupDisabled:      cleanupDisabled,
			}
		}

		ginkgo.It("plants the marker on a managed-image write", func(ctx ginkgo.SpecContext) {
			d := newDecorator(false)
			gomega.Expect(d.AddManagedImage(ctx, proj, "app")).To(gomega.Succeed())
			gomega.Expect(reg.has(markerRef(proj))).To(gomega.BeTrue())
		})

		ginkgo.It("plants the marker on a deletion-only first use", func(ctx ginkgo.SpecContext) {
			d := newDecorator(false)
			gomega.Expect(d.RmManagedImage(ctx, proj, "app")).To(gomega.Succeed())
			gomega.Expect(reg.has(markerRef(proj))).To(gomega.BeTrue())
		})

		ginkgo.It("plants on an enabled cleanup-record-only write", func(ctx ginkgo.SpecContext) {
			d := newDecorator(false)
			gomega.Expect(d.PostLastCleanupRecord(ctx, proj)).To(gomega.Succeed())
			gomega.Expect(reg.has(markerRef(proj))).To(gomega.BeTrue())
		})

		ginkgo.It("does not plant when cleanup is disabled", func(ctx ginkgo.SpecContext) {
			d := newDecorator(true)
			gomega.Expect(d.PostLastCleanupRecord(ctx, proj)).To(gomega.Succeed())
			gomega.Expect(reg.has(markerRef(proj))).To(gomega.BeFalse())
		})

		ginkgo.It("writes neither record nor marker when cleanup is disabled", func(ctx ginkgo.SpecContext) {
			d := newDecorator(true)
			meta.cleanupDisabled = true
			gomega.Expect(d.PostLastCleanupRecord(ctx, proj)).To(gomega.Succeed())
			gomega.Expect(reg.has(fmt.Sprintf(RepoCleanUpRecord_ImageNameFormat, metaRepo))).To(gomega.BeFalse())
			gomega.Expect(reg.has(markerRef(proj))).To(gomega.BeFalse())
		})

		ginkgo.It("writes both record and marker when cleanup is enabled", func(ctx ginkgo.SpecContext) {
			d := newDecorator(false)
			gomega.Expect(d.PostLastCleanupRecord(ctx, proj)).To(gomega.Succeed())
			gomega.Expect(reg.has(fmt.Sprintf(RepoCleanUpRecord_ImageNameFormat, metaRepo))).To(gomega.BeTrue())
			gomega.Expect(reg.has(markerRef(proj))).To(gomega.BeTrue())
		})

		ginkgo.It("does not plant on read-only access", func(ctx ginkgo.SpecContext) {
			d := newDecorator(false)
			_, err := d.GetManagedImages(ctx, proj)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(reg.has(markerRef(proj))).To(gomega.BeFalse())
		})

		ginkgo.It("fails when a conflicting marker already exists", func(ctx ginkgo.SpecContext) {
			gomega.Expect(stages.PutMetaRepoMarker(ctx, proj, "registry.example/other")).To(gomega.Succeed())
			d := newDecorator(false)
			err := d.AddManagedImage(ctx, proj, "app")
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("must match"))
		})

		ginkgo.It("plants exactly once across repeated and concurrent writes", func(ctx ginkgo.SpecContext) {
			d := newDecorator(false)
			var wg sync.WaitGroup
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer ginkgo.GinkgoRecover()
					gomega.Expect(d.AddManagedImage(ctx, proj, "app")).To(gomega.Succeed())
				}()
			}
			wg.Wait()
			gomega.Expect(reg.pushCount(markerRef(proj))).To(gomega.Equal(1))
		})

		ginkgo.DescribeTable("plants the marker on every metadata write",
			func(ctx ginkgo.SpecContext, write func(ctx context.Context, d *metaRepoMarkerStorage) error) {
				d := newDecorator(false)
				gomega.Expect(write(ctx, d)).To(gomega.Succeed())
				gomega.Expect(reg.has(markerRef(proj))).To(gomega.BeTrue())
			},
			ginkgo.Entry("AddManagedImage", func(ctx context.Context, d *metaRepoMarkerStorage) error {
				return d.AddManagedImage(ctx, proj, "app")
			}),
			ginkgo.Entry("RmManagedImage", func(ctx context.Context, d *metaRepoMarkerStorage) error {
				return d.RmManagedImage(ctx, proj, "app")
			}),
			ginkgo.Entry("PutImageMetadata", func(ctx context.Context, d *metaRepoMarkerStorage) error {
				return d.PutImageMetadata(ctx, proj, "app", "commit", "stage")
			}),
			ginkgo.Entry("RmImageMetadata", func(ctx context.Context, d *metaRepoMarkerStorage) error {
				return d.RmImageMetadata(ctx, proj, "app", "commit", "stage")
			}),
			ginkgo.Entry("RegisterStageCustomTag", func(ctx context.Context, d *metaRepoMarkerStorage) error {
				stageDesc := &image.StageDesc{StageID: image.NewStageID("digest", 1)}
				return d.RegisterStageCustomTag(ctx, proj, stageDesc, "release")
			}),
			ginkgo.Entry("UnregisterStageCustomTag", func(ctx context.Context, d *metaRepoMarkerStorage) error {
				reg.put(makeRepoCustomTagMetadataRecord(metaRepo, "release"), nil)
				return d.UnregisterStageCustomTag(ctx, "release")
			}),
			ginkgo.Entry("PostLastCleanupRecord", func(ctx context.Context, d *metaRepoMarkerStorage) error {
				return d.PostLastCleanupRecord(ctx, proj)
			}),
		)

		// metaRepoMarkerStorage embeds the PrimaryStagesStorage interface, so a
		// method it does not override is forwarded to the meta storage unwrapped
		// and writes metadata without planting the marker — and the compile-time
		// interface check still passes. Pin the method set so growing the
		// interface forces a decision about the new method.
		ginkgo.It("classifies every PrimaryStagesStorage method", func() {
			iface := reflect.TypeOf((*PrimaryStagesStorage)(nil)).Elem()

			methods := make([]string, 0, iface.NumMethod())
			for i := 0; i < iface.NumMethod(); i++ {
				methods = append(methods, iface.Method(i).Name)
			}

			gomega.Expect(methods).To(gomega.ConsistOf(append(markerGuardedMethods, markerUnguardedMethods...)),
				"PrimaryStagesStorage changed: a method that writes project metadata must be overridden in metaRepoMarkerStorage, covered by the table above and listed in markerGuardedMethods; anything else goes into markerUnguardedMethods")
		})
	})

	ginkgo.Describe("SetupMetaRepoSafeguard", func() {
		var meta *RepoStagesStorage
		ginkgo.BeforeEach(func() { meta = newRepoStorage(reg, metaRepo) })

		ginkgo.It("skips non-registry stages storage", func(ctx ginkgo.SpecContext) {
			local := NewLocalStagesStorage(nil)
			got, err := SetupMetaRepoSafeguard(ctx, proj, local, meta, false)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(got).To(gomega.BeIdenticalTo(meta))
		})

		ginkgo.It("legacy: no flag, no marker → passthrough", func(ctx ginkgo.SpecContext) {
			got, err := SetupMetaRepoSafeguard(ctx, proj, stages, stages, false)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(got).To(gomega.BeIdenticalTo(stages))
		})

		ginkgo.It("no flag but marker present → hard error", func(ctx ginkgo.SpecContext) {
			gomega.Expect(stages.PutMetaRepoMarker(ctx, proj, metaRepo)).To(gomega.Succeed())
			_, err := SetupMetaRepoSafeguard(ctx, proj, stages, stages, false)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("pass --meta-repo"))
		})

		ginkgo.It("rejects a malformed marker (present, no address)", func(ctx ginkgo.SpecContext) {
			reg.put(markerRef(proj), map[string]string{image.WerfLabel: proj})
			_, err := SetupMetaRepoSafeguard(ctx, proj, stages, meta, false)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("malformed"))
		})

		ginkgo.It("configured + matching marker → decorated", func(ctx ginkgo.SpecContext) {
			gomega.Expect(stages.PutMetaRepoMarker(ctx, proj, metaRepo)).To(gomega.Succeed())
			got, err := SetupMetaRepoSafeguard(ctx, proj, stages, meta, false)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			_, ok := got.(*metaRepoMarkerStorage)
			gomega.Expect(ok).To(gomega.BeTrue())
		})

		ginkgo.It("configured + mismatching marker → hard error", func(ctx ginkgo.SpecContext) {
			gomega.Expect(stages.PutMetaRepoMarker(ctx, proj, "registry.example/other")).To(gomega.Succeed())
			_, err := SetupMetaRepoSafeguard(ctx, proj, stages, meta, false)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("they must match"))
		})

		ginkgo.It("configured + no marker + metadata already in --repo → decorated", func(ctx ginkgo.SpecContext) {
			reg.put(stagesRepo+":managed-image-app", map[string]string{image.WerfLabel: proj})
			got, err := SetupMetaRepoSafeguard(ctx, proj, stages, meta, false)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			_, ok := got.(*metaRepoMarkerStorage)
			gomega.Expect(ok).To(gomega.BeTrue())
		})

		ginkgo.It("configured + no marker + clean repo → decorated", func(ctx ginkgo.SpecContext) {
			got, err := SetupMetaRepoSafeguard(ctx, proj, stages, meta, false)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			_, ok := got.(*metaRepoMarkerStorage)
			gomega.Expect(ok).To(gomega.BeTrue())
		})

		ginkgo.It("meta-repo equal to repo + marker to a distinct repo → hard error", func(ctx ginkgo.SpecContext) {
			gomega.Expect(stages.PutMetaRepoMarker(ctx, proj, metaRepo)).To(gomega.Succeed())
			sameMeta := newRepoStorage(reg, stagesRepo)
			_, err := SetupMetaRepoSafeguard(ctx, proj, stages, sameMeta, false)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("resolves to the same repository as --repo"))
			gomega.Expect(err.Error()).To(gomega.ContainSubstring(metaRepo))
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("werf meta-repo detach"))
		})

		ginkgo.It("refuses when the marker tag is occupied by a multi-platform stage image", func(ctx ginkgo.SpecContext) {
			reg.putIndex(markerRef(proj))
			_, err := SetupMetaRepoSafeguard(ctx, proj, stages, meta, false)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("reserves that tag for the meta-repo safeguard"))
			gomega.Expect(err.Error()).NotTo(gomega.ContainSubstring("malformed"))
		})

		ginkgo.It("refuses when the marker tag is occupied by a stage image", func(ctx ginkgo.SpecContext) {
			reg.put(markerRef(proj), map[string]string{
				image.WerfLabel:                   proj,
				image.WerfStageContentDigestLabel: "deadbeef",
			})
			_, err := SetupMetaRepoSafeguard(ctx, proj, stages, meta, false)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("reserves that tag for the meta-repo safeguard"))
		})

		ginkgo.It("meta-repo equal to repo + no marker → passthrough", func(ctx ginkgo.SpecContext) {
			sameMeta := newRepoStorage(reg, stagesRepo)
			got, err := SetupMetaRepoSafeguard(ctx, proj, stages, sameMeta, false)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			_, ok := got.(*metaRepoMarkerStorage)
			gomega.Expect(ok).To(gomega.BeFalse())
		})
	})

	ginkgo.Describe("MigrateMetaRepo", func() {
		var meta *RepoStagesStorage
		ginkgo.BeforeEach(func() {
			meta = newRepoStorage(reg, metaRepo)
			reg.put(stagesRepo+":managed-image-app", map[string]string{image.WerfLabel: proj})
			reg.put(stagesRepo+":meta-abc_commit_stage", map[string]string{image.WerfLabel: proj})
			reg.put(stagesRepo+":abc-123-rejected", map[string]string{image.WerfLabel: proj})
		})

		ginkgo.It("copies the metadata records, plants the marker, keeps the source", func(ctx ginkgo.SpecContext) {
			gomega.Expect(MigrateMetaRepo(ctx, proj, stages, meta, MigrateMetaRepoOptions{})).To(gomega.Succeed())
			gomega.Expect(reg.has(metaRepo + ":managed-image-app")).To(gomega.BeTrue())
			gomega.Expect(reg.has(metaRepo + ":meta-abc_commit_stage")).To(gomega.BeTrue())
			gomega.Expect(reg.has(metaRepo + ":abc-123-rejected")).To(gomega.BeFalse())
			gomega.Expect(reg.has(stagesRepo + ":managed-image-app")).To(gomega.BeTrue())
			addr, found, _ := stages.GetMetaRepoMarker(ctx, proj)
			gomega.Expect(found).To(gomega.BeTrue())
			gomega.Expect(addr).To(gomega.Equal(metaRepo))
		})

		ginkgo.DescribeTable("migrates metadata names ending in -rejected without moving rejected stages",
			func(ctx ginkgo.SpecContext, removeSource bool) {
				stageID := image.NewStageID("2604b86b2c7a1c6d19c62601aadb19e7d5c6bb8f17bc2bf26a390ea7", 1700000000000)
				gomega.Expect(stages.AddManagedImage(ctx, proj, "app-rejected")).To(gomega.Succeed())
				gomega.Expect(stages.PutImageMetadata(ctx, proj, "app-rejected", "commit", stageID.String())).To(gomega.Succeed())
				gomega.Expect(stages.RegisterStageCustomTag(ctx, proj, &image.StageDesc{StageID: stageID}, "release-rejected")).To(gomega.Succeed())
				rejectedTag := stageID.String() + RepoRejectedStageImageRecord_ImageTagSuffix
				reg.put(stagesRepo+":"+rejectedTag, map[string]string{image.WerfLabel: proj})

				gomega.Expect(MigrateMetaRepo(ctx, proj, stages, meta, MigrateMetaRepoOptions{RemoveSource: removeSource})).To(gomega.Succeed())

				managed, err := meta.GetManagedImages(ctx, proj)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(managed).To(gomega.ContainElement("app-rejected"))
				metadata, _, err := meta.GetAllAndGroupImageMetadataByImageName(ctx, proj, managed)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(metadata["app-rejected"]).To(gomega.Equal(map[string][]string{stageID.String(): {"commit"}}))
				customTag, err := meta.GetStageCustomTagMetadata(ctx, "release-rejected")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(customTag).To(gomega.Equal(&CustomTagMetadata{StageID: stageID.String(), Tag: "release-rejected"}))
				gomega.Expect(reg.has(stagesRepo + ":managed-image-app-rejected")).To(gomega.Equal(!removeSource))
				gomega.Expect(reg.has(stagesRepo + ":custom-tag-meta-release-rejected")).To(gomega.Equal(!removeSource))
				gomega.Expect(reg.has(stagesRepo + ":" + rejectedTag)).To(gomega.BeTrue())
				gomega.Expect(reg.has(metaRepo + ":" + rejectedTag)).To(gomega.BeFalse())
			},
			ginkgo.Entry("copy", false),
			ginkgo.Entry("move", true),
		)

		ginkgo.It("succeeds and plants the marker for a project with no metadata", func(ctx ginkgo.SpecContext) {
			emptyReg := newMarkerRegistry()
			src := newRepoStorage(emptyReg, stagesRepo)
			dst := newRepoStorage(emptyReg, metaRepo)
			gomega.Expect(MigrateMetaRepo(ctx, proj, src, dst, MigrateMetaRepoOptions{})).To(gomega.Succeed())
			addr, found, _ := src.GetMetaRepoMarker(ctx, proj)
			gomega.Expect(found).To(gomega.BeTrue())
			gomega.Expect(addr).To(gomega.Equal(metaRepo))
		})

		ginkgo.It("is idempotent and skips already-present destination records", func(ctx ginkgo.SpecContext) {
			reg.put(metaRepo+":managed-image-app", map[string]string{image.WerfLabel: proj})
			gomega.Expect(MigrateMetaRepo(ctx, proj, stages, meta, MigrateMetaRepoOptions{})).To(gomega.Succeed())
			gomega.Expect(reg.has(metaRepo + ":meta-abc_commit_stage")).To(gomega.BeTrue())
		})

		ginkgo.It("rejects a source marker pointing at a different meta-repo", func(ctx ginkgo.SpecContext) {
			gomega.Expect(stages.PutMetaRepoMarker(ctx, proj, "registry.example/other")).To(gomega.Succeed())
			err := MigrateMetaRepo(ctx, proj, stages, meta, MigrateMetaRepoOptions{})
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("refusing to migrate"))
		})

		ginkgo.It("refuses when --meta-repo resolves to the same repo as --repo and deletes nothing", func(ctx ginkgo.SpecContext) {
			sameMeta := newRepoStorage(reg, stagesRepo)
			err := MigrateMetaRepo(ctx, proj, stages, sameMeta, MigrateMetaRepoOptions{RemoveSource: true})
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("same repository"))
			gomega.Expect(reg.has(stagesRepo + ":managed-image-app")).To(gomega.BeTrue())
			gomega.Expect(reg.has(stagesRepo + ":meta-abc_commit_stage")).To(gomega.BeTrue())
		})

		ginkgo.It("refuses when a destination record is a stage image rather than metadata", func(ctx ginkgo.SpecContext) {
			reg.putIndex(metaRepo + ":managed-image-app")
			err := MigrateMetaRepo(ctx, proj, stages, meta, MigrateMetaRepoOptions{})
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("not a valid metadata record"))
		})

		ginkgo.It("refuses to delete source when the destination cleanup record lacks the timestamp", func(ctx ginkgo.SpecContext) {
			reg.put(stagesRepo+":cleanup", map[string]string{image.WerfLabel: proj, RepoCleanUpRecord_LabelTimestamp: "123"})
			reg.put(metaRepo+":cleanup", map[string]string{image.WerfLabel: proj})
			err := MigrateMetaRepo(ctx, proj, stages, meta, MigrateMetaRepoOptions{RemoveSource: true})
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("not a valid metadata record"))
			gomega.Expect(reg.has(stagesRepo + ":cleanup")).To(gomega.BeTrue())
		})

		ginkgo.It("keeps the destination cleanup record over the source one", func(ctx ginkgo.SpecContext) {
			reg.put(stagesRepo+":cleanup", map[string]string{image.WerfLabel: proj, RepoCleanUpRecord_LabelTimestamp: "999"})
			reg.put(metaRepo+":cleanup", map[string]string{image.WerfLabel: proj, RepoCleanUpRecord_LabelTimestamp: "111"})
			gomega.Expect(MigrateMetaRepo(ctx, proj, stages, meta, MigrateMetaRepoOptions{RemoveSource: true})).To(gomega.Succeed())
			img, err := reg.TryGetRepoImage(ctx, metaRepo+":cleanup")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(img).NotTo(gomega.BeNil())
			gomega.Expect(img.Labels[RepoCleanUpRecord_LabelTimestamp]).To(gomega.Equal("111"))
			gomega.Expect(reg.has(stagesRepo + ":cleanup")).To(gomega.BeFalse())
		})

		ginkgo.It("keeps the retargeted destination custom-tag metadata over the source one", func(ctx ginkgo.SpecContext) {
			reg.put(stagesRepo+":custom-tag-meta-tag1", map[string]string{
				image.WerfLabel:                         proj,
				image.WerfCustomTagMetadataStageIDLabel: "stage-b",
				image.WerfCustomTagMetadataTag:          "tag1",
			})
			reg.put(metaRepo+":custom-tag-meta-tag1", map[string]string{
				image.WerfLabel:                         proj,
				image.WerfCustomTagMetadataStageIDLabel: "stage-a",
				image.WerfCustomTagMetadataTag:          "tag1",
			})
			gomega.Expect(MigrateMetaRepo(ctx, proj, stages, meta, MigrateMetaRepoOptions{})).To(gomega.Succeed())
			img, err := reg.TryGetRepoImage(ctx, metaRepo+":custom-tag-meta-tag1")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(img).NotTo(gomega.BeNil())
			gomega.Expect(img.Labels[image.WerfCustomTagMetadataStageIDLabel]).To(gomega.Equal("stage-a"))
		})

		ginkgo.It("never migrates or deletes a multi-platform custom tag alias that collides with a metadata tag", func(ctx ginkgo.SpecContext) {
			reg.putIndex(stagesRepo + ":managed-image-release")
			reg.putIndex(stagesRepo + ":custom-tag-meta-release")
			reg.putIndex(stagesRepo + ":meta-a_b_c")

			gomega.Expect(MigrateMetaRepo(ctx, proj, stages, meta, MigrateMetaRepoOptions{RemoveSource: true})).To(gomega.Succeed())

			gomega.Expect(reg.has(stagesRepo+":managed-image-release")).To(gomega.BeTrue(), "an index reached through a colliding alias carries no labels and must never be deleted")
			gomega.Expect(reg.has(stagesRepo + ":custom-tag-meta-release")).To(gomega.BeTrue())
			gomega.Expect(reg.has(stagesRepo + ":meta-a_b_c")).To(gomega.BeTrue())
			gomega.Expect(reg.has(metaRepo + ":managed-image-release")).To(gomega.BeFalse())
		})

		ginkgo.It("never migrates or deletes a custom tag alias that collides with a metadata tag", func(ctx ginkgo.SpecContext) {
			stageLabels := map[string]string{
				image.WerfLabel:                   proj,
				image.WerfStageContentDigestLabel: "deadbeef",
			}
			reg.put(stagesRepo+":managed-image-release", stageLabels)
			reg.put(stagesRepo+":custom-tag-meta-release", stageLabels)
			reg.put(stagesRepo+":meta-a_b_c", stageLabels)

			gomega.Expect(MigrateMetaRepo(ctx, proj, stages, meta, MigrateMetaRepoOptions{RemoveSource: true})).To(gomega.Succeed())

			gomega.Expect(reg.has(stagesRepo+":managed-image-release")).To(gomega.BeTrue(), "a stage manifest reached through a colliding alias must never be deleted")
			gomega.Expect(reg.has(stagesRepo + ":custom-tag-meta-release")).To(gomega.BeTrue())
			gomega.Expect(reg.has(stagesRepo + ":meta-a_b_c")).To(gomega.BeTrue())
			gomega.Expect(reg.has(metaRepo+":managed-image-release")).To(gomega.BeFalse(), "and must not be copied into the meta-repo")
			gomega.Expect(reg.has(metaRepo + ":custom-tag-meta-release")).To(gomega.BeFalse())
			gomega.Expect(reg.has(metaRepo + ":meta-a_b_c")).To(gomega.BeFalse())
		})

		ginkgo.It("leaves source intact when a copy fails", func(ctx ginkgo.SpecContext) {
			reg.copyErr = fmt.Errorf("boom")
			err := MigrateMetaRepo(ctx, proj, stages, meta, MigrateMetaRepoOptions{RemoveSource: true})
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(reg.has(stagesRepo + ":managed-image-app")).To(gomega.BeTrue())
		})

		ginkgo.It("removes source after a verified copy", func(ctx ginkgo.SpecContext) {
			gomega.Expect(MigrateMetaRepo(ctx, proj, stages, meta, MigrateMetaRepoOptions{RemoveSource: true})).To(gomega.Succeed())
			gomega.Expect(reg.has(stagesRepo + ":managed-image-app")).To(gomega.BeFalse())
			gomega.Expect(reg.has(stagesRepo + ":meta-abc_commit_stage")).To(gomega.BeFalse())
			gomega.Expect(reg.has(metaRepo + ":managed-image-app")).To(gomega.BeTrue())
		})

		ginkgo.It("keeps the source when the destination is absent despite a reported copy", func(ctx ginkgo.SpecContext) {
			reg.copyNoop = true
			err := MigrateMetaRepo(ctx, proj, stages, meta, MigrateMetaRepoOptions{RemoveSource: true})
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("not deleting source"))
			gomega.Expect(reg.has(stagesRepo + ":managed-image-app")).To(gomega.BeTrue())
			gomega.Expect(reg.has(stagesRepo + ":meta-abc_commit_stage")).To(gomega.BeTrue())
		})

		ginkgo.It("moves an unlabeled legacy record like any other", func(ctx ginkgo.SpecContext) {
			reg.put(stagesRepo+":managed-image-legacy", map[string]string{})
			gomega.Expect(MigrateMetaRepo(ctx, proj, stages, meta, MigrateMetaRepoOptions{RemoveSource: true})).To(gomega.Succeed())
			gomega.Expect(reg.has(metaRepo + ":managed-image-legacy")).To(gomega.BeTrue())
			gomega.Expect(reg.has(stagesRepo + ":managed-image-legacy")).To(gomega.BeFalse())
		})

		ginkgo.It("copies an unlabeled legacy record without --remove-source", func(ctx ginkgo.SpecContext) {
			reg.put(stagesRepo+":managed-image-legacy", map[string]string{})
			gomega.Expect(MigrateMetaRepo(ctx, proj, stages, meta, MigrateMetaRepoOptions{})).To(gomega.Succeed())
			gomega.Expect(reg.has(metaRepo + ":managed-image-legacy")).To(gomega.BeTrue())
			gomega.Expect(reg.has(stagesRepo + ":managed-image-legacy")).To(gomega.BeTrue())
		})
	})
})
