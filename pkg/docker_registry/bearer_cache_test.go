package docker_registry

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("registry bearer reuse", func() {
	BeforeEach(func() {
		configDir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(configDir, "config.json"), []byte(`{"auths":{}}`), 0o600)).To(Succeed())
		GinkgoT().Setenv("DOCKER_CONFIG", configDir)
	})

	It("reuses a valid token across repeated API tag reads", func() {
		fixture := newBearerRegistryFixture()
		DeferCleanup(fixture.server.Close)
		fixture.tokenReply = func(w http.ResponseWriter, _ *http.Request, _ int64) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Cache-Control", "no-store")
			fmt.Fprint(w, `{"token":"fixture-token"}`)
		}
		registry := newAPI(apiOptions{InsecureRegistry: true})
		reference := strings.TrimPrefix(fixture.server.URL, "http://") + "/repo"

		for range 2 {
			tags, err := registry.Tags(context.Background(), reference)
			Expect(err).NotTo(HaveOccurred())
			Expect(tags).To(Equal([]string{"latest"}))
		}
		Expect(fixture.exchanges.Load()).To(Equal(int64(1)))
	})

	DescribeTable("renews a cached credential when its lifetime ends", func(explicitLifetime bool) {
		fixture := newBearerRegistryFixture()
		DeferCleanup(fixture.server.Close)
		fixture.tokenReply = func(w http.ResponseWriter, _ *http.Request, _ int64) {
			w.Header().Set("Content-Type", "application/json")
			issuedAt := time.Now().Add(-20 * time.Second).Format(time.RFC3339Nano)
			if explicitLifetime {
				fmt.Fprintf(w, `{"token":"fixture-token","expires_in":60,"issued_at":%q}`, issuedAt)
				return
			}
			fmt.Fprintf(w, `{"token":"fixture-token","issued_at":%q}`, issuedAt)
		}
		registry := newAPI(apiOptions{InsecureRegistry: true})
		reference := strings.TrimPrefix(fixture.server.URL, "http://") + "/repo"
		for range 2 {
			_, err := registry.Tags(context.Background(), reference)
			Expect(err).NotTo(HaveOccurred())
		}
		Expect(fixture.exchanges.Load()).To(Equal(int64(1)))
		Eventually(func() (int64, error) {
			_, err := registry.Tags(context.Background(), reference)
			return fixture.exchanges.Load(), err
		}, 15*time.Second, 100*time.Millisecond).Should(Equal(int64(2)))
	},
		Entry("explicit 60-second lifetime near the refresh margin", true),
		Entry("default 60-second lifetime near the refresh margin", false),
	)

	It("coalesces concurrent token-only exchanges", func() {
		fixture := newBearerRegistryFixture()
		DeferCleanup(fixture.server.Close)
		fixture.tokenReply = func(w http.ResponseWriter, _ *http.Request, _ int64) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"token":"fixture-token"}`)
		}
		registry := newAPI(apiOptions{InsecureRegistry: true})
		reference := strings.TrimPrefix(fixture.server.URL, "http://") + "/repo"
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		DeferCleanup(cancel)
		var workers sync.WaitGroup
		results := make(chan error, 12)
		for range 12 {
			workers.Add(1)
			go func() {
				defer workers.Done()
				_, err := registry.Tags(ctx, reference)
				results <- err
			}()
		}
		done := make(chan struct{})
		go func() { workers.Wait(); close(done) }()
		select {
		case <-done:
		case <-ctx.Done():
			Fail("concurrent tag reads did not finish")
		}
		close(results)
		for err := range results {
			Expect(err).NotTo(HaveOccurred())
		}
		Expect(fixture.exchanges.Load()).To(Equal(int64(1)))
	})

	It("uses separate tokens for separate repository scopes", func() {
		fixture := newBearerRegistryFixture()
		DeferCleanup(fixture.server.Close)
		var scopes sync.Map
		fixture.tokenReply = func(w http.ResponseWriter, r *http.Request, _ int64) {
			scopes.Store(r.URL.Query().Get("scope"), true)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"token":"fixture-token"}`)
		}
		registry := newAPI(apiOptions{InsecureRegistry: true})
		base := strings.TrimPrefix(fixture.server.URL, "http://")
		for _, repo := range []string{"repo-a", "repo-b", "repo-a", "repo-b"} {
			_, err := registry.Tags(context.Background(), base+"/"+repo)
			Expect(err).NotTo(HaveOccurred())
		}
		Expect(fixture.exchanges.Load()).To(Equal(int64(2)))
		_, first := scopes.Load("repository:repo-a:pull")
		_, second := scopes.Load("repository:repo-b:pull")
		Expect(first && second).To(BeTrue())
	})

	It("renews a bearer token rejected by the registry", func() {
		fixture := newBearerRegistryFixture()
		DeferCleanup(fixture.server.Close)
		var accepted atomic.Value
		accepted.Store("token-1")
		fixture.tokenReply = func(w http.ResponseWriter, _ *http.Request, count int64) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"token":"token-%d","expires_in":300}`, count)
		}
		fixture.authorize = func(r *http.Request) bool {
			return r.Header.Get("Authorization") == "Bearer "+accepted.Load().(string)
		}
		registry := newAPI(apiOptions{InsecureRegistry: true})
		reference := strings.TrimPrefix(fixture.server.URL, "http://") + "/repo"
		_, err := registry.Tags(context.Background(), reference)
		Expect(err).NotTo(HaveOccurred())
		accepted.Store("token-2")
		_, err = registry.Tags(context.Background(), reference)
		Expect(err).NotTo(HaveOccurred())
		_, err = registry.Tags(context.Background(), reference)
		Expect(err).NotTo(HaveOccurred())
		Expect(fixture.exchanges.Load()).To(Equal(int64(2)))
	})

	It("does not evict a replacement after a delayed rejection of the old token", func() {
		fixture := newBearerRegistryFixture()
		DeferCleanup(fixture.server.Close)
		fixture.tokenReply = func(w http.ResponseWriter, _ *http.Request, count int64) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"token":"token-%d","expires_in":300}`, count)
		}
		var rejecting atomic.Bool
		var oldRequests atomic.Int64
		oldArrived := make(chan struct{})
		releaseOld := make(chan struct{})
		var releaseOnce sync.Once
		DeferCleanup(func() { releaseOnce.Do(func() { close(releaseOld) }) })
		fixture.authorize = func(r *http.Request) bool {
			if r.Header.Get("Authorization") == "Bearer token-2" {
				return true
			}
			if r.Header.Get("Authorization") != "Bearer token-1" {
				return false
			}
			if !rejecting.Load() {
				return true
			}
			if oldRequests.Add(1) == 1 {
				close(oldArrived)
				<-releaseOld
			}
			return false
		}
		registry := newAPI(apiOptions{InsecureRegistry: true})
		reference := strings.TrimPrefix(fixture.server.URL, "http://") + "/repo"
		_, err := registry.Tags(context.Background(), reference)
		Expect(err).NotTo(HaveOccurred())
		rejecting.Store(true)
		lateCtx, cancelLate := context.WithCancel(context.Background())
		DeferCleanup(cancelLate)
		lateResult := make(chan error, 1)
		go func() {
			_, err := registry.Tags(lateCtx, reference)
			lateResult <- err
		}()
		select {
		case <-oldArrived:
		case <-time.After(3 * time.Second):
			Fail("old-token request did not reach the registry")
		}
		_, err = registry.Tags(context.Background(), reference)
		Expect(err).NotTo(HaveOccurred())
		releaseOnce.Do(func() { close(releaseOld) })
		select {
		case err := <-lateResult:
			Expect(err).NotTo(HaveOccurred())
		case <-time.After(3 * time.Second):
			Fail("delayed old-token request did not finish")
		}
		_, err = registry.Tags(context.Background(), reference)
		Expect(err).NotTo(HaveOccurred())
		Expect(fixture.exchanges.Load()).To(Equal(int64(2)))
	})

	It("separates registries sharing a token realm and scope", func() {
		first := newBearerRegistryFixture()
		second := newBearerRegistryFixture()
		DeferCleanup(first.server.Close)
		DeferCleanup(second.server.Close)
		first.realmURL = strings.Replace(first.server.URL, "127.0.0.1", "localhost", 1) + "/token"
		second.realmURL = first.realmURL
		registry := newAPI(apiOptions{InsecureRegistry: true})
		for range 2 {
			for _, fixture := range []*bearerRegistryFixture{first, second} {
				_, err := registry.Tags(context.Background(), strings.TrimPrefix(fixture.server.URL, "http://")+"/repo")
				Expect(err).NotTo(HaveOccurred())
			}
		}
		Expect(first.exchanges.Load()).To(Equal(int64(2)))
		Expect(second.exchanges.Load()).To(Equal(int64(0)))
	})

	It("refreshes rejected tokens after a redirect to a registry sharing the realm", func() {
		fixture := newBearerRegistryFixture()
		DeferCleanup(fixture.server.Close)
		fixture.realmURL = strings.Replace(fixture.server.URL, "127.0.0.1", "localhost", 1) + "/token"
		fixture.tokenReply = func(w http.ResponseWriter, _ *http.Request, count int64) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"token":"fixture-token-%d","expires_in":300}`, count)
		}
		fixture.authorize = func(r *http.Request) bool {
			return strings.HasPrefix(r.Header.Get("Authorization"), "Bearer fixture-token-")
		}
		var rejectedThrough atomic.Int64
		mirror := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := strconv.ParseInt(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer fixture-token-"), 10, 64)
			if err != nil || token <= rejectedThrough.Load() {
				w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer realm="%s",service="fixture",scope="repository:mirror/repo:pull"`, fixture.realmURL))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"name":"repo","tags":["latest"]}`)
		}))
		DeferCleanup(mirror.Close)
		mirrorURL := strings.Replace(mirror.URL, "127.0.0.1", "localhost", 1)
		fixture.backend = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, mirrorURL+"/v2/repo/tags/list", http.StatusTemporaryRedirect)
		})
		registry := newAPI(apiOptions{InsecureRegistry: true})
		reference := strings.TrimPrefix(fixture.server.URL, "http://") + "/repo"
		for range 2 {
			tags, err := registry.Tags(context.Background(), reference)
			Expect(err).NotTo(HaveOccurred())
			Expect(tags).To(Equal([]string{"latest"}))
		}
		rejectedThrough.Store(fixture.exchanges.Load())
		for range 2 {
			tags, err := registry.Tags(context.Background(), reference)
			Expect(err).NotTo(HaveOccurred())
			Expect(tags).To(Equal([]string{"latest"}))
		}
		Expect(fixture.exchanges.Load()).To(Equal(rejectedThrough.Load() + 2))
	})

	It("keeps caching the registry's own tokens when a redirected host rejects one", func() {
		const registryHost = "registry.example.test"
		const registryScope = "repository:repo:pull"
		const mirrorScope = "repository:mirror/repo:pull"
		transport, exchanges := newForeignChallengeTransport(registryHost, registryScope, mirrorScope)

		roundTripBearerGet(transport, "https://"+registryHost+"/v2/")
		roundTripBearerGet(transport, bearerTokenURL(registryScope))
		Expect(exchanges.Load()).To(Equal(int64(1)))

		roundTripBearerGet(transport, "https://mirror.example.test/v2/repo/tags/list")

		roundTripBearerGet(transport, bearerTokenURL(registryScope))
		Expect(exchanges.Load()).To(Equal(int64(1)))

		for i := range 2 {
			roundTripBearerGet(transport, bearerTokenURL(mirrorScope))
			Expect(exchanges.Load()).To(Equal(int64(2 + i)))
		}
	})

	It("keeps caching the registry's own tokens when a redirected host demands several scopes", func() {
		const registryHost = "registry.example.test"
		const registryScope = "repository:repo:pull"
		const mirrorScopes = "repository:mirror/repo:pull repository:mirror/other:pull"
		transport, exchanges := newForeignChallengeTransport(registryHost, registryScope, mirrorScopes)

		roundTripBearerGet(transport, "https://"+registryHost+"/v2/")
		roundTripBearerGet(transport, bearerTokenURL(registryScope))
		Expect(exchanges.Load()).To(Equal(int64(1)))

		roundTripBearerGet(transport, "https://mirror.example.test/v2/repo/tags/list")

		// go-containerregistry sends the challenge's scope value as a single scope query parameter.
		for i := range 2 {
			roundTripBearerGet(transport, bearerTokenURL(mirrorScopes))
			Expect(exchanges.Load()).To(Equal(int64(2 + i)))
		}

		roundTripBearerGet(transport, bearerTokenURL(registryScope))
		Expect(exchanges.Load()).To(Equal(int64(3)))
	})

	It("keeps caching the registry's own tokens when a redirected host sends a realmless challenge first", func() {
		const registryHost = "registry.example.test"
		const registryScope = "repository:repo:pull"
		const decoyScope = "repository:decoy:pull"
		const mirrorScope = "repository:mirror/repo:pull"
		// go-containerregistry exchanges through the first bearer challenge carrying a realm.
		transport, exchanges := newForeignChallengesTransport(registryHost, registryScope,
			fmt.Sprintf(`Bearer service="fixture",scope=%q`, decoyScope),
			bearerChallenge(mirrorScope),
		)

		roundTripBearerGet(transport, "https://"+registryHost+"/v2/")
		roundTripBearerGet(transport, bearerTokenURL(registryScope))
		Expect(exchanges.Load()).To(Equal(int64(1)))

		roundTripBearerGet(transport, "https://mirror.example.test/v2/repo/tags/list")

		for i := range 2 {
			roundTripBearerGet(transport, bearerTokenURL(mirrorScope))
			Expect(exchanges.Load()).To(Equal(int64(2 + i)))
		}

		roundTripBearerGet(transport, bearerTokenURL(registryScope))
		Expect(exchanges.Load()).To(Equal(int64(3)))
	})

	It("serves parallel rejections and token requests", func() {
		const registryHost = "registry.example.test"
		const registryScope = "repository:repo:pull"
		const mirrorScope = "repository:mirror/repo:pull"
		transport, exchanges := newForeignChallengeTransport(registryHost, registryScope, mirrorScope)

		roundTripBearerGet(transport, "https://"+registryHost+"/v2/")
		roundTripBearerGet(transport, "https://mirror.example.test/v2/repo/tags/list")

		var wg sync.WaitGroup
		for i := range 16 {
			wg.Add(2)
			go func() {
				defer GinkgoRecover()
				defer wg.Done()
				roundTripBearerGet(transport, fmt.Sprintf("https://mirror%d.example.test/v2/repo/tags/list", i))
			}()
			go func() {
				defer GinkgoRecover()
				defer wg.Done()
				roundTripBearerGet(transport, bearerTokenURL(mirrorScope))
			}()
		}
		wg.Wait()

		Expect(exchanges.Load()).To(Equal(int64(16)))

		for range 2 {
			roundTripBearerGet(transport, bearerTokenURL(registryScope))
		}
		Expect(exchanges.Load()).To(Equal(int64(17)))
	})

	DescribeTable("matches only equivalent registry authorities on bearer rejection", func(resourceURL string, expectedExchanges int64) {
		var exchanges atomic.Int64
		challenge := `Bearer realm="https://auth.example.test/token",service="fixture"`
		inner := bearerRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			resp := &http.Response{Request: req, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header)}
			switch req.URL.Path {
			case "/token":
				exchanges.Add(1)
				resp.StatusCode = http.StatusOK
				resp.Body = io.NopCloser(strings.NewReader(`{"token":"fixture-token","expires_in":300}`))
			case "/v2/", "/v2/repo/tags/list":
				resp.StatusCode = http.StatusUnauthorized
				resp.Header.Set("WWW-Authenticate", challenge)
			default:
				resp.StatusCode = http.StatusNotFound
			}
			return resp, nil
		})
		cache := newBearerTokenCache()
		transport := newBearerTokenTransport(inner, cache, "registry.example.test", false)
		ping, err := http.NewRequest(http.MethodGet, "https://registry.example.test/v2/", nil)
		Expect(err).NotTo(HaveOccurred())
		pingResponse, err := transport.RoundTrip(ping)
		Expect(err).NotTo(HaveOccurred())
		Expect(pingResponse.Body.Close()).To(Succeed())
		tokenRequest, err := http.NewRequest(http.MethodGet, "https://auth.example.test/token?service=fixture&scope=repository%3Arepo%3Apull", nil)
		Expect(err).NotTo(HaveOccurred())
		for range 2 {
			resp, err := transport.RoundTrip(tokenRequest)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Body.Close()).To(Succeed())
		}
		Expect(exchanges.Load()).To(Equal(int64(1)))
		resourceRequest, err := http.NewRequest(http.MethodGet, resourceURL, nil)
		Expect(err).NotTo(HaveOccurred())
		resourceRequest.Header.Set("Authorization", "Bearer fixture-token")
		resourceResponse, err := transport.RoundTrip(resourceRequest)
		Expect(err).NotTo(HaveOccurred())
		Expect(resourceResponse.Body.Close()).To(Succeed())
		transport = newBearerTokenTransport(inner, cache, "registry.example.test", false)
		pingResponse, err = transport.RoundTrip(ping)
		Expect(err).NotTo(HaveOccurred())
		Expect(pingResponse.Body.Close()).To(Succeed())
		resp, err := transport.RoundTrip(tokenRequest)
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.Body.Close()).To(Succeed())
		Expect(exchanges.Load()).To(Equal(expectedExchanges))
	},
		Entry("default HTTPS port", "https://registry.example.test:443/v2/repo/tags/list", int64(2)),
		Entry("different HTTPS port", "https://registry.example.test:444/v2/repo/tags/list", int64(1)),
		Entry("different scheme", "http://registry.example.test:443/v2/repo/tags/list", int64(1)),
		Entry("different scheme with default HTTP port", "http://registry.example.test/v2/repo/tags/list", int64(1)),
	)

	It("separates credentials changed for the same registry and scope", func() {
		fixture := newBearerRegistryFixture()
		DeferCleanup(fixture.server.Close)
		configDir := GinkgoT().TempDir()
		GinkgoT().Setenv("DOCKER_CONFIG", configDir)
		host := strings.TrimPrefix(fixture.server.URL, "http://")
		writeCredentials := func(user string) {
			auth := base64.StdEncoding.EncodeToString([]byte(user + ":password"))
			config := fmt.Sprintf(`{"auths":{%q:{"auth":%q}}}`, host, auth)
			Expect(os.WriteFile(filepath.Join(configDir, "config.json"), []byte(config), 0o600)).To(Succeed())
		}
		fixture.tokenReply = func(w http.ResponseWriter, r *http.Request, _ int64) {
			user, _, ok := r.BasicAuth()
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"token":%q,"expires_in":300}`, user)
		}
		fixture.authorize = func(r *http.Request) bool {
			return r.Header.Get("Authorization") == "Bearer first" || r.Header.Get("Authorization") == "Bearer second"
		}
		registry := newAPI(apiOptions{InsecureRegistry: true})
		for _, user := range []string{"first", "second", "first"} {
			writeCredentials(user)
			_, err := registry.Tags(context.Background(), host+"/repo")
			Expect(err).NotTo(HaveOccurred())
		}
		Expect(fixture.exchanges.Load()).To(Equal(int64(2)))
	})

	It("passes through OAuth fallback GETs without sharing identity tokens", func() {
		fixture := newBearerRegistryFixture()
		DeferCleanup(fixture.server.Close)
		configDir := GinkgoT().TempDir()
		GinkgoT().Setenv("DOCKER_CONFIG", configDir)
		host := strings.TrimPrefix(fixture.server.URL, "http://")
		config := fmt.Sprintf(`{"auths":{%q:{"identitytoken":"synthetic-identity"}}}`, host)
		Expect(os.WriteFile(filepath.Join(configDir, "config.json"), []byte(config), 0o600)).To(Succeed())
		var getCount atomic.Int64
		fixture.tokenReply = func(w http.ResponseWriter, r *http.Request, _ int64) {
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			getCount.Add(1)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"token":"fixture-token"}`)
		}
		registry := newAPI(apiOptions{InsecureRegistry: true})
		for range 2 {
			_, err := registry.Tags(context.Background(), host+"/repo")
			Expect(err).NotTo(HaveOccurred())
		}
		Expect(getCount.Load()).To(Equal(int64(2)))
	})

	It("separates TLS transport policies for the same registry", func() {
		fixture := newBearerRegistryFixtureWithTLS(true)
		DeferCleanup(fixture.server.Close)
		host := strings.TrimPrefix(fixture.server.URL, "https://")
		registry := newAPI(apiOptions{})
		registry.httpTransport = fixture.server.Client().Transport
		for range 2 {
			_, err := registry.Tags(context.Background(), host+"/repo")
			Expect(err).NotTo(HaveOccurred())
		}
		registry.insecureRegistryHosts[host] = true
		for range 2 {
			_, err := registry.Tags(context.Background(), host+"/repo")
			Expect(err).NotTo(HaveOccurred())
		}
		Expect(fixture.exchanges.Load()).To(Equal(int64(2)))
	})

	DescribeTable("does not retain unsupported or nearly expired responses", func(body string, wantError bool) {
		fixture := newBearerRegistryFixture()
		DeferCleanup(fixture.server.Close)
		fixture.tokenReply = func(w http.ResponseWriter, _ *http.Request, _ int64) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, body)
		}
		registry := newAPI(apiOptions{InsecureRegistry: true})
		reference := strings.TrimPrefix(fixture.server.URL, "http://") + "/repo"
		for range 2 {
			_, err := registry.Tags(context.Background(), reference)
			if wantError {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
			}
		}
		Expect(fixture.exchanges.Load()).To(Equal(int64(2)))
	},
		Entry("explicit zero lifetime", `{"token":"fixture-token","expires_in":0}`, false),
		Entry("explicit null lifetime", `{"token":"fixture-token","expires_in":null}`, false),
		Entry("malformed lifetime", `{"token":"fixture-token","expires_in":"300"}`, true),
		Entry("refresh token", `{"token":"fixture-token","expires_in":300,"refresh_token":"refresh"}`, false),
		Entry("earlier issue time", fmt.Sprintf(`{"token":"fixture-token","expires_in":60,"issued_at":%q}`, time.Now().Add(-2*time.Minute).Format(time.RFC3339)), false),
	)

	It("uses access_token before token when both are present", func() {
		fixture := newBearerRegistryFixture()
		DeferCleanup(fixture.server.Close)
		fixture.tokenReply = func(w http.ResponseWriter, _ *http.Request, _ int64) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"token":"wrong","access_token":"fixture-token","expires_in":300}`)
		}
		registry := newAPI(apiOptions{InsecureRegistry: true})
		reference := strings.TrimPrefix(fixture.server.URL, "http://") + "/repo"
		for range 2 {
			_, err := registry.Tags(context.Background(), reference)
			Expect(err).NotTo(HaveOccurred())
		}
		Expect(fixture.exchanges.Load()).To(Equal(int64(1)))
	})

	It("recovers after a failed token exchange", func() {
		fixture := newBearerRegistryFixture()
		DeferCleanup(fixture.server.Close)
		fixture.tokenReply = func(w http.ResponseWriter, _ *http.Request, count int64) {
			if count == 1 {
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, `{"error":`)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"token":"fixture-token"}`)
		}
		registry := newAPI(apiOptions{InsecureRegistry: true})
		reference := strings.TrimPrefix(fixture.server.URL, "http://") + "/repo"
		_, err := registry.Tags(context.Background(), reference)
		Expect(err).To(HaveOccurred())
		_, err = registry.Tags(context.Background(), reference)
		Expect(err).NotTo(HaveOccurred())
		Expect(fixture.exchanges.Load()).To(Equal(int64(2)))
	})

	It("preserves a one-shot error while reading a token response", func() {
		fixture := newBearerRegistryFixture()
		DeferCleanup(fixture.server.Close)
		registry := newAPI(apiOptions{InsecureRegistry: true})
		registry.httpTransport = &oneShotTokenReadErrorTransport{inner: registry.httpTransport}
		reference := strings.TrimPrefix(fixture.server.URL, "http://") + "/repo"
		for range 2 {
			_, err := registry.Tags(context.Background(), reference)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unexpected EOF"))
		}
		Expect(fixture.exchanges.Load()).To(Equal(int64(2)))
	})

	It("allows a live request after a canceled token exchange", func() {
		fixture := newBearerRegistryFixture()
		DeferCleanup(fixture.server.Close)
		started := make(chan struct{})
		fixture.tokenReply = func(w http.ResponseWriter, r *http.Request, count int64) {
			if count == 1 {
				close(started)
				<-r.Context().Done()
				return
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"token":"fixture-token"}`)
		}
		registry := newAPI(apiOptions{InsecureRegistry: true})
		reference := strings.TrimPrefix(fixture.server.URL, "http://") + "/repo"
		ctx, cancel := context.WithCancel(context.Background())
		DeferCleanup(cancel)
		result := make(chan error, 1)
		go func() {
			_, err := registry.Tags(ctx, reference)
			result <- err
		}()
		select {
		case <-started:
		case <-time.After(3 * time.Second):
			Fail("canceled token exchange did not start")
		}
		cancel()
		select {
		case err := <-result:
			Expect(err).To(HaveOccurred())
		case <-time.After(3 * time.Second):
			Fail("canceled token exchange did not finish")
		}
		_, err := registry.Tags(context.Background(), reference)
		Expect(err).NotTo(HaveOccurred())
		Expect(fixture.exchanges.Load()).To(Equal(int64(2)))
	})

	It("lets a waiting operation cancel without canceling the leader", func() {
		fixture := newBearerRegistryFixture()
		DeferCleanup(fixture.server.Close)
		started := make(chan struct{})
		release := make(chan struct{})
		var releaseOnce sync.Once
		DeferCleanup(func() { releaseOnce.Do(func() { close(release) }) })
		fixture.tokenReply = func(w http.ResponseWriter, _ *http.Request, count int64) {
			if count == 1 {
				close(started)
				<-release
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"token":"fixture-token"}`)
		}
		transport := newBearerTokenTransport(http.DefaultTransport, newBearerTokenCache(), strings.TrimPrefix(fixture.server.URL, "http://"), false)
		ping, err := http.NewRequest(http.MethodGet, fixture.server.URL+"/v2/", nil)
		Expect(err).NotTo(HaveOccurred())
		pingResponse, err := transport.RoundTrip(ping)
		Expect(err).NotTo(HaveOccurred())
		Expect(pingResponse.Body.Close()).To(Succeed())
		leaderCtx, cancelLeader := context.WithCancel(context.Background())
		DeferCleanup(cancelLeader)
		tokenRequest, err := http.NewRequestWithContext(leaderCtx, http.MethodGet, fixture.server.URL+"/token?scope="+url.QueryEscape("repository:repo:pull")+"&service=fixture", nil)
		Expect(err).NotTo(HaveOccurred())
		leader := make(chan error, 1)
		go func() {
			resp, err := transport.RoundTrip(tokenRequest)
			if err == nil {
				err = resp.Body.Close()
			}
			leader <- err
		}()
		select {
		case <-started:
		case <-time.After(3 * time.Second):
			Fail("leader did not start the token exchange")
		}
		waiterCtx, cancelWaiter := context.WithCancel(context.Background())
		DeferCleanup(cancelWaiter)
		observed := &observedDoneContext{Context: waiterCtx, entered: make(chan struct{})}
		waiter := make(chan error, 1)
		go func() {
			resp, err := transport.RoundTrip(tokenRequest.Clone(observed))
			if err == nil {
				err = resp.Body.Close()
			}
			waiter <- err
		}()
		select {
		case <-observed.entered:
		case <-time.After(3 * time.Second):
			releaseOnce.Do(func() { close(release) })
			Fail("waiter did not reach the occupied token key")
		}
		cancelWaiter()
		select {
		case err := <-waiter:
			Expect(err).To(MatchError(context.Canceled))
		case <-time.After(3 * time.Second):
			releaseOnce.Do(func() { close(release) })
			Fail("canceled waiter did not finish while the leader was blocked")
		}
		releaseOnce.Do(func() { close(release) })
		select {
		case err := <-leader:
			Expect(err).NotTo(HaveOccurred())
		case <-time.After(3 * time.Second):
			Fail("leader did not finish after token response was released")
		}
		resp, err := transport.RoundTrip(tokenRequest.Clone(context.Background()))
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.Body.Close()).To(Succeed())
		Expect(fixture.exchanges.Load()).To(Equal(int64(1)))
	})

	It("keeps sequential push progress independent while reusing authorization", func() {
		fixture := newWritableBearerRegistryFixture()
		DeferCleanup(fixture.server.Close)
		registry := newAPI(apiOptions{InsecureRegistry: true})
		base := strings.TrimPrefix(fixture.server.URL, "http://") + "/repo"
		for _, tag := range []string{"first", "second"} {
			ref, err := name.NewTag(base+":"+tag, name.Insecure)
			Expect(err).NotTo(HaveOccurred())
			Expect(registry.writeToRemote(context.Background(), ref, empty.Image)).To(Succeed())
		}
		Expect(fixture.exchanges.Load()).To(Equal(int64(1)))
		tags, err := registry.Tags(context.Background(), base)
		Expect(err).NotTo(HaveOccurred())
		Expect(tags).To(ConsistOf("first", "second"))
	})

	It("keeps concurrent image and index push progress independent", func() {
		fixture := newWritableBearerRegistryFixture()
		DeferCleanup(fixture.server.Close)
		registry := newAPI(apiOptions{InsecureRegistry: true})
		base := strings.TrimPrefix(fixture.server.URL, "http://") + "/repo"
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		DeferCleanup(cancel)
		var workers sync.WaitGroup
		results := make(chan error, 4)
		for _, tag := range []string{"image-a", "image-b", "image-c", "index"} {
			workers.Add(1)
			go func(tag string) {
				defer workers.Done()
				ref, err := name.NewTag(base+":"+tag, name.Insecure)
				if err != nil {
					results <- err
					return
				}
				if tag == "index" {
					results <- registry.writeToRemote(ctx, ref, empty.Index)
					return
				}
				results <- registry.writeToRemote(ctx, ref, empty.Image)
			}(tag)
		}
		done := make(chan struct{})
		go func() { workers.Wait(); close(done) }()
		select {
		case <-done:
		case <-ctx.Done():
			Fail("concurrent pushes did not finish")
		}
		close(results)
		for err := range results {
			Expect(err).NotTo(HaveOccurred())
		}
		Expect(fixture.exchanges.Load()).To(Equal(int64(1)))
		tags, err := registry.Tags(context.Background(), base)
		Expect(err).NotTo(HaveOccurred())
		Expect(tags).To(ConsistOf("image-a", "image-b", "image-c", "index"))
	})

	It("uses the same cache through tag and direct tag delete paths", func() {
		fixture := newWritableBearerRegistryFixtureWithTLS(true)
		DeferCleanup(fixture.server.Close)
		var scopes sync.Map
		fixture.tokenReply = func(w http.ResponseWriter, r *http.Request, _ int64) {
			scopes.Store(r.URL.Query().Get("scope"), true)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"token":"fixture-token","expires_in":300}`)
		}
		host := strings.TrimPrefix(fixture.server.URL, "https://")
		registry := newAPI(apiOptions{InsecureRegistryHosts: []string{host}})
		registry.httpTransport = fixture.server.Client().Transport
		base := host + "/repo"
		ref, err := name.NewTag(base + ":source")
		Expect(err).NotTo(HaveOccurred())
		Expect(registry.writeToRemote(context.Background(), ref, empty.Image)).To(Succeed())
		Expect(fixture.exchanges.Load()).To(Equal(int64(1)))
		for _, tag := range []string{"copy-a", "copy-b"} {
			Expect(registry.tagImage(context.Background(), ref.String(), tag)).To(Succeed())
		}
		tags, err := registry.Tags(context.Background(), base)
		Expect(err).NotTo(HaveOccurred())
		Expect(tags).To(ConsistOf("source", "copy-a", "copy-b"))
		for _, tag := range []string{"copy-a", "copy-b"} {
			Expect(registry.deleteImageByTag(context.Background(), base+":"+tag)).To(Succeed())
		}
		Expect(fixture.exchanges.Load()).To(Equal(int64(3)))
		_, pull := scopes.Load("repository:repo:pull")
		_, push := scopes.Load("repository:repo:push,pull")
		Expect(pull && push).To(BeTrue())
	})
})
