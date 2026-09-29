package docker_registry

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"golang.org/x/sync/singleflight"
)

type bearerRegistryFixture struct {
	server     *httptest.Server
	exchanges  atomic.Int64
	tokenReply func(http.ResponseWriter, *http.Request, int64)
	authorize  func(*http.Request) bool
	realmURL   string
	backend    http.Handler
}

func newBearerRegistryFixture() *bearerRegistryFixture {
	return newBearerRegistryFixtureWithTLS(false)
}

func newBearerRegistryFixtureWithTLS(useTLS bool) *bearerRegistryFixture {
	fixture := &bearerRegistryFixture{}
	fixture.authorize = func(r *http.Request) bool {
		return r.Header.Get("Authorization") == "Bearer fixture-token"
	}
	fixture.server = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		realm := fixture.realmURL
		if realm == "" {
			realm = fixture.server.URL + "/token"
		}
		switch r.URL.Path {
		case "/v2/":
			w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer realm="%s",service="fixture"`, realm))
			w.WriteHeader(http.StatusUnauthorized)
		case "/token":
			count := fixture.exchanges.Add(1)
			if fixture.tokenReply != nil {
				fixture.tokenReply(w, r, count)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"token":"fixture-token","expires_in":300}`)
		default:
			if fixture.backend != nil {
				if !fixture.authorize(r) {
					w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer realm="%s",service="fixture"`, realm))
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				fixture.backend.ServeHTTP(w, r)
				return
			}
			if !strings.HasPrefix(r.URL.Path, "/v2/") || !strings.HasSuffix(r.URL.Path, "/tags/list") {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			if !fixture.authorize(r) {
				w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer realm="%s",service="fixture"`, realm))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"name":"repo","tags":["latest"]}`)
		}
	}))
	if useTLS {
		fixture.server.StartTLS()
	} else {
		fixture.server.Start()
	}
	return fixture
}

func newWritableBearerRegistryFixture() *bearerRegistryFixture {
	return newWritableBearerRegistryFixtureWithTLS(false)
}

func newWritableBearerRegistryFixtureWithTLS(useTLS bool) *bearerRegistryFixture {
	fixture := newBearerRegistryFixtureWithTLS(useTLS)
	fixture.backend = registry.New()
	return fixture
}

type oneShotTokenReadErrorTransport struct {
	inner http.RoundTripper
}

var _ http.RoundTripper = (*oneShotTokenReadErrorTransport)(nil)

func (t *oneShotTokenReadErrorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.inner.RoundTrip(req)
	if err != nil || req.Method != http.MethodGet || req.URL.Path != "/token" || resp.StatusCode != http.StatusOK {
		return resp, err
	}
	if err := resp.Body.Close(); err != nil {
		return nil, err
	}
	resp.Body = &oneShotTokenReadErrorBody{data: []byte(`{"token":"fixture-token","expires_in":300}`)}
	return resp, nil
}

type oneShotTokenReadErrorBody struct {
	data []byte
}

var _ io.ReadCloser = (*oneShotTokenReadErrorBody)(nil)

func (body *oneShotTokenReadErrorBody) Read(dst []byte) (int, error) {
	if len(body.data) == 0 {
		return 0, io.EOF
	}
	n := copy(dst, body.data)
	body.data = nil
	return n, io.ErrUnexpectedEOF
}

func (body *oneShotTokenReadErrorBody) Close() error {
	return nil
}

type observedDoneContext struct {
	context.Context
	entered chan struct{}
	once    sync.Once
}

var _ context.Context = (*observedDoneContext)(nil)

func (ctx *observedDoneContext) Done() <-chan struct{} {
	ctx.once.Do(func() { close(ctx.entered) })
	return ctx.Context.Done()
}

type bearerRoundTripperFunc func(*http.Request) (*http.Response, error)

var _ http.RoundTripper = bearerRoundTripperFunc(nil)

func (roundTrip bearerRoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return roundTrip(req)
}

var _ Interface = (*listingRegistryStub)(nil)

type listingRegistryStub struct {
	Interface

	mu       sync.Mutex
	tags     []string
	calls    int
	started  chan struct{}
	release  chan struct{}
	failWith error
}

func newListingRegistryStub(tags ...string) *listingRegistryStub {
	return &listingRegistryStub{tags: tags}
}

func (r *listingRegistryStub) Tags(_ context.Context, _ string, _ ...Option) ([]string, error) {
	r.mu.Lock()
	r.calls++
	started, release := r.started, r.release
	r.mu.Unlock()

	if started != nil {
		close(started)
		<-release
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.failWith != nil {
		return nil, r.failWith
	}
	return append([]string(nil), r.tags...), nil
}

func (r *listingRegistryStub) parseReferenceParts(reference string) (referenceParts, error) {
	return (&api{}).parseReferenceParts(reference)
}

func (r *listingRegistryStub) setTags(tags ...string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tags = tags
}

func (r *listingRegistryStub) callCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.calls
}

func newCachedRegistryStub(inner Interface) *DockerRegistryWithCache {
	return &DockerRegistryWithCache{
		Interface:          inner,
		cachedTagsMap:      &sync.Map{},
		listTagsQueryGroup: &singleflight.Group{},
	}
}

func cachedEntry(r *DockerRegistryWithCache, cachedTagsID string) tagsCacheEntry {
	value, ok := r.cachedTagsMap.Load(cachedTagsID)
	gomega.Expect(ok).To(gomega.BeTrue())
	entry, err := castTagsEntry(value)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return entry
}

func startBlockedListing(ctx context.Context, r *DockerRegistryWithCache, inner *listingRegistryStub, reference string) (chan []string, func()) {
	inner.started, inner.release = make(chan struct{}), make(chan struct{})

	listedTags := make(chan []string, 1)
	go func() {
		defer ginkgo.GinkgoRecover()
		tags, err := r.Tags(ctx, reference)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		listedTags <- tags
	}()

	gomega.Eventually(inner.started).Should(gomega.BeClosed())
	return listedTags, func() { close(inner.release) }
}
