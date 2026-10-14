package docker_registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/samber/lo"

	"github.com/werf/werf/v3/pkg/werf"
)

const prefixTagsPageSize = 100

func (r *defaultImplementation) TagsByPrefix(ctx context.Context, reference, prefix string) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	host := r.extractRegistryHost(reference)
	_, unsupported := r.prefixUnsupported.Load(host)
	implementation := r.Implementation
	if implementation == DefaultImplementationName {
		detected, err := detectImplementation(reference)
		if err != nil {
			return nil, err
		}
		implementation = detected
	}
	eligible := implementation == DefaultImplementationName || implementation == GitLabRegistryImplementationName || implementation == HarborImplementationName
	if eligible && !unsupported && isHexTagPrefix(prefix) {
		tags, supported, err := r.tagsByPrefix(ctx, reference, prefix, implementation == DefaultImplementationName)
		if err != nil {
			return nil, fmt.Errorf("list tags by prefix: %w", err)
		}
		if supported {
			return tags, nil
		}
		r.prefixUnsupported.Store(host, struct{}{})
	}
	tags, err := r.Tags(ctx, reference)
	if err != nil {
		return nil, err
	}
	return lo.Filter(tags, func(tag string, _ int) bool { return strings.HasPrefix(tag, prefix) }), nil
}

func isHexTagPrefix(prefix string) bool {
	if prefix == "" {
		return false
	}
	for _, char := range prefix {
		if !(char >= '0' && char <= '9' || char >= 'a' && char <= 'f') {
			return false
		}
	}
	return true
}

func (api *api) ManifestExists(ctx context.Context, reference string) (bool, error) {
	ref, err := name.ParseReference(reference, api.parseReferenceOptionsForHost(reference)...)
	if err != nil {
		return false, fmt.Errorf("parse manifest reference: %w", err)
	}
	_, err = remote.Head(ref, api.defaultRemoteOptionsForHost(ctx, reference)...)
	var registryErr *transport.Error
	if errors.As(err, &registryErr) && (registryErr.StatusCode == http.StatusMethodNotAllowed || registryErr.StatusCode == http.StatusNotImplemented) {
		_, err = remote.Get(ref, api.defaultRemoteOptionsForHost(ctx, reference)...)
	}
	if IsStatusNotFoundErr(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check manifest existence: %w", err)
	}
	return true, nil
}

func (api *api) prefixListingClient(ctx context.Context, reference string, repo name.Repository) (*http.Client, error) {
	auth, err := authn.Resolve(ctx, authn.DefaultKeychain, repo)
	if err != nil {
		return nil, fmt.Errorf("resolve registry authentication: %w", err)
	}
	base, skipTLS := api.httpTransport, api.SkipTlsVerifyRegistry
	if !api.InsecureRegistry && !api.SkipTlsVerifyRegistry && api.shouldUseInsecureRegistry(reference) {
		base, skipTLS = api.insecureHttpTransport, true
	}
	base = newBearerTokenTransport(base, api.bearerTokens, api.extractRegistryHost(reference), skipTLS)
	base = transport.NewUserAgent(base, werf.UserAgent)
	authenticated, err := transport.NewWithContext(ctx, repo.Registry, auth, base, []string{repo.Scope(transport.PullScope)})
	if err != nil {
		return nil, fmt.Errorf("authenticate tag listing: %w", err)
	}
	return &http.Client{Transport: authenticated, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}, nil
}

