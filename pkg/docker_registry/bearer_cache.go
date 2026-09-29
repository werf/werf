package docker_registry

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	authchallenge "github.com/docker/distribution/registry/client/auth/challenge"
)

const (
	bearerTokenCapacity     = 128
	bearerTokenBodyLimit    = 64 << 10
	bearerTokenExpiryMargin = 30 * time.Second
)

type bearerTokenEntry struct {
	token  string
	expiry time.Time
	target string
	policy bool
}

type bearerTokenCache struct {
	mu       sync.Mutex
	entries  map[[32]byte]bearerTokenEntry
	inflight map[[32]byte]chan struct{}
}

func newBearerTokenCache() *bearerTokenCache {
	return &bearerTokenCache{
		entries:  make(map[[32]byte]bearerTokenEntry),
		inflight: make(map[[32]byte]chan struct{}),
	}
}

type bearerTokenTransport struct {
	inner  http.RoundTripper
	cache  *bearerTokenCache
	target string
	policy bool

	mu            sync.Mutex
	realm         *url.URL
	scheme        string
	cacheDisabled bool
}

var _ http.RoundTripper = (*bearerTokenTransport)(nil)

func newBearerTokenTransport(inner http.RoundTripper, cache *bearerTokenCache, target string, policy bool) *bearerTokenTransport {
	return &bearerTokenTransport{inner: inner, cache: cache, target: target, policy: policy}
}

