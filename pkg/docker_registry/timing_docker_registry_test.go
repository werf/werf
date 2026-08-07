package docker_registry

import (
	"context"
	"errors"
	"io"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	registry_api "github.com/werf/werf/v2/pkg/docker_registry/api"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/opstats"
)

type fakeRegistry struct {
	Interface
}

func (r *fakeRegistry) CreateRepo(_ context.Context, _ string) error { return nil }
func (r *fakeRegistry) DeleteRepo(_ context.Context, _ string) error { return nil }

func (r *fakeRegistry) Tags(_ context.Context, _ string, _ ...Option) ([]string, error) {
	return []string{"latest"}, nil
}

func (r *fakeRegistry) IsTagExist(_ context.Context, _ string, _ ...Option) (bool, error) {
	return true, nil
}

func (r *fakeRegistry) TagRepoImage(_ context.Context, _ *image.Info, _ string) error { return nil }

func (r *fakeRegistry) GetRepoImage(_ context.Context, _ string) (*image.Info, error) {
	return &image.Info{}, nil
}

func (r *fakeRegistry) TryGetRepoImage(_ context.Context, _ string) (*image.Info, error) {
	return &image.Info{}, nil
}

func (r *fakeRegistry) DeleteRepoImage(_ context.Context, _ *image.Info) error { return nil }

func (r *fakeRegistry) PushImage(_ context.Context, _ string, _ *PushImageOptions) error {
	return nil
}

func (r *fakeRegistry) MutateAndPushImage(_ context.Context, _, _ string, _ ...registry_api.MutateOption) error {
	return nil
}

func (r *fakeRegistry) CopyImage(_ context.Context, _, _ string, _ CopyImageOptions) error {
	return nil
}

func (r *fakeRegistry) PushImageArchive(_ context.Context, _ ArchiveOpener, _ string) error {
	return nil
}

func (r *fakeRegistry) PullImageArchive(_ context.Context, _ io.Writer, _ string) error {
	return nil
}

func (r *fakeRegistry) PushManifestList(_ context.Context, _ string, _ ManifestListOptions) error {
	return nil
}

type failingRegistry struct {
	Interface
}

func (r *failingRegistry) Tags(_ context.Context, _ string, _ ...Option) ([]string, error) {
	return nil, errors.New("boom")
}

type fakeGenericApi struct {
	GenericApiInterface
}

func (r *fakeGenericApi) GetRepoImage(_ context.Context, _ string) (*image.Info, error) {
	return &image.Info{}, nil
}

func (r *fakeGenericApi) MutateAndPushImage(_ context.Context, _, _ string, _ ...registry_api.MutateOption) error {
	return nil
}

func (r *fakeGenericApi) GetRepoImageConfigFile(_ context.Context, _ string) (*v1.ConfigFile, error) {
	return &v1.ConfigFile{}, nil
}

func expectSingleMeasurement(collector *opstats.Collector, method string) {
	GinkgoHelper()
	summary := collector.Summary()
	Expect(summary).To(HaveLen(1))
	Expect(summary[0].Operation).To(Equal(opstats.Operation("registry: " + method)))
	Expect(summary[0].Count).To(Equal(1))
}

var _ = Describe("timingDockerRegistry", func() {
	DescribeTable("records one measurement per method",
		func(method string, call func(ctx context.Context, r Interface) error) {
			collector := opstats.NewCollector()
			ctx := opstats.NewContext(context.Background(), collector)

			Expect(call(ctx, newTimingDockerRegistry(&fakeRegistry{}))).To(Succeed())
			expectSingleMeasurement(collector, method)
		},
		Entry("CreateRepo", "CreateRepo", func(ctx context.Context, r Interface) error {
			return r.CreateRepo(ctx, "repo")
		}),
		Entry("DeleteRepo", "DeleteRepo", func(ctx context.Context, r Interface) error {
			return r.DeleteRepo(ctx, "repo")
		}),
		Entry("Tags", "Tags", func(ctx context.Context, r Interface) error {
			_, err := r.Tags(ctx, "repo")
			return err
		}),
		Entry("IsTagExist", "IsTagExist", func(ctx context.Context, r Interface) error {
			_, err := r.IsTagExist(ctx, "repo:tag")
			return err
		}),
		Entry("TagRepoImage", "TagRepoImage", func(ctx context.Context, r Interface) error {
			return r.TagRepoImage(ctx, &image.Info{}, "tag")
		}),
		Entry("GetRepoImage", "GetRepoImage", func(ctx context.Context, r Interface) error {
			_, err := r.GetRepoImage(ctx, "repo:tag")
			return err
		}),
		Entry("TryGetRepoImage", "TryGetRepoImage", func(ctx context.Context, r Interface) error {
			_, err := r.TryGetRepoImage(ctx, "repo:tag")
			return err
		}),
		Entry("DeleteRepoImage", "DeleteRepoImage", func(ctx context.Context, r Interface) error {
			return r.DeleteRepoImage(ctx, &image.Info{})
		}),
		Entry("PushImage", "PushImage", func(ctx context.Context, r Interface) error {
			return r.PushImage(ctx, "repo:tag", &PushImageOptions{})
		}),
		Entry("MutateAndPushImage", "MutateAndPushImage", func(ctx context.Context, r Interface) error {
			return r.MutateAndPushImage(ctx, "src", "dst")
		}),
		Entry("CopyImage", "CopyImage", func(ctx context.Context, r Interface) error {
			return r.CopyImage(ctx, "src", "dst", CopyImageOptions{})
		}),
		Entry("PushImageArchive", "PushImageArchive", func(ctx context.Context, r Interface) error {
			return r.PushImageArchive(ctx, nil, "repo:tag")
		}),
		Entry("PullImageArchive", "PullImageArchive", func(ctx context.Context, r Interface) error {
			return r.PullImageArchive(ctx, io.Discard, "repo:tag")
		}),
		Entry("PushManifestList", "PushManifestList", func(ctx context.Context, r Interface) error {
			return r.PushManifestList(ctx, "repo:tag", ManifestListOptions{})
		}),
	)

	It("records the measurement when the wrapped call fails", func() {
		collector := opstats.NewCollector()
		ctx := opstats.NewContext(context.Background(), collector)

		_, err := newTimingDockerRegistry(&failingRegistry{}).Tags(ctx, "repo")
		Expect(err).To(MatchError("boom"))

		expectSingleMeasurement(collector, "Tags")
	})
})

var _ = Describe("timingGenericApi", func() {
	DescribeTable("records one measurement per method",
		func(method string, call func(ctx context.Context, api GenericApiInterface) error) {
			collector := opstats.NewCollector()
			ctx := opstats.NewContext(context.Background(), collector)

			Expect(call(ctx, newTimingGenericApi(&fakeGenericApi{}))).To(Succeed())
			expectSingleMeasurement(collector, method)
		},
		Entry("GetRepoImage", "GetRepoImage", func(ctx context.Context, api GenericApiInterface) error {
			_, err := api.GetRepoImage(ctx, "repo:tag")
			return err
		}),
		Entry("MutateAndPushImage", "MutateAndPushImage", func(ctx context.Context, api GenericApiInterface) error {
			return api.MutateAndPushImage(ctx, "src", "dst")
		}),
		Entry("GetRepoImageConfigFile", "GetRepoImageConfigFile", func(ctx context.Context, api GenericApiInterface) error {
			_, err := api.GetRepoImageConfigFile(ctx, "repo:tag")
			return err
		}),
	)
})
