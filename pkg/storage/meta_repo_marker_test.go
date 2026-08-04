package storage

import (
	"context"
	"fmt"
	"strings"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/image"
)

type markerRegistry struct {
	docker_registry.Interface

	mu         sync.Mutex
	images     map[string]*image.Info
	pushCounts map[string]int
	copyErr    error
}

func newMarkerRegistry() *markerRegistry {
	return &markerRegistry{
		images:     map[string]*image.Info{},
		pushCounts: map[string]int{},
	}
}

func refTag(reference string) string {
	i := strings.LastIndex(reference, ":")
	if i < 0 {
		return ""
	}
	return reference[i+1:]
}

func refRepo(reference string) string {
	i := strings.LastIndex(reference, ":")
	if i < 0 {
		return reference
	}
	return reference[:i]
}

func (r *markerRegistry) PushImage(_ context.Context, reference string, opts *docker_registry.PushImageOptions) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	labels := map[string]string{}
	if opts != nil {
		for k, v := range opts.Labels {
			labels[k] = v
		}
	}
	r.images[reference] = &image.Info{Name: reference, Tag: refTag(reference), Labels: labels}
	r.pushCounts[reference]++
	return nil
}

func (r *markerRegistry) TryGetRepoImage(_ context.Context, reference string) (*image.Info, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.images[reference], nil
}

func (r *markerRegistry) GetRepoImage(_ context.Context, reference string) (*image.Info, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.images[reference], nil
}

func (r *markerRegistry) IsTagExist(_ context.Context, reference string, _ ...docker_registry.Option) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.images[reference]
	return ok, nil
}

func (r *markerRegistry) DeleteRepoImage(_ context.Context, repoImage *image.Info) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.images, repoImage.Name)
	return nil
}

func (r *markerRegistry) Tags(_ context.Context, reference string, _ ...docker_registry.Option) ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var res []string
	for name := range r.images {
		if refRepo(name) == reference {
			res = append(res, refTag(name))
		}
	}
	return res, nil
}

func (r *markerRegistry) CopyImage(_ context.Context, sourceReference, destinationReference string, _ docker_registry.CopyImageOptions) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.copyErr != nil {
		return r.copyErr
	}
	src := r.images[sourceReference]
	if src == nil {
		return fmt.Errorf("source %q not found", sourceReference)
	}
	labels := map[string]string{}
	for k, v := range src.Labels {
		labels[k] = v
	}
	r.images[destinationReference] = &image.Info{Name: destinationReference, Tag: refTag(destinationReference), Labels: labels}
	return nil
}

func (r *markerRegistry) put(reference string, labels map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.images[reference] = &image.Info{Name: reference, Tag: refTag(reference), Labels: labels}
}

func (r *markerRegistry) pushCount(reference string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.pushCounts[reference]
}

func (r *markerRegistry) has(reference string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.images[reference]
	return ok
}

const (
	stagesRepo = "registry.example/stages"
	metaRepo   = "registry.example/meta"
	proj       = "myproject"
)

func newRepoStorage(reg docker_registry.Interface, address string) *RepoStagesStorage {
	return &RepoStagesStorage{RepoAddress: address, DockerRegistry: reg, skipMetaCheck: true}
}