func (api *api) tagsByPrefix(ctx context.Context, reference, prefix string, requireDeclaration bool) ([]string, bool, error) {
	repo, err := name.NewRepository(reference, api.newRepositoryOptionsForHost(reference)...)
	if err != nil {
		return nil, false, fmt.Errorf("parse tag repository: %w", err)
	}
	client, err := api.prefixListingClient(ctx, reference, repo)
	if err != nil {
		return nil, false, err
	}
	endpoint := url.URL{Scheme: repo.Registry.Scheme(), Host: repo.RegistryStr(), Path: "/v2/" + repo.RepositoryStr() + "/tags/list"}
	current := endpoint
	query := url.Values{"n": {fmt.Sprint(prefixTagsPageSize)}, "last": {prefix}}
	current.RawQuery = query.Encode()
	cursor := prefix
	asciiOrdered, foldedOrdered := true, true
	var result []string
	for {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, current.String(), nil)
		if err != nil {
			return nil, false, fmt.Errorf("create prefix listing request: %w", err)
		}
		response, err := client.Do(request)
		if err != nil {
			return nil, false, fmt.Errorf("request prefix listing: %w", err)
		}
		tags, next, supported, err := readPrefixPage(response, &endpoint, &current, requireDeclaration)
		if err != nil {
			return nil, false, err
		}
		if !supported {
			return nil, false, nil
		}
		previous := cursor
		for _, tag := range tags {
			asciiOrdered = asciiOrdered && tag > previous
			foldedOrdered = foldedOrdered && strings.ToLower(tag) > strings.ToLower(previous)
			previous = tag
			if strings.HasPrefix(tag, prefix) {
				result = append(result, tag)
			}
		}
		if !asciiOrdered || !foldedOrdered {
			return nil, false, nil
		}
		if next != nil && (len(tags) == 0 || next.Query().Get("last") != previous) {
			return nil, false, nil
		}
		boundary := len(tags) != 0 && !strings.HasPrefix(previous, prefix) &&
			(!asciiOrdered || previous > prefix) && (!foldedOrdered || strings.ToLower(previous) > prefix)
		if boundary || next == nil && len(tags) < prefixTagsPageSize {
			break
		}
		if len(tags) == 0 {
			return nil, false, nil
		}
		cursor = previous
		if next != nil {
			current = *next
		} else {
			query.Set("last", cursor)
			current = endpoint
			current.RawQuery = query.Encode()
		}
	}
	// The cursor is exclusive, while old stage formats also use the bare digest tag.
	exists, err := api.ManifestExists(ctx, repo.Tag(prefix).Name())
	if err != nil {
		return nil, false, err
	}
	if exists {
		result = append(result, prefix)
	}
	return result, true, nil
}

func readPrefixPage(response *http.Response, endpoint, current *url.URL, requireDeclaration bool) ([]string, *url.URL, bool, error) {
	defer response.Body.Close()
	switch response.StatusCode {
	case http.StatusBadRequest, http.StatusMethodNotAllowed, http.StatusNotImplemented:
		return nil, nil, false, nil
	case http.StatusNotFound:
		return nil, nil, true, nil
	}
	if err := transport.CheckError(response, http.StatusOK); err != nil {
		return nil, nil, false, fmt.Errorf("read prefix listing: %w", err)
	}
	if requireDeclaration && response.Header.Get("Docker-Distribution-API-Version") != "registry/2.0" {
		return nil, nil, false, nil
	}
	var page struct {
		Tags []string `json:"tags"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 4*1024*1024)).Decode(&page); err != nil {
		return nil, nil, false, fmt.Errorf("decode prefix listing: %w", err)
	}
	if len(page.Tags) > prefixTagsPageSize {
		return nil, nil, false, nil
	}
	var next *url.URL
	for _, header := range response.Header.Values("Link") {
		for _, part := range strings.Split(header, ",") {
			target, parameters, ok := strings.Cut(strings.TrimSpace(part), ";")
			if !ok || !(strings.Contains(parameters, `rel="next"`) || strings.Contains(parameters, "rel=next")) {
				return nil, nil, false, nil
			}
			if next != nil || !strings.HasPrefix(target, "<") || !strings.HasSuffix(target, ">") {
				return nil, nil, false, nil
			}
			parsed, err := url.Parse(strings.TrimSuffix(strings.TrimPrefix(target, "<"), ">"))
			if err != nil {
				return nil, nil, false, nil
			}
			next = current.ResolveReference(parsed)
			if next.Scheme != endpoint.Scheme || next.Host != endpoint.Host || next.Path != endpoint.Path || next.User != nil || next.Fragment != "" {
				return nil, nil, false, nil
			}
			if next.Query().Get("n") != fmt.Sprint(prefixTagsPageSize) {
				return nil, nil, false, nil
			}
		}
	}
	return page.Tags, next, true, nil
}
