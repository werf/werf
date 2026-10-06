package docker_registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"golang.org/x/sync/singleflight"

	"github.com/werf/werf/v3/pkg/opstats"
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

type tagsPageSizeFixture struct {
	server       *httptest.Server
	httpRequests atomic.Int64

	mu      sync.Mutex
	queries []url.Values

	tags  []string
	chunk int

	rejectAbove         int
	rejectAfterRequests int
	rejectLimit         int
	rejections          int
	rejectStatus        int
	rejectCode          string
	rejectMessage       string
	onReject            func()
	onPage              func(*http.Request)
}

func newTagsPageSizeFixture(tags ...string) *tagsPageSizeFixture {
	fixture := &tagsPageSizeFixture{tags: tags}
	fixture.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer ginkgo.GinkgoRecover()
		fixture.httpRequests.Add(1)

		if r.URL.Path == "/v2/" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if !strings.HasSuffix(r.URL.Path, "/tags/list") {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		query := r.URL.Query()
		requestedPageSize := 0
		if rawPageSize := query.Get("n"); rawPageSize != "" {
			parsedPageSize, err := strconv.Atoi(rawPageSize)
			if err != nil {
				http.Error(w, fmt.Sprintf("invalid page size %q", rawPageSize), http.StatusBadRequest)
				return
			}
			requestedPageSize = parsedPageSize
		}

		fixture.mu.Lock()
		fixture.queries = append(fixture.queries, query)
		reject := fixture.rejectStatus != 0 &&
			requestedPageSize > fixture.rejectAbove &&
			len(fixture.queries) > fixture.rejectAfterRequests &&
			(fixture.rejectLimit == 0 || fixture.rejections < fixture.rejectLimit)
		rejectStatus, rejectCode, rejectMessage := fixture.rejectStatus, fixture.rejectCode, fixture.rejectMessage
		if reject {
			fixture.rejections++
		}
		fixture.mu.Unlock()

		if reject {
			if fixture.onReject != nil {
				fixture.onReject()
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(rejectStatus)
			_, err := fmt.Fprintf(w, `{"errors":[{"code":%q,"message":%q}]}`, rejectCode, rejectMessage)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			return
		}

		if fixture.onPage != nil {
			fixture.onPage(r)
		}

		fixture.writePage(w, r)
	}))
	return fixture
}

func (fixture *tagsPageSizeFixture) stopRejecting() {
	fixture.mu.Lock()
	defer fixture.mu.Unlock()
	fixture.rejectStatus = 0
}

func (fixture *tagsPageSizeFixture) writePage(w http.ResponseWriter, r *http.Request) {
	start := 0
	if last := r.URL.Query().Get("last"); last != "" {
		for i, tag := range fixture.tags {
			if tag == last {
				start = i + 1
				break
			}
		}
	}

	end := len(fixture.tags)
	if fixture.chunk > 0 && start+fixture.chunk < end {
		end = start + fixture.chunk
	}
	page := fixture.tags[start:end]

	w.Header().Set("Content-Type", "application/json")
	if end < len(fixture.tags) {
		w.Header().Set("Link", fmt.Sprintf(`<http://%s%s?n=%d&last=%s>; rel="next"`, r.Host, r.URL.Path, fixture.chunk, page[len(page)-1]))
	}
	gomega.Expect(json.NewEncoder(w).Encode(map[string]any{"name": "repo", "tags": page})).To(gomega.Succeed())
}

func (fixture *tagsPageSizeFixture) requestedPageSizes() []string {
	fixture.mu.Lock()
	defer fixture.mu.Unlock()
	sizes := make([]string, 0, len(fixture.queries))
	for _, query := range fixture.queries {
		sizes = append(sizes, query.Get("n"))
	}
	return sizes
}

type tagsPageSizeTransport struct {
	serverHost string
	inner      http.RoundTripper
}

var _ http.RoundTripper = (*tagsPageSizeTransport)(nil)

func (t *tagsPageSizeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	routed := req.Clone(req.Context())
	routed.Host = req.URL.Host
	routed.URL.Scheme = "http"
	routed.URL.Host = t.serverHost
	return t.inner.RoundTrip(routed)
}

var tagsPageSizeHosts atomic.Int64

func nextTagsPageSizeHost() string {
	return fmt.Sprintf("rejecting-%d.example.test", tagsPageSizeHosts.Add(1))
}

func newTagsPageSizeAPI(fixture *tagsPageSizeFixture) *api {
	return newTagsPageSizeAPIForImplementation(fixture, DefaultImplementationName)
}