var _ = Describe("meta-repo marker", func() {
	var reg *markerRegistry
	var stages *RepoStagesStorage

	BeforeEach(func() {
		reg = newMarkerRegistry()
		stages = newRepoStorage(reg, stagesRepo)
	})

	markerRef := func(project string) string {
		return fmt.Sprintf("%s:%s%s", stagesRepo, RepoMetaRepoMarker_ImageTagPrefix, getMetaRepoMarkerID(project))
	}

	Describe("marker record", func() {
		It("round-trips address and found flag", func(ctx SpecContext) {
			addr, found, err := stages.GetMetaRepoMarker(ctx, proj)
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeFalse())
			Expect(addr).To(BeEmpty())

			Expect(stages.PutMetaRepoMarker(ctx, proj, "registry.example/meta")).To(Succeed())

			addr, found, err = stages.GetMetaRepoMarker(ctx, proj)
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(addr).To(Equal("registry.example/meta"))
			Expect(reg.has(markerRef(proj))).To(BeTrue())
		})

		It("reports a malformed marker as found with empty address", func(ctx SpecContext) {
			reg.put(markerRef(proj), map[string]string{image.WerfLabel: proj})
			addr, found, err := stages.GetMetaRepoMarker(ctx, proj)
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(addr).To(BeEmpty())
		})

		It("deletes idempotently", func(ctx SpecContext) {
			Expect(stages.RmMetaRepoMarker(ctx, proj)).To(Succeed())
			Expect(stages.PutMetaRepoMarker(ctx, proj, metaRepo)).To(Succeed())
			Expect(stages.RmMetaRepoMarker(ctx, proj)).To(Succeed())
			Expect(reg.has(markerRef(proj))).To(BeFalse())
		})

		It("encodes the project with the managed-image slug (identity for tag-safe names)", func() {
			Expect(getMetaRepoMarkerID(proj)).To(Equal(proj))
		})
	})

	Describe("marker isolation from other namespaces", func() {
		It("is not classified as image-metadata by groupImageMetadataTagsByImageName", func(ctx SpecContext) {
			tags := []string{RepoMetaRepoMarker_ImageTagPrefix + getMetaRepoMarkerID(proj)}
			managed, notManaged, err := groupImageMetadataTagsByImageName(ctx, []string{"app"}, tags, RepoImageMetadataByCommitRecord_ImageTagPrefix)
			Expect(err).NotTo(HaveOccurred())
			Expect(managed).To(BeEmpty())
			Expect(notManaged).To(BeEmpty())
		})

		It("is not a metadata candidate tag", func() {
			Expect(isMetadataCandidateTag(RepoMetaRepoMarker_ImageTagPrefix + getMetaRepoMarkerID(proj))).To(BeFalse())
		})

		It("gates metadata detection to exact record formats", func() {
			Expect(isMetadataCandidateTag("cleanup")).To(BeTrue())
			Expect(isMetadataCandidateTag("cleanup-foo")).To(BeFalse())
			Expect(isMetadataCandidateTag("meta-abc_commit_stage")).To(BeTrue())
			Expect(isMetadataCandidateTag("meta-foo")).To(BeFalse())
		})

		It("with a -rejected-suffixed project name is not returned by GetRejectedStageIDs", func(ctx SpecContext) {
			rejectedProj := "myproj-rejected"
			Expect(stages.PutMetaRepoMarker(ctx, rejectedProj, metaRepo)).To(Succeed())
			ids, err := stages.GetRejectedStageIDs(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(ids).To(BeEmpty())
		})
	})

	Describe("HasProjectMetadata / collectProjectMetadataRecords", func() {
		It("detects each owned family and ignores foreign / excluded records", func(ctx SpecContext) {
			reg.put(stagesRepo+":managed-image-app", map[string]string{image.WerfLabel: proj})
			reg.put(stagesRepo+":meta-abc_commit_stage", map[string]string{image.WerfLabel: proj})
			reg.put(stagesRepo+":custom-tag-meta-v1", map[string]string{image.WerfLabel: proj})
			reg.put(stagesRepo+":cleanup", map[string]string{image.WerfLabel: proj, RepoCleanUpRecord_LabelTimestamp: "123"})
			reg.put(stagesRepo+":managed-image-other", map[string]string{image.WerfLabel: "otherproject"})
			reg.put(stagesRepo+":abc-123-rejected", map[string]string{image.WerfLabel: proj})
			reg.put(markerRef(proj), map[string]string{image.WerfLabel: proj, RepoMetaRepoMarker_LabelMetaRepo: metaRepo})

			has, err := stages.HasProjectMetadata(ctx, proj)
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeTrue())

			records, err := stages.collectProjectMetadataRecords(ctx, proj, false)
			Expect(err).NotTo(HaveOccurred())
			var tags []string
			for _, rec := range records {
				tags = append(tags, rec.tag)
			}
			Expect(tags).To(ConsistOf("managed-image-app", "meta-abc_commit_stage", "custom-tag-meta-v1", "cleanup"))
		})

		It("ignores a cleanup record without the timestamp label", func(ctx SpecContext) {
			reg.put(stagesRepo+":cleanup", map[string]string{image.WerfLabel: proj})
			has, err := stages.HasProjectMetadata(ctx, proj)
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeFalse())
		})

		It("returns false when only another project's metadata is present", func(ctx SpecContext) {
			reg.put(stagesRepo+":managed-image-app", map[string]string{image.WerfLabel: "otherproject"})
			has, err := stages.HasProjectMetadata(ctx, proj)
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeFalse())
		})

		It("conservatively treats unlabeled image-metadata as the project's (legacy records)", func(ctx SpecContext) {
			reg.put(stagesRepo+":meta-abc_commit_stage", map[string]string{})
			has, err := stages.HasProjectMetadata(ctx, proj)
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeTrue())
		})

		It("conservatively treats unlabeled managed-image records as the project's (legacy records)", func(ctx SpecContext) {
			reg.put(stagesRepo+":managed-image-app", map[string]string{})
			has, err := stages.HasProjectMetadata(ctx, proj)
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeTrue())
		})

		It("excludes image-metadata explicitly labeled for another project", func(ctx SpecContext) {
			reg.put(stagesRepo+":meta-abc_commit_stage", map[string]string{image.WerfLabel: "otherproject"})
			has, err := stages.HasProjectMetadata(ctx, proj)
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeFalse())
		})
	})

	Describe("decorator", func() {
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

		It("plants the marker on a managed-image write", func(ctx SpecContext) {
			d := newDecorator(false)
			Expect(d.AddManagedImage(ctx, proj, "app")).To(Succeed())
			Expect(reg.has(markerRef(proj))).To(BeTrue())
		})

		It("plants the marker on a deletion-only first use", func(ctx SpecContext) {
			d := newDecorator(false)
			Expect(d.RmManagedImage(ctx, proj, "app")).To(Succeed())
			Expect(reg.has(markerRef(proj))).To(BeTrue())
		})

		It("plants on an enabled cleanup-record-only write", func(ctx SpecContext) {
			d := newDecorator(false)
			Expect(d.PostLastCleanupRecord(ctx, proj)).To(Succeed())
			Expect(reg.has(markerRef(proj))).To(BeTrue())
		})

		It("does not plant when cleanup is disabled", func(ctx SpecContext) {
			d := newDecorator(true)
			Expect(d.PostLastCleanupRecord(ctx, proj)).To(Succeed())
			Expect(reg.has(markerRef(proj))).To(BeFalse())
		})

		It("writes neither record nor marker when cleanup is disabled", func(ctx SpecContext) {
			d := newDecorator(true)
			meta.cleanupDisabled = true
			Expect(d.PostLastCleanupRecord(ctx, proj)).To(Succeed())
			Expect(reg.has(fmt.Sprintf(RepoCleanUpRecord_ImageNameFormat, metaRepo))).To(BeFalse())
			Expect(reg.has(markerRef(proj))).To(BeFalse())
		})

		It("writes both record and marker when cleanup is enabled", func(ctx SpecContext) {
			d := newDecorator(false)
			Expect(d.PostLastCleanupRecord(ctx, proj)).To(Succeed())
			Expect(reg.has(fmt.Sprintf(RepoCleanUpRecord_ImageNameFormat, metaRepo))).To(BeTrue())
			Expect(reg.has(markerRef(proj))).To(BeTrue())
		})

		It("does not plant on read-only access", func(ctx SpecContext) {
			d := newDecorator(false)
			_, err := d.GetManagedImages(ctx, proj)
			Expect(err).NotTo(HaveOccurred())
			Expect(reg.has(markerRef(proj))).To(BeFalse())
		})

		It("fails when a conflicting marker already exists", func(ctx SpecContext) {
			Expect(stages.PutMetaRepoMarker(ctx, proj, "registry.example/other")).To(Succeed())
			d := newDecorator(false)
			err := d.AddManagedImage(ctx, proj, "app")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("must match"))
		})

		It("plants exactly once across repeated and concurrent writes", func(ctx SpecContext) {
			d := newDecorator(false)
			var wg sync.WaitGroup
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()
					Expect(d.AddManagedImage(ctx, proj, "app")).To(Succeed())
				}()
			}
			wg.Wait()
			Expect(reg.pushCount(markerRef(proj))).To(Equal(1))
		})
	})

	Describe("SetupMetaRepoSafeguard", func() {
		var meta *RepoStagesStorage
		BeforeEach(func() { meta = newRepoStorage(reg, metaRepo) })

		It("skips non-registry stages storage", func(ctx SpecContext) {
			local := NewLocalStagesStorage(nil)
			got, err := SetupMetaRepoSafeguard(ctx, proj, local, meta, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(BeIdenticalTo(meta))
		})

		It("legacy: no flag, no marker → passthrough", func(ctx SpecContext) {
			got, err := SetupMetaRepoSafeguard(ctx, proj, stages, stages, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(BeIdenticalTo(stages))
		})

		It("no flag but marker present → hard error", func(ctx SpecContext) {
			Expect(stages.PutMetaRepoMarker(ctx, proj, metaRepo)).To(Succeed())
			_, err := SetupMetaRepoSafeguard(ctx, proj, stages, stages, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("pass --meta-repo"))
		})

		It("rejects a malformed marker (present, no address)", func(ctx SpecContext) {
			reg.put(markerRef(proj), map[string]string{image.WerfLabel: proj})
			_, err := SetupMetaRepoSafeguard(ctx, proj, stages, meta, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("malformed"))
		})

		It("configured + matching marker → decorated", func(ctx SpecContext) {
			Expect(stages.PutMetaRepoMarker(ctx, proj, metaRepo)).To(Succeed())
			got, err := SetupMetaRepoSafeguard(ctx, proj, stages, meta, false)
			Expect(err).NotTo(HaveOccurred())
			_, ok := got.(*metaRepoMarkerStorage)
			Expect(ok).To(BeTrue())
		})

		It("configured + mismatching marker → hard error", func(ctx SpecContext) {
			Expect(stages.PutMetaRepoMarker(ctx, proj, "registry.example/other")).To(Succeed())
			_, err := SetupMetaRepoSafeguard(ctx, proj, stages, meta, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("they must match"))
		})

		It("configured + no marker + existing metadata → migrate error", func(ctx SpecContext) {
			reg.put(stagesRepo+":managed-image-app", map[string]string{image.WerfLabel: proj})
			_, err := SetupMetaRepoSafeguard(ctx, proj, stages, meta, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("werf meta-repo migrate"))
		})

		It("configured + no marker + clean repo → decorated", func(ctx SpecContext) {
			got, err := SetupMetaRepoSafeguard(ctx, proj, stages, meta, false)
			Expect(err).NotTo(HaveOccurred())
			_, ok := got.(*metaRepoMarkerStorage)
			Expect(ok).To(BeTrue())
		})

		It("meta-repo equal to repo + marker to a distinct repo → hard error", func(ctx SpecContext) {
			Expect(stages.PutMetaRepoMarker(ctx, proj, metaRepo)).To(Succeed())
			sameMeta := newRepoStorage(reg, stagesRepo)
			_, err := SetupMetaRepoSafeguard(ctx, proj, stages, sameMeta, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("resolves to the same repository as --repo"))
			Expect(err.Error()).To(ContainSubstring(metaRepo))
			Expect(err.Error()).To(ContainSubstring("werf meta-repo detach"))
		})

		It("meta-repo equal to repo + no marker → passthrough", func(ctx SpecContext) {
			sameMeta := newRepoStorage(reg, stagesRepo)
			got, err := SetupMetaRepoSafeguard(ctx, proj, stages, sameMeta, false)
			Expect(err).NotTo(HaveOccurred())
			_, ok := got.(*metaRepoMarkerStorage)
			Expect(ok).To(BeFalse())
		})
	})

	Describe("MigrateMetaRepo", func() {
		var meta *RepoStagesStorage
		BeforeEach(func() {
			meta = newRepoStorage(reg, metaRepo)
			reg.put(stagesRepo+":managed-image-app", map[string]string{image.WerfLabel: proj})
			reg.put(stagesRepo+":meta-abc_commit_stage", map[string]string{image.WerfLabel: proj})
			reg.put(stagesRepo+":managed-image-other", map[string]string{image.WerfLabel: "otherproject"})
			reg.put(stagesRepo+":abc-123-rejected", map[string]string{image.WerfLabel: proj})
		})

		It("copies owned records, plants marker, keeps source", func(ctx SpecContext) {
			Expect(MigrateMetaRepo(ctx, proj, stages, meta, MigrateMetaRepoOptions{})).To(Succeed())
			Expect(reg.has(metaRepo + ":managed-image-app")).To(BeTrue())
			Expect(reg.has(metaRepo + ":meta-abc_commit_stage")).To(BeTrue())
			Expect(reg.has(metaRepo + ":managed-image-other")).To(BeFalse())
			Expect(reg.has(metaRepo + ":abc-123-rejected")).To(BeFalse())
			Expect(reg.has(stagesRepo + ":managed-image-app")).To(BeTrue())
			addr, found, _ := stages.GetMetaRepoMarker(ctx, proj)
			Expect(found).To(BeTrue())
			Expect(addr).To(Equal(metaRepo))
		})

		It("succeeds and plants the marker for a project with no metadata", func(ctx SpecContext) {
			emptyReg := newMarkerRegistry()
			src := newRepoStorage(emptyReg, stagesRepo)
			dst := newRepoStorage(emptyReg, metaRepo)
			Expect(MigrateMetaRepo(ctx, proj, src, dst, MigrateMetaRepoOptions{})).To(Succeed())
			addr, found, _ := src.GetMetaRepoMarker(ctx, proj)
			Expect(found).To(BeTrue())
			Expect(addr).To(Equal(metaRepo))
		})

		It("is idempotent and skips already-present destination records", func(ctx SpecContext) {
			reg.put(metaRepo+":managed-image-app", map[string]string{image.WerfLabel: proj})
			Expect(MigrateMetaRepo(ctx, proj, stages, meta, MigrateMetaRepoOptions{})).To(Succeed())
			Expect(reg.has(metaRepo + ":meta-abc_commit_stage")).To(BeTrue())
		})

		It("rejects a source marker pointing at a different meta-repo", func(ctx SpecContext) {
			Expect(stages.PutMetaRepoMarker(ctx, proj, "registry.example/other")).To(Succeed())
			err := MigrateMetaRepo(ctx, proj, stages, meta, MigrateMetaRepoOptions{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("refusing to migrate"))
		})

		It("refuses when --meta-repo resolves to the same repo as --repo and deletes nothing", func(ctx SpecContext) {
			sameMeta := newRepoStorage(reg, stagesRepo)
			err := MigrateMetaRepo(ctx, proj, stages, sameMeta, MigrateMetaRepoOptions{RemoveSource: true})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("same repository"))
			Expect(reg.has(stagesRepo + ":managed-image-app")).To(BeTrue())
			Expect(reg.has(stagesRepo + ":meta-abc_commit_stage")).To(BeTrue())
		})

		It("refuses when a destination record is owned by another project", func(ctx SpecContext) {
			reg.put(metaRepo+":managed-image-app", map[string]string{image.WerfLabel: "otherproject"})
			err := MigrateMetaRepo(ctx, proj, stages, meta, MigrateMetaRepoOptions{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not a valid metadata record"))
		})

		It("refuses to delete source when the destination cleanup record lacks the timestamp", func(ctx SpecContext) {
			reg.put(stagesRepo+":cleanup", map[string]string{image.WerfLabel: proj, RepoCleanUpRecord_LabelTimestamp: "123"})
			reg.put(metaRepo+":cleanup", map[string]string{image.WerfLabel: proj})
			err := MigrateMetaRepo(ctx, proj, stages, meta, MigrateMetaRepoOptions{RemoveSource: true})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not a valid metadata record"))
			Expect(reg.has(stagesRepo + ":cleanup")).To(BeTrue())
		})

		It("leaves source intact when a copy fails", func(ctx SpecContext) {
			reg.copyErr = fmt.Errorf("boom")
			err := MigrateMetaRepo(ctx, proj, stages, meta, MigrateMetaRepoOptions{RemoveSource: true})
			Expect(err).To(HaveOccurred())
			Expect(reg.has(stagesRepo + ":managed-image-app")).To(BeTrue())
		})

		It("removes source only after verifying the destination copy", func(ctx SpecContext) {
			Expect(MigrateMetaRepo(ctx, proj, stages, meta, MigrateMetaRepoOptions{RemoveSource: true})).To(Succeed())
			Expect(reg.has(stagesRepo + ":managed-image-app")).To(BeFalse())
			Expect(reg.has(stagesRepo + ":meta-abc_commit_stage")).To(BeFalse())
			Expect(reg.has(metaRepo + ":managed-image-app")).To(BeTrue())
			Expect(reg.has(stagesRepo + ":managed-image-other")).To(BeTrue())
		})
	})
})
