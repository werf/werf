package docker_registry

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("registry tags page size", func() {
	ginkgo.BeforeEach(func() {
		configDir := ginkgo.GinkgoT().TempDir()
		gomega.Expect(os.WriteFile(filepath.Join(configDir, "config.json"), []byte(`{"auths":{}}`), 0o600)).To(gomega.Succeed())
		ginkgo.GinkgoT().Setenv("DOCKER_CONFIG", configDir)
	})

	ginkgo.DescribeTable("requests one page sized for the registry host", func(host, expectedPageSize string) {
		fixture := newTagsPageSizeFixture("latest")
		ginkgo.DeferCleanup(fixture.server.Close)

		tags, err := newTagsPageSizeAPI(fixture).Tags(context.Background(), host+"/repo")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(tags).To(gomega.Equal([]string{"latest"}))
		gomega.Expect(fixture.requestedPageSizes()).To(gomega.Equal([]string{expectedPageSize}))
	},
		ginkgo.Entry("ordinary registry", "registry.example.test", "1000000"),
		ginkgo.Entry("mixed-case ordinary registry", "Registry.Example.Test", "1000000"),
		ginkgo.Entry("public ECR", "public.ecr.aws", "1000"),
		ginkgo.Entry("uppercase public ECR", "PUBLIC.ECR.AWS", "1000"),
		ginkgo.Entry("public ECR with an explicit port", "public.ecr.aws:443", "1000"),
		ginkgo.Entry("host merely prefixed with the public ECR domain", "public.ecr.aws.example.test", "1000000"),
		ginkgo.Entry("host merely containing the ECR domain", "123456789012.dkr.ecr.eu-central-1.amazonaws.com.example.test", "1000000"),
	)

	ginkgo.DescribeTable("requests one page sized for the registry implementation", func(implementation, host, expectedPageSize string) {
		fixture := newTagsPageSizeFixture("latest")
		ginkgo.DeferCleanup(fixture.server.Close)

		tags, err := newTagsPageSizeAPIForImplementation(fixture, implementation).Tags(context.Background(), host+"/repo")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(tags).To(gomega.Equal([]string{"latest"}))
		gomega.Expect(fixture.requestedPageSizes()).To(gomega.Equal([]string{expectedPageSize}))
	},
		ginkgo.Entry("ECR", AwsEcrImplementationName, "123456789012.dkr.ecr.eu-central-1.amazonaws.com", "1000"),
		ginkgo.Entry("ECR whatever the host", AwsEcrImplementationName, "registry.example.test", "1000"),
		ginkgo.Entry("GitLab registry", GitLabRegistryImplementationName, "registry.example.test", "1000000"),
		ginkgo.Entry("default", DefaultImplementationName, "registry.example.test", "1000000"),
	)

	ginkgo.DescribeTable("takes the page size from the environment", func(configured, expectedPageSize string) {
		ginkgo.GinkgoT().Setenv(tagsPageSizeEnv, configured)
		fixture := newTagsPageSizeFixture("latest")
		ginkgo.DeferCleanup(fixture.server.Close)

		tags, err := newTagsPageSizeAPI(fixture).Tags(context.Background(), "registry.example.test/repo")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(tags).To(gomega.Equal([]string{"latest"}))
		gomega.Expect(fixture.requestedPageSizes()).To(gomega.Equal([]string{expectedPageSize}))
		gomega.Expect(fixture.httpRequests.Load()).To(gomega.BeNumerically(">", 0))
	},
		ginkgo.Entry("an empty value", "", "1000000"),
		ginkgo.Entry("a smaller page size", "200000", "200000"),
		ginkgo.Entry("a larger page size", "2000000", "2000000"),
		ginkgo.Entry("the client library default", "0", "1000"),
		ginkgo.Entry("a padded value", "  200000  ", "200000"),
	)

	ginkgo.DescribeTable("rejects an invalid page size in the environment before making any HTTP request", func(configured string) {
		ginkgo.GinkgoT().Setenv(tagsPageSizeEnv, configured)
		fixture := newTagsPageSizeFixture("latest")
		ginkgo.DeferCleanup(fixture.server.Close)

		_, err := newTagsPageSizeAPI(fixture).Tags(context.Background(), "registry.example.test/repo")
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring(tagsPageSizeEnv)))
		gomega.Expect(fixture.requestedPageSizes()).To(gomega.BeEmpty())
		gomega.Expect(fixture.httpRequests.Load()).To(gomega.BeZero())
	},
		ginkgo.Entry("a negative page size", "-1"),
		ginkgo.Entry("a non-integer page size", "many"),
		ginkgo.Entry("a fractional page size", "1000.5"),
		ginkgo.Entry("a page size with a suffix", "1000000 tags"),
	)

	ginkgo.It("keeps the ECR page size over a configured page size", func() {
		ginkgo.GinkgoT().Setenv(tagsPageSizeEnv, "200000")
		fixture := newTagsPageSizeFixture("latest")
		ginkgo.DeferCleanup(fixture.server.Close)

		_, err := newTagsPageSizeAPIForImplementation(fixture, AwsEcrImplementationName).Tags(context.Background(), "registry.example.test/repo")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(fixture.requestedPageSizes()).To(gomega.Equal([]string{"1000"}))
	})

	ginkgo.It("keeps the fallback page size of a rejecting host over a configured page size", func() {
		fixture := newTagsPageSizeFixture("latest")
		fixture.rejectAbove = 1000
		fixture.rejectStatus, fixture.rejectCode, fixture.rejectMessage = http.StatusBadRequest, "PAGINATION_NUMBER_INVALID", "invalid number of tags requested"
		ginkgo.DeferCleanup(fixture.server.Close)

		host := nextTagsPageSizeHost()
		_, err := newTagsPageSizeAPI(fixture).Tags(context.Background(), host+"/repo")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(fixture.requestedPageSizes()).To(gomega.Equal([]string{"1000000", "1000"}))

		ginkgo.GinkgoT().Setenv(tagsPageSizeEnv, "200000")
		_, err = newTagsPageSizeAPI(fixture).Tags(context.Background(), host+"/other-repo")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(fixture.requestedPageSizes()).To(gomega.Equal([]string{"1000000", "1000", "1000"}))
	})

	ginkgo.It("retries a private ECR host driven by another implementation with the fallback page size", func() {
		fixture := newTagsPageSizeFixture("latest")
		fixture.rejectAbove = 1000
		fixture.rejectStatus, fixture.rejectCode, fixture.rejectMessage = http.StatusBadRequest, "UNSUPPORTED", "Invalid parameter at 'maxResults' failed to satisfy constraint: 'Member must have value less than or equal to 1000'"
		ginkgo.DeferCleanup(fixture.server.Close)

		tags, err := newTagsPageSizeAPI(fixture).Tags(context.Background(), "123456789012.dkr.ecr.eu-central-1.amazonaws.com/repo")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(tags).To(gomega.Equal([]string{"latest"}))
		gomega.Expect(fixture.requestedPageSizes()).To(gomega.Equal([]string{"1000000", "1000"}))
	})

	ginkgo.It("follows the server pagination link without replacing its page size", func() {
		fixture := newTagsPageSizeFixture("a", "b", "c", "d", "e")
		fixture.chunk = 2
		ginkgo.DeferCleanup(fixture.server.Close)

		tags, err := newTagsPageSizeAPI(fixture).Tags(context.Background(), "paginated.example.test/repo")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(tags).To(gomega.Equal([]string{"a", "b", "c", "d", "e"}))
		gomega.Expect(fixture.requestedPageSizes()).To(gomega.Equal([]string{"1000000", "2", "2"}))
	})

	ginkgo.It("restarts the listing from the first page when a later page is rejected", func() {
		fixture := newTagsPageSizeFixture("a", "b", "c", "d", "e")
		fixture.chunk = 2
		fixture.rejectAfterRequests, fixture.rejectLimit = 1, 1
		fixture.rejectStatus, fixture.rejectCode, fixture.rejectMessage = http.StatusBadRequest, "PAGINATION_NUMBER_INVALID", "invalid number of tags requested"
		ginkgo.DeferCleanup(fixture.server.Close)

		tags, err := newTagsPageSizeAPI(fixture).Tags(context.Background(), nextTagsPageSizeHost()+"/repo")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(tags).To(gomega.Equal([]string{"a", "b", "c", "d", "e"}))
		gomega.Expect(fixture.requestedPageSizes()).To(gomega.Equal([]string{"1000000", "2", "1000", "2", "2"}))
	})

	ginkgo.DescribeTable("retries with the fallback page size only on an explicit page size rejection", func(status int, code, message string, expectFallback bool) {
		fixture := newTagsPageSizeFixture("latest")
		fixture.rejectAbove = 1000
		fixture.rejectStatus, fixture.rejectCode, fixture.rejectMessage = status, code, message
		if status == http.StatusTooManyRequests || status >= http.StatusInternalServerError {
			fixture.rejectLimit = 1
		}
		ginkgo.DeferCleanup(fixture.server.Close)

		tags, err := newTagsPageSizeAPI(fixture).Tags(context.Background(), nextTagsPageSizeHost()+"/repo")
		if expectFallback {
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(tags).To(gomega.Equal([]string{"latest"}))
			gomega.Expect(fixture.requestedPageSizes()).To(gomega.Equal([]string{"1000000", "1000"}))
			return
		}

		if fixture.rejectLimit == 0 && status != http.StatusNotFound {
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(tags).To(gomega.BeEmpty())
		}
		gomega.Expect(fixture.requestedPageSizes()).NotTo(gomega.BeEmpty())
		gomega.Expect(fixture.requestedPageSizes()).To(gomega.HaveEach("1000000"))
	},
		ginkgo.Entry("OCI pagination code", http.StatusBadRequest, "PAGINATION_NUMBER_INVALID", "invalid number of tags requested", true),
		ginkgo.Entry("OCI pagination code with an unprocessable status", http.StatusUnprocessableEntity, "PAGINATION_NUMBER_INVALID", "invalid number of tags requested", true),
		ginkgo.Entry("OCI pagination code with a request entity too large status", http.StatusRequestEntityTooLarge, "PAGINATION_NUMBER_INVALID", "invalid number of tags requested", true),
		ginkgo.Entry("ECR maxResults rejection", http.StatusBadRequest, "UNSUPPORTED", "Invalid parameter at 'maxResults' failed to satisfy constraint: 'Member must have value less than or equal to 1000'", true),
		ginkgo.Entry("ECR maxResults rejection in another casing and bound", http.StatusBadRequest, "UNSUPPORTED", "invalid parameter at 'maxresults' failed to satisfy constraint: 'member must have value less than or equal to 100'", true),
		ginkgo.Entry("unsupported operation", http.StatusBadRequest, "UNSUPPORTED", "The operation is unsupported", false),
		ginkgo.Entry("invalid name", http.StatusBadRequest, "NAME_INVALID", "invalid repository name", false),
		ginkgo.Entry("unauthorized", http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", false),
		ginkgo.Entry("denied", http.StatusForbidden, "DENIED", "requested access to the resource is denied", false),
		ginkgo.Entry("unknown repository", http.StatusNotFound, "NAME_UNKNOWN", "repository name not known to registry", false),
		ginkgo.Entry("too many requests", http.StatusTooManyRequests, "TOOMANYREQUESTS", "too many requests", false),
		ginkgo.Entry("server error", http.StatusInternalServerError, "UNKNOWN", "internal error", false),
		ginkgo.Entry("pagination code with an unauthorized status", http.StatusUnauthorized, "PAGINATION_NUMBER_INVALID", "invalid number of tags requested", false),
		ginkgo.Entry("pagination code with a forbidden status", http.StatusForbidden, "PAGINATION_NUMBER_INVALID", "invalid number of tags requested", false),
		ginkgo.Entry("pagination code with a too many requests status", http.StatusTooManyRequests, "PAGINATION_NUMBER_INVALID", "invalid number of tags requested", false),
		ginkgo.Entry("pagination code with a service unavailable status", http.StatusServiceUnavailable, "PAGINATION_NUMBER_INVALID", "invalid number of tags requested", false),
	)

	ginkgo.DescribeTable("treats a pagination rejection as such only for a tag listing request", func(method, path string, expected bool) {
		request, err := http.NewRequest(method, "https://registry.example.test"+path, nil)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		err = &transport.Error{
			StatusCode: http.StatusBadRequest,
			Request:    request,
			Errors: []transport.Diagnostic{
				{Code: paginationNumberInvalidErrorCode, Message: "invalid number of tags requested"},
			},
		}
		gomega.Expect(isTagsPageSizeRejectedErr(err)).To(gomega.Equal(expected))
	},
		ginkgo.Entry("tag listing", http.MethodGet, "/v2/repo/tags/list", true),
		ginkgo.Entry("another method", http.MethodPost, "/v2/repo/tags/list", false),
		ginkgo.Entry("another path", http.MethodGet, "/v2/repo/manifests/latest", false),
		ginkgo.Entry("the registry api endpoint", http.MethodGet, "/v2/", false),
		ginkgo.Entry("the token endpoint", http.MethodGet, "/token", false),
	)

	ginkgo.It("keeps the fallback page size for the rejecting host only", func() {
		rejecting := newTagsPageSizeFixture("latest")
		rejecting.rejectAbove = 1000
		rejecting.rejectStatus, rejecting.rejectCode, rejecting.rejectMessage = http.StatusBadRequest, "PAGINATION_NUMBER_INVALID", "invalid number of tags requested"
		ginkgo.DeferCleanup(rejecting.server.Close)

		host := nextTagsPageSizeHost()
		_, err := newTagsPageSizeAPI(rejecting).Tags(context.Background(), host+"/repo")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(rejecting.requestedPageSizes()).To(gomega.Equal([]string{"1000000", "1000"}))

		_, err = newTagsPageSizeAPI(rejecting).Tags(context.Background(), host+"/other-repo")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(rejecting.requestedPageSizes()).To(gomega.Equal([]string{"1000000", "1000", "1000"}))

		other := newTagsPageSizeFixture("latest")
		ginkgo.DeferCleanup(other.server.Close)
		_, err = newTagsPageSizeAPI(other).Tags(context.Background(), "unrelated.example.test/repo")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(other.requestedPageSizes()).To(gomega.Equal([]string{"1000000"}))
	})

	ginkgo.It("keeps the fallback page size whatever the casing of the registry host", func() {
		fixture := newTagsPageSizeFixture("latest")
		fixture.rejectAbove = 1000
		fixture.rejectStatus, fixture.rejectCode, fixture.rejectMessage = http.StatusBadRequest, "PAGINATION_NUMBER_INVALID", "invalid number of tags requested"
		ginkgo.DeferCleanup(fixture.server.Close)

		host := nextTagsPageSizeHost()
		_, err := newTagsPageSizeAPI(fixture).Tags(context.Background(), host+"/repo")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(fixture.requestedPageSizes()).To(gomega.Equal([]string{"1000000", "1000"}))

		_, err = newTagsPageSizeAPI(fixture).Tags(context.Background(), strings.ToUpper(host)+"/repo")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(fixture.requestedPageSizes()).To(gomega.Equal([]string{"1000000", "1000", "1000"}))
	})

	ginkgo.It("keeps the fallback page size for the rejecting port only", func() {
		fixture := newTagsPageSizeFixture("latest")
		fixture.rejectAbove = 1000
		fixture.rejectStatus, fixture.rejectCode, fixture.rejectMessage = http.StatusBadRequest, "PAGINATION_NUMBER_INVALID", "invalid number of tags requested"
		ginkgo.DeferCleanup(fixture.server.Close)

		host := nextTagsPageSizeHost()
		_, err := newTagsPageSizeAPI(fixture).Tags(context.Background(), host+":5000/repo")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(fixture.requestedPageSizes()).To(gomega.Equal([]string{"1000000", "1000"}))

		_, err = newTagsPageSizeAPI(fixture).Tags(context.Background(), host+":5001/repo")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(fixture.requestedPageSizes()).To(gomega.Equal([]string{"1000000", "1000", "1000000", "1000"}))
	})

	ginkgo.It("fails with the fallback rejection when the fallback page size is rejected too", func() {
		fixture := newTagsPageSizeFixture("latest")
		fixture.rejectStatus, fixture.rejectCode, fixture.rejectMessage = http.StatusBadRequest, "PAGINATION_NUMBER_INVALID", "invalid number of tags requested"
		ginkgo.DeferCleanup(fixture.server.Close)

		host := nextTagsPageSizeHost()
		tags, err := newTagsPageSizeAPI(fixture).Tags(context.Background(), host+"/repo")
		gomega.Expect(tags).To(gomega.BeNil())
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("PAGINATION_NUMBER_INVALID")))
		gomega.Expect(fixture.requestedPageSizes()).To(gomega.Equal([]string{"1000000", "1000"}))

		fixture.stopRejecting()
		_, err = newTagsPageSizeAPI(fixture).Tags(context.Background(), host+"/repo")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(fixture.requestedPageSizes()[2:]).To(gomega.Equal([]string{"1000000"}))
	})

	ginkgo.DescribeTable("never caps the registry host on a failure that is not a page size rejection", func(status int, code, message string) {
		fixture := newTagsPageSizeFixture("latest")
		fixture.rejectStatus, fixture.rejectCode, fixture.rejectMessage = status, code, message
		ginkgo.DeferCleanup(fixture.server.Close)
		ctx, cancel := context.WithCancel(context.Background())
		ginkgo.DeferCleanup(cancel)
		// werf retries these statuses for minutes, so end the listing once the registry has answered.
		fixture.onReject = cancel

		host := nextTagsPageSizeHost()
		_, err := newTagsPageSizeAPI(fixture).Tags(ctx, host+"/repo")
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(fixture.requestedPageSizes()).NotTo(gomega.BeEmpty())
		gomega.Expect(fixture.requestedPageSizes()).To(gomega.HaveEach("1000000"))

		served := len(fixture.requestedPageSizes())
		fixture.stopRejecting()

		tags, err := newTagsPageSizeAPI(fixture).Tags(context.Background(), host+"/repo")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(tags).To(gomega.Equal([]string{"latest"}))
		gomega.Expect(fixture.requestedPageSizes()[served:]).To(gomega.Equal([]string{"1000000"}))
	},
		ginkgo.Entry("persistent too many requests", http.StatusTooManyRequests, "TOOMANYREQUESTS", "too many requests"),
		ginkgo.Entry("persistent service unavailable", http.StatusServiceUnavailable, "UNAVAILABLE", "service unavailable"),
	)

	ginkgo.It("never caps the registry host when the listing cannot reach the registry", func() {
		unreachable := newTagsPageSizeFixture("latest")
		unreachable.server.Close()

		host := nextTagsPageSizeHost()
		_, err := newTagsPageSizeAPI(unreachable).Tags(context.Background(), host+"/repo")
		gomega.Expect(err).To(gomega.HaveOccurred())

		healthy := newTagsPageSizeFixture("latest")
		ginkgo.DeferCleanup(healthy.server.Close)

		tags, err := newTagsPageSizeAPI(healthy).Tags(context.Background(), host+"/repo")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(tags).To(gomega.Equal([]string{"latest"}))
		gomega.Expect(healthy.requestedPageSizes()).To(gomega.Equal([]string{"1000000"}))
	})

	ginkgo.It("carries the caller cancellation into the fallback request", func() {
		fixture := newTagsPageSizeFixture("latest")
		fixture.rejectAbove = 1000
		fixture.rejectStatus, fixture.rejectCode, fixture.rejectMessage = http.StatusBadRequest, "PAGINATION_NUMBER_INVALID", "invalid number of tags requested"
		ginkgo.DeferCleanup(fixture.server.Close)
		ctx, cancel := context.WithCancel(context.Background())
		ginkgo.DeferCleanup(cancel)
		fixture.onReject = cancel

		_, err := newTagsPageSizeAPI(fixture).Tags(ctx, nextTagsPageSizeHost()+"/repo")
		gomega.Expect(err).To(gomega.MatchError(context.Canceled))
		gomega.Expect(fixture.requestedPageSizes()).To(gomega.Equal([]string{"1000000"}))
	})

	ginkgo.It("carries a cancellation arriving during the fallback request", func() {
		fixture := newTagsPageSizeFixture("latest")
		fixture.rejectAbove = 1000
		fixture.rejectStatus, fixture.rejectCode, fixture.rejectMessage = http.StatusBadRequest, "PAGINATION_NUMBER_INVALID", "invalid number of tags requested"
		ginkgo.DeferCleanup(fixture.server.Close)
		ctx, cancel := context.WithCancel(context.Background())
		ginkgo.DeferCleanup(cancel)

		// Cancel only once the fallback request itself is being served, and let the handler return
		// as soon as the cancellation reaches the server, so that the listing cannot complete first.
		fixture.onPage = func(r *http.Request) {
			if r.URL.Query().Get("n") != "1000" {
				return
			}
			cancel()
			select {
			case <-r.Context().Done():
			case <-time.After(10 * time.Second): // safety net: never block the suite on a missing cancellation
			}
		}

		host := nextTagsPageSizeHost()
		_, err := newTagsPageSizeAPI(fixture).Tags(ctx, host+"/repo")
		gomega.Expect(err).To(gomega.MatchError(context.Canceled))
		gomega.Expect(fixture.requestedPageSizes()).To(gomega.Equal([]string{"1000000", "1000"}))

		fixture.onPage = nil
		_, err = newTagsPageSizeAPI(fixture).Tags(context.Background(), host+"/repo")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(fixture.requestedPageSizes()[2:]).To(gomega.Equal([]string{"1000000", "1000"}))
	})

	ginkgo.It("reads tags concurrently while the fallback page size is recorded", func() {
		fixture := newTagsPageSizeFixture("latest")
		fixture.rejectAbove = 1000
		fixture.rejectStatus, fixture.rejectCode, fixture.rejectMessage = http.StatusBadRequest, "PAGINATION_NUMBER_INVALID", "invalid number of tags requested"
		ginkgo.DeferCleanup(fixture.server.Close)

		registryAPI := newTagsPageSizeAPI(fixture)
		host := nextTagsPageSizeHost()
		var readers sync.WaitGroup
		results := make(chan error, 8)
		for i := range 8 {
			readers.Add(1)
			go func() {
				defer ginkgo.GinkgoRecover()
				defer readers.Done()
				_, err := registryAPI.Tags(context.Background(), fmt.Sprintf("%s/repo-%d", host, i))
				results <- err
			}()
		}
		readers.Wait()
		close(results)
		for err := range results {
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		}
	})
})
