package common

import (
	"context"
	"strings"
	"testing"

	"github.com/werf/werf/v2/pkg/storage"
)

// cmdDataFor builds a CmdData with all registry-model flag pointers initialized
// to the given values (mimics what the Setup* funcs produce after parsing).
func cmdDataFor(repo string, cacheFrom, cacheTo, imagesRepo []string, meta, secondary, finalRepo string) *CmdData {
	repoAddr := repo
	metaAddr := meta
	finalAddr := finalRepo
	cf := append([]string{}, cacheFrom...)
	ct := append([]string{}, cacheTo...)
	ir := append([]string{}, imagesRepo...)
	var sec []string
	if secondary != "" {
		sec = []string{secondary}
	}
	return &CmdData{
		Repo:                   &RepoData{Address: &repoAddr},
		FinalRepo:              &RepoData{Address: &finalAddr},
		MetaRepo:               &metaAddr,
		CacheFrom:              &cf,
		CacheTo:                &ct,
		ImagesRepo:             &ir,
		SecondaryStagesStorage: &sec,
	}
}

func TestResolveRepos_MutualExclusion(t *testing.T) {
	c := cmdDataFor("registry.io/proj", nil, nil, nil, "registry.io/meta", "", "")
	err := ResolveRepos(context.Background(), c, ResolveReposOptions{})
	if err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("expected mutual-exclusion error, got: %v", err)
	}
}

func TestResolveRepos_PresetOnlyOK(t *testing.T) {
	c := cmdDataFor("registry.io/proj", nil, nil, nil, "", "", "")
	if err := ResolveRepos(context.Background(), c, ResolveReposOptions{ImagesRepoRequired: true, MetaRepoRequired: true}); err != nil {
		t.Fatalf("preset-only should satisfy requiredness, got: %v", err)
	}
}

func TestResolveRepos_CacheFromDefaultsLocal(t *testing.T) {
	c := cmdDataFor("", nil, nil, nil, "", "", "")
	if err := ResolveRepos(context.Background(), c, ResolveReposOptions{}); err != nil {
		t.Fatal(err)
	}
	if got := *c.CacheFrom; len(got) != 1 || got[0] != storage.LocalStorageAddress {
		t.Fatalf("expected cache-from defaulted to [%q], got %v", storage.LocalStorageAddress, got)
	}
}

func TestResolveRepos_SecondaryAliasFoldsIntoCacheFrom(t *testing.T) {
	c := cmdDataFor("", []string{"registry.io/cache"}, nil, nil, "", "registry.io/secondary", "")
	if err := ResolveRepos(context.Background(), c, ResolveReposOptions{}); err != nil {
		t.Fatal(err)
	}
	got := *c.CacheFrom
	if len(got) != 2 || got[0] != "registry.io/cache" || got[1] != "registry.io/secondary" {
		t.Fatalf("expected secondary folded into cache-from, got %v", got)
	}
}

func TestResolveRepos_FinalRepoAliasFoldsIntoImagesRepo(t *testing.T) {
	c := cmdDataFor("", nil, nil, []string{"registry.io/images"}, "", "", "registry.io/final")
	if err := ResolveRepos(context.Background(), c, ResolveReposOptions{}); err != nil {
		t.Fatal(err)
	}
	got := *c.ImagesRepo
	if len(got) != 2 || got[0] != "registry.io/final" || got[1] != "registry.io/images" {
		t.Fatalf("expected final-repo folded into images-repo, got %v", got)
	}
}

func TestResolveRepos_ImagesRepoRequired(t *testing.T) {
	c := cmdDataFor("", nil, nil, nil, "", "", "")
	err := ResolveRepos(context.Background(), c, ResolveReposOptions{ImagesRepoRequired: true})
	if err == nil || !strings.Contains(err.Error(), "images-repo") {
		t.Fatalf("expected images-repo required error, got: %v", err)
	}
}

func TestResolveRepos_MetaRepoRequired(t *testing.T) {
	c := cmdDataFor("", nil, nil, nil, "", "", "")
	err := ResolveRepos(context.Background(), c, ResolveReposOptions{MetaRepoRequired: true})
	if err == nil || !strings.Contains(err.Error(), "meta-repo") {
		t.Fatalf("expected meta-repo required error, got: %v", err)
	}
}
