package docker_registry

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"golang.org/x/sync/singleflight"

	"github.com/werf/werf/v2/pkg/opstats"
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
	failWith error
}

func newListingRegistryStub(tags ...string) *listingRegistryStub {
	return &listingRegistryStub{tags: tags}
}

func (r *listingRegistryStub) Tags(_ context.Context, _ string, _ ...Option) ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls++
	if r.failWith != nil {
		return nil, r.failWith
	}
	return append([]string(nil), r.tags...), nil
}

func (r *listingRegistryStub) parseReferenceParts(reference string) (referenceParts, error) {
	return (&api{}).parseReferenceParts(reference)
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

func collectingContext(ctx context.Context) (context.Context, *opstats.Collector) {
	collector := opstats.NewCollector()
	return opstats.NewContext(ctx, collector), collector
}

func tagsListCounters(collector *opstats.Collector, ctx context.Context) opstats.CacheSummary {
	summary := collector.CacheSummary(ctx)
	gomega.Expect(summary).To(gomega.HaveLen(1))
	gomega.Expect(summary[0].Operation).To(gomega.Equal(opstats.OperationRegistryTagsList))
	return summary[0]
}

// admittingRegistryStub holds the listing of the caller that started it until the test admits
// it, so that joiners are admitted by a barrier on observed state instead of by a sleep.
type admittingRegistryStub struct {
	Interface

	tags     []string
	admitted chan struct{}
	admit    sync.Once

	mu    sync.Mutex
	calls int
}

var _ Interface = (*admittingRegistryStub)(nil)

func newAdmittingRegistryStub(tags ...string) *admittingRegistryStub {
	return &admittingRegistryStub{tags: tags, admitted: make(chan struct{})}
}

func (r *admittingRegistryStub) Tags(_ context.Context, _ string, _ ...Option) ([]string, error) {
	r.mu.Lock()
	r.calls++
	r.mu.Unlock()

	<-r.admitted

	return append([]string(nil), r.tags...), nil
}

func (r *admittingRegistryStub) parseReferenceParts(reference string) (referenceParts, error) {
	return (&api{}).parseReferenceParts(reference)
}

func (r *admittingRegistryStub) callCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.calls
}

func (r *admittingRegistryStub) release() {
	r.admit.Do(func() { close(r.admitted) })
}

// joinersWaitingForTagsList counts the goroutines that have already registered with singleflight
// as waiters of a tags listing started by someone else. The caller that started it is parked in
// the stub instead of in WaitGroup.Wait, so it is not counted, and a caller that merely entered
// Do without registering yet has no Wait frame either.
func joinersWaitingForTagsList() int {
	return goroutinesWithFrames(goroutineDump(),
		"(*DockerRegistryWithCache).getTagsListFromRegistry(",
		"singleflight.(*Group).Do(",
		"sync.(*WaitGroup).Wait(")
}

func goroutineDump() string {
	for size := 1 << 20; size <= 8<<20; size *= 2 {
		buf := make([]byte, size)
		if n := runtime.Stack(buf, true); n < size {
			return string(buf[:n])
		}
	}
	ginkgo.Fail("the goroutine dump does not fit in 8MiB")
	return ""
}

// goroutinesWithFrames counts the per-goroutine stacks of a runtime.Stack dump, which are
// separated by a blank line, that contain every one of the given frames.
func goroutinesWithFrames(dump string, frames ...string) int {
	var count int
	for _, stack := range strings.Split(dump, "\n\n") {
		matched := true
		for _, frame := range frames {
			if !strings.Contains(stack, frame) {
				matched = false
				break
			}
		}
		if matched {
			count++
		}
	}
	return count
}
