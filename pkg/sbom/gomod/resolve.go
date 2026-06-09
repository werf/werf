package gomod

import (
	"context"
	"fmt"
	"path/filepath"

	cdx "github.com/CycloneDX/cyclonedx-go"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/sbom/cyclonedxutil"
)

type BOMPatcher struct {
	gitRepo      git_repo.GitRepo
	commit       string
	imageContext string
}

func NewBOMPatcher(gitRepo git_repo.GitRepo, commit, imageContext string) *BOMPatcher {
	return &BOMPatcher{
		gitRepo:      gitRepo,
		commit:       commit,
		imageContext: imageContext,
	}
}

func (p *BOMPatcher) Apply(ctx context.Context, bom *cdx.BOM) (*cdx.BOM, error) {
	if p.gitRepo == nil || p.commit == "" {
		return bom, nil
	}

	goModPath := filepath.Join(p.imageContext, "go.mod")

	exists, err := p.gitRepo.IsCommitFileExist(ctx, p.commit, goModPath)
	if err != nil {
		return nil, fmt.Errorf("check go.mod existence: %w", err)
	}
	if !exists {
		// Do not clutter stdout with warnings for non-Go projects
		logboek.Context(ctx).Debug().LogF("No go.mod found at %s, skipping version resolution\n", goModPath)
		return bom, nil
	}

	content, err := p.gitRepo.ReadCommitFile(ctx, p.commit, goModPath)
	if err != nil {
		return nil, fmt.Errorf("read go.mod: %w", err)
	}
	if len(content) == 0 {
		return bom, nil
	}

	info, err := ParseLocalReplaces(content)
	if err != nil {
		return nil, err
	}

	tags, err := p.gitRepo.TagsList(ctx)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}

	version, err := ResolveVersionFromTags(tags, func(tag string) (string, error) {
		return p.gitRepo.TagCommit(ctx, tag)
	}, p.commit)
	if err != nil {
		return nil, err
	}

	return cyclonedxutil.ResolveUnknownGoVersions(bom, version, info.ModulePath, info.LocalReplaceTargets, info.LocalReplacePaths), nil
}