func (t *bearerTokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Method == http.MethodPost {
		t.mu.Lock()
		if t.realm != nil && req.URL.Scheme == t.realm.Scheme && req.URL.Host == t.realm.Host && req.URL.Path == t.realm.Path {
			t.cacheDisabled = true
		}
		t.mu.Unlock()
	}
	if t.isTokenRequest(req) {
		return t.tokenRoundTrip(req)
	}

	resp, err := t.inner.RoundTrip(req)
	if err != nil {
		return resp, err
	}
	if !t.isRegistryRequest(req) {
		if resp.StatusCode == http.StatusUnauthorized {
			for _, challenge := range authchallenge.ResponseChallenges(resp) {
				if strings.EqualFold(challenge.Scheme, "bearer") {
					// Redirected hosts exchange tokens through this transport too, but their rejections cannot invalidate the origin's cache.
					t.mu.Lock()
					t.cacheDisabled = true
					t.mu.Unlock()
					break
				}
			}
		}
		return resp, nil
	}

	if req.Method == http.MethodGet && req.URL.Path == "/v2/" {
		for _, challenge := range authchallenge.ResponseChallenges(resp) {
			if !strings.EqualFold(challenge.Scheme, "bearer") {
				continue
			}
			realm, err := url.Parse(challenge.Parameters["realm"])
			if err == nil && realm.IsAbs() && realm.Host != "" {
				t.mu.Lock()
				t.realm = realm
				t.scheme = req.URL.Scheme
				t.mu.Unlock()
			}
			break
		}
	}

	if resp.StatusCode == http.StatusUnauthorized && strings.HasPrefix(req.Header.Get("Authorization"), "Bearer ") {
		for _, challenge := range authchallenge.ResponseChallenges(resp) {
			if strings.EqualFold(challenge.Scheme, "bearer") {
				t.cache.invalidate(t.target, t.policy, strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer "))
				break
			}
		}
	}
	return resp, nil
}

func (t *bearerTokenTransport) isRegistryRequest(req *http.Request) bool {
	t.mu.Lock()
	scheme := t.scheme
	t.mu.Unlock()
	if scheme != "" && req.URL.Scheme != scheme {
		return false
	}
	registry := &url.URL{Scheme: req.URL.Scheme, Host: t.target}
	return strings.EqualFold(req.URL.Hostname(), registry.Hostname()) && registryPort(req.URL) == registryPort(registry)
}

func registryPort(u *url.URL) string {
	if port := u.Port(); port != "" {
		return port
	}
	if u.Scheme == "http" {
		return "80"
	}
	if u.Scheme == "https" {
		return "443"
	}
	return ""
}

func (t *bearerTokenTransport) isTokenRequest(req *http.Request) bool {
	if req.Method != http.MethodGet {
		return false
	}
	t.mu.Lock()
	realm := t.realm
	cacheDisabled := t.cacheDisabled
	t.mu.Unlock()
	if realm == nil || cacheDisabled || req.URL.Scheme != realm.Scheme || req.URL.Host != realm.Host || req.URL.Path != realm.Path {
		return false
	}
	for key, values := range realm.Query() {
		actual := req.URL.Query()[key]
		if len(actual) < len(values) {
			return false
		}
		for _, value := range values {
			found := false
			for _, candidate := range actual {
				if candidate == value {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}
	return true
}

func (t *bearerTokenTransport) tokenRoundTrip(req *http.Request) (*http.Response, error) {
	t.mu.Lock()
	scheme := t.scheme
	t.mu.Unlock()
	identity, err := json.Marshal(struct {
		Target string
		Scheme string
		Policy bool
		URL    string
		Host   string
		Header http.Header
	}{t.target, scheme, t.policy, req.URL.String(), req.Host, req.Header})
	if err != nil {
		return t.inner.RoundTrip(req)
	}
	key := sha256.Sum256(identity)
	for {
		t.cache.mu.Lock()
		if entry, ok := t.cache.entries[key]; ok {
			if time.Now().Add(bearerTokenExpiryMargin).Before(entry.expiry) {
				t.cache.mu.Unlock()
				return bearerTokenResponse(req, entry.token)
			}
			delete(t.cache.entries, key)
		}
		if pending, ok := t.cache.inflight[key]; ok {
			t.cache.mu.Unlock()
			select {
			case <-pending:
				continue
			case <-req.Context().Done():
				return nil, req.Context().Err()
			}
		}
		pending := make(chan struct{})
		t.cache.inflight[key] = pending
		t.cache.mu.Unlock()

		started := time.Now()
		resp, err := t.inner.RoundTrip(req)
		if err == nil && resp != nil {
			var token string
			var expiry time.Time
			resp, token, expiry = inspectBearerTokenResponse(resp, started)
			if token != "" && req.Context().Err() == nil {
				t.cache.store(key, bearerTokenEntry{token: token, expiry: expiry, target: t.target, policy: t.policy})
			}
		}
		t.cache.mu.Lock()
		delete(t.cache.inflight, key)
		close(pending)
		t.cache.mu.Unlock()
		return resp, err
	}
}

func bearerTokenResponse(req *http.Request, token string) (*http.Response, error) {
	body, err := json.Marshal(struct {
		Token string `json:"token"`
	}{Token: token})
	if err != nil {
		return nil, fmt.Errorf("marshal cached bearer response: %w", err)
	}
	return &http.Response{
		StatusCode:    http.StatusOK,
		Status:        "200 OK",
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        http.Header{"Content-Type": []string{"application/json"}},
		Body:          io.NopCloser(bytes.NewReader(body)),
		ContentLength: int64(len(body)),
		Request:       req,
	}, nil
}

func (c *bearerTokenCache) store(key [32]byte, entry bearerTokenEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	for existingKey, existing := range c.entries {
		if !now.Add(bearerTokenExpiryMargin).Before(existing.expiry) {
			delete(c.entries, existingKey)
		}
	}
	if len(c.entries) >= bearerTokenCapacity {
		var oldestKey [32]byte
		var oldest time.Time
		for existingKey, existing := range c.entries {
			if oldest.IsZero() || existing.expiry.Before(oldest) {
				oldestKey, oldest = existingKey, existing.expiry
			}
		}
		delete(c.entries, oldestKey)
	}
	c.entries[key] = entry
}

func (c *bearerTokenCache) invalidate(target string, policy bool, token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key, entry := range c.entries {
		if entry.target == target && entry.policy == policy && entry.token == token {
			delete(c.entries, key)
		}
	}
}

type bearerResponseBody struct {
	io.Reader
	io.Closer
}

var _ io.ReadCloser = bearerResponseBody{}

type bearerResponseReadError struct {
	err error
}

var _ io.Reader = bearerResponseReadError{}

func (r bearerResponseReadError) Read([]byte) (int, error) {
	return 0, r.err
}

func inspectBearerTokenResponse(resp *http.Response, started time.Time) (*http.Response, string, time.Time) {
	if resp.StatusCode != http.StatusOK || resp.Body == nil {
		return resp, "", time.Time{}
	}
	prefix, err := io.ReadAll(io.LimitReader(resp.Body, bearerTokenBodyLimit+1))
	if err != nil {
		resp.Body = bearerResponseBody{Reader: io.MultiReader(bytes.NewReader(prefix), bearerResponseReadError{err: err}), Closer: resp.Body}
		return resp, "", time.Time{}
	}
	if len(prefix) > bearerTokenBodyLimit {
		resp.Body = bearerResponseBody{Reader: io.MultiReader(bytes.NewReader(prefix), resp.Body), Closer: resp.Body}
		return resp, "", time.Time{}
	}
	resp.Body = bearerResponseBody{Reader: bytes.NewReader(prefix), Closer: resp.Body}
	var payload struct {
		Token        string          `json:"token"`
		AccessToken  string          `json:"access_token"`
		RefreshToken string          `json:"refresh_token"`
		ExpiresIn    json.RawMessage `json:"expires_in"`
		IssuedAt     string          `json:"issued_at"`
	}
	if json.Unmarshal(prefix, &payload) != nil || payload.RefreshToken != "" {
		return resp, "", time.Time{}
	}
	seconds := int64(60)
	if len(payload.ExpiresIn) != 0 {
		var explicit *int64
		if json.Unmarshal(payload.ExpiresIn, &explicit) != nil || explicit == nil || *explicit <= 0 {
			return resp, "", time.Time{}
		}
		seconds = *explicit
	}
	if seconds > math.MaxInt64/int64(time.Second) {
		return resp, "", time.Time{}
	}
	token := payload.Token
	if payload.AccessToken != "" {
		token = payload.AccessToken
	}
	if token == "" {
		return resp, "", time.Time{}
	}
	expiry := started.Add(time.Duration(seconds) * time.Second)
	if issuedAt, err := time.Parse(time.RFC3339, payload.IssuedAt); err == nil && issuedAt.Before(started) {
		issuedExpiry := issuedAt.Add(time.Duration(seconds) * time.Second)
		if issuedExpiry.Before(expiry) {
			expiry = issuedExpiry
		}
	}
	if !time.Now().Add(bearerTokenExpiryMargin).Before(expiry) {
		return resp, "", time.Time{}
	}
	return resp, token, expiry
}
