package stage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/samber/lo"
)

type gitSource struct {
	Source     string  `json:"source"`
	Tag        *string `json:"tag"`
	Branch     *string `json:"branch"`
	Commit     string  `json:"commit"`
	CommitDate string  `json:"commitDate"`
	Add        string  `json:"add"`
	To         string  `json:"to"`
}

func GitSourcesJSON(ctx context.Context, mappings []*GitMapping) (string, error) {
	sources := make([]json.RawMessage, 0, len(mappings))
	for _, mapping := range mappings {
		source, err := mapping.gitSource(ctx)
		if err != nil {
			return "", fmt.Errorf("read git source metadata: %w", err)
		}
		data, err := json.Marshal(source)
		if err != nil {
			return "", fmt.Errorf("encode git source metadata: %w", err)
		}
		sources = append(sources, data)
	}
	sort.Slice(sources, func(i, j int) bool { return string(sources[i]) < string(sources[j]) })
	data, err := json.Marshal(sources)
	if err != nil {
		return "", fmt.Errorf("encode git sources metadata: %w", err)
	}
	return string(data), nil
}

func (gm *GitMapping) gitSource(ctx context.Context) (gitSource, error) {
	commit, err := gm.getCommit(ctx)
	if err != nil {
		return gitSource{}, fmt.Errorf("resolve git source commit: %w", err)
	}
	metadata, err := gm.GitRepo().CommitMetadata(ctx, commit)
	if err != nil {
		return gitSource{}, err
	}
	source := gitSource{Source: ".", Commit: metadata.Hash, CommitDate: metadata.CommitterDate.UTC().Format(time.RFC3339), Add: gm.Add, To: gm.To}
	if !gm.IsLocal() {
		origin, err := gm.GitRepo().RemoteOriginUrl(ctx)
		if err != nil {
			return gitSource{}, fmt.Errorf("read git source URL: %w", err)
		}
		source.Source, err = sanitizeGitSourceURL(origin)
		if err != nil {
			return gitSource{}, err
		}
	}
	if gm.Commit == "" {
		if gm.Tag != "" {
			source.Tag = lo.ToPtr(gm.Tag)
		} else if gm.Branch != "" {
			source.Branch = lo.ToPtr(gm.Branch)
		}
	}
	return source, nil
}

func sanitizeGitSourceURL(source string) (string, error) {
	if !strings.Contains(source, "://") {
		if host, _, ok := strings.Cut(source, ":"); ok && !strings.Contains(host, "/") {
			if i := strings.LastIndex(host, "@"); i >= 0 {
				source = source[i+1:]
			}
		}
		return strings.SplitN(strings.SplitN(source, "?", 2)[0], "#", 2)[0], nil
	}
	u, err := url.Parse(source)
	if err != nil {
		return "", fmt.Errorf("parse git source URL")
	}
	u.User = nil
	u.RawQuery = ""
	u.ForceQuery = false
	u.Fragment = ""
	u.RawFragment = ""
	return u.String(), nil
}