func newTagsPageSizeAPIForImplementation(fixture *tagsPageSizeFixture, implementation string) *api {
	registryImplementation, err := newDefaultAPIForImplementation(implementation, defaultImplementationOptions{apiOptions{InsecureRegistry: true}})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	registryImplementation.api.httpTransport = &tagsPageSizeTransport{
		serverHost: strings.TrimPrefix(fixture.server.URL, "http://"),
		inner:      registryImplementation.api.httpTransport,
	}
	return registryImplementation.api
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

// newForeignChallengeTransport returns a bearer transport talking to a registry that answers every
// request with its own bearer challenge, while any other host answers with foreignScope's
// challenge, plus the number of token exchanges served.
func newForeignChallengeTransport(registryHost, registryScope, foreignScope string) (*bearerTokenTransport, *atomic.Int64) {
	return newForeignChallengesTransport(registryHost, registryScope, bearerChallenge(foreignScope))
}

func bearerChallenge(scope string) string {
	return fmt.Sprintf(`Bearer realm="https://auth.example.test/token",service="fixture",scope=%q`, scope)
}

// newForeignChallengesTransport is newForeignChallengeTransport with the foreign host answering
// with several WWW-Authenticate challenges, given verbatim.
func newForeignChallengesTransport(registryHost, registryScope string, foreignChallenges ...string) (*bearerTokenTransport, *atomic.Int64) {
	var exchanges atomic.Int64
	inner := bearerRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		resp := &http.Response{Request: req, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header)}
		switch {
		case req.URL.Path == "/token":
			exchanges.Add(1)
			resp.StatusCode = http.StatusOK
			resp.Body = io.NopCloser(strings.NewReader(`{"token":"fixture-token","expires_in":300}`))
		case req.URL.Host == registryHost:
			resp.StatusCode = http.StatusUnauthorized
			resp.Header.Set("WWW-Authenticate", bearerChallenge(registryScope))
		default:
			resp.StatusCode = http.StatusUnauthorized
			for _, challenge := range foreignChallenges {
				resp.Header.Add("WWW-Authenticate", challenge)
			}
		}
		return resp, nil
	})

	return newBearerTokenTransport(inner, newBearerTokenCache(), registryHost, false), &exchanges
}

func bearerTokenURL(scope string) string {
	return "https://auth.example.test/token?service=fixture&scope=" + url.QueryEscape(scope)
}

func roundTripBearerGet(transport http.RoundTripper, requestURL string) {
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	resp, err := transport.RoundTrip(req)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(resp.Body.Close()).To(gomega.Succeed())
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

type snapshotListingRegistry struct {
	Interface
	calls   atomic.Int32
	started chan struct{}
	release chan struct{}
}

var _ Interface = (*snapshotListingRegistry)(nil)

func (r *snapshotListingRegistry) Tags(_ context.Context, _ string, _ ...Option) ([]string, error) {
	if r.calls.Add(1) == 1 {
		close(r.started)
		<-r.release
		return []string{"before"}, nil
	}
	return []string{"before", "winner"}, nil
}

func (r *snapshotListingRegistry) parseReferenceParts(reference string) (referenceParts, error) {
	return (&api{}).parseReferenceParts(reference)
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

type registryRetryFixture struct {
	api           *api
	reference     string
	closedAddress string
	tagStatus     int
	tagDials      atomic.Int64
	failure       func(context.Context, int) error
}

func newRegistryRetryFixture() *registryRetryFixture {
	fixture := &registryRetryFixture{}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	fixture.closedAddress = listener.Addr().String()
	gomega.Expect(listener.Close()).To(gomega.Succeed())
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer ginkgo.GinkgoRecover()
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/v2/" {
			_, err := io.WriteString(w, "{}")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			return
		}
		if fixture.tagStatus != 0 {
			w.WriteHeader(fixture.tagStatus)
		}
		_, err := io.WriteString(w, `{"name":"fixture","tags":["ok"]}`)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}))
	ginkgo.DeferCleanup(server.Close)
	original := remote.DefaultTransport
	base := original.(*http.Transport).Clone()
	base.DisableKeepAlives = true
	base.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		defer ginkgo.GinkgoRecover()
		if ctx.Value(registryRetryContextKey{}) == true {
			attempt := fixture.tagDials.Add(1)
			if err := fixture.failure(ctx, int(attempt)); err != nil {
				return nil, err
			}
		}
		return (&net.Dialer{}).DialContext(ctx, network, address)
	}
	remote.DefaultTransport = base
	ginkgo.DeferCleanup(func() { remote.DefaultTransport = original })
	fixture.api = newAPI(apiOptions{InsecureRegistry: true})
	fixture.api.httpTransport = &registryRetryRequestTransport{fixture.api.httpTransport}
	fixture.api.insecureHttpTransport = &registryRetryRequestTransport{fixture.api.insecureHttpTransport}
	fixture.reference = strings.TrimPrefix(server.URL, "http://") + "/fixture"
	return fixture
}

func (f *registryRetryFixture) refusedDial(ctx context.Context) error {
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", f.closedAddress)
	if conn != nil {
		gomega.Expect(conn.Close()).To(gomega.Succeed())
	}
	gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("connection refused")))
	return err
}

type (
	registryRetryContextKey       struct{}
	registryRetryRequestTransport struct{ underlying http.RoundTripper }
)

var _ http.RoundTripper = (*registryRetryRequestTransport)(nil)

func (t *registryRetryRequestTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if strings.HasSuffix(req.URL.Path, "/tags/list") {
		req = req.WithContext(context.WithValue(req.Context(), registryRetryContextKey{}, true))
	}
	return t.underlying.RoundTrip(req)
}
