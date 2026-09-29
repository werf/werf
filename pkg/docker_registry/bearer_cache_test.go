package docker_registry

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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
			if explicitLifetime {
				fmt.Fprint(w, `{"token":"fixture-token","expires_in":8}`)
				return
			}
			issuedAt := time.Now().Add(-52 * time.Second).Format(time.RFC3339Nano)
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
		}, 6*time.Second, 100*time.Millisecond).Should(Equal(int64(2)))
	},
		Entry("explicit eight-second lifetime", true),
		Entry("default 60-second lifetime shortened by issued_at", false),
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
		var workers sync.WaitGroup
		results := make(chan error, 12)
		for range 12 {
			workers.Add(1)
			go func() {
				defer workers.Done()
				_, err := registry.Tags(context.Background(), reference)
				results <- err
			}()
		}
		workers.Wait()
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
		lateResult := make(chan error, 1)
		go func() {
			_, err := registry.Tags(context.Background(), reference)
			lateResult <- err
		}()
		<-oldArrived
		_, err = registry.Tags(context.Background(), reference)
		Expect(err).NotTo(HaveOccurred())
		releaseOnce.Do(func() { close(releaseOld) })
		Expect(<-lateResult).NotTo(HaveOccurred())
		_, err = registry.Tags(context.Background(), reference)
		Expect(err).NotTo(HaveOccurred())
		Expect(fixture.exchanges.Load()).To(Equal(int64(2)))
	})

	It("separates registries sharing a token realm and scope", func() {
		first := newBearerRegistryFixture()
		second := newBearerRegistryFixture()
		DeferCleanup(first.server.Close)
		DeferCleanup(second.server.Close)
		second.realmURL = first.server.URL + "/token"
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
		result := make(chan error, 1)
		go func() {
			_, err := registry.Tags(ctx, reference)
			result <- err
		}()
		<-started
		cancel()
		Expect(<-result).To(HaveOccurred())
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
		tokenRequest, err := http.NewRequest(http.MethodGet, fixture.server.URL+"/token?scope="+url.QueryEscape("repository:repo:pull")+"&service=fixture", nil)
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
		Expect(<-leader).NotTo(HaveOccurred())
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
					results <- registry.writeToRemote(context.Background(), ref, empty.Index)
					return
				}
				results <- registry.writeToRemote(context.Background(), ref, empty.Image)
			}(tag)
		}
		workers.Wait()
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
		fixture := newWritableBearerRegistryFixture()
		DeferCleanup(fixture.server.Close)
		registry := newAPI(apiOptions{InsecureRegistry: true})
		base := strings.TrimPrefix(fixture.server.URL, "http://") + "/repo"
		ref, err := name.NewTag(base+":source", name.Insecure)
		Expect(err).NotTo(HaveOccurred())
		Expect(registry.writeToRemote(context.Background(), ref, empty.Image)).To(Succeed())
		Expect(registry.tagImage(context.Background(), ref.String(), "copy")).To(Succeed())
		tags, err := registry.Tags(context.Background(), base)
		Expect(err).NotTo(HaveOccurred())
		Expect(tags).To(ConsistOf("source", "copy"))
		Expect(registry.deleteImageByTag(context.Background(), base+":copy")).To(Succeed())
	})
})
