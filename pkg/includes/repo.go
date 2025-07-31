package includes

import (
	"context"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/git_repo"
)

type GitRepository interface {
	PlainOpen() (*git.Repository, error)
	GetName() string
	ReadCommitFile(ctx context.Context, commit, path string) (data []byte, err error)
	IsCommitDirectoryExist(ctx context.Context, commit, path string) (exist bool, err error)
	IsCommitFileExist(ctx context.Context, commit, path string) (exist bool, err error)
}

type gitRepository struct {
	repo GitRepository
}

type gitRepositoriesWithCache struct {
	repositories map[string]*gitRepository
}

func newGitRepositoriesWithCache() *gitRepositoriesWithCache {
	return &gitRepositoriesWithCache{
		repositories: make(map[string]*gitRepository),
	}
}

func (g *gitRepositoriesWithCache) add(ctx context.Context, i includeConf) error {
	if _, ok := g.repositories[i.Git]; !ok {
		r, err := newRepo(ctx, i)
		if err != nil {
			return fmt.Errorf("unable to initialize repository %s: %w", i.Git, err)
		}
		g.repositories[i.Git] = r
	}
	return nil
}

func (g *gitRepositoriesWithCache) getRepository(git string) (*gitRepository, error) {
	repo, ok := g.repositories[git]
	if !ok {
		return nil, fmt.Errorf("repository %s not found", git)
	}
	return repo, nil
}

func initRemoteRepos(ctx context.Context, cfg Config) (*gitRepositoriesWithCache, error) {
	repoCache := newGitRepositoriesWithCache()
	err := logboek.Context(ctx).Default().LogBlock("Initializing remote repositories").DoError(func() error {
		for _, i := range cfg.Includes {
			if err := repoCache.add(ctx, i); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return repoCache, nil
}

func isLocalRepo(path string) (bool, string) {
	absPath, err := util.ExpandPath(path)
	if err != nil {
		return false, ""
	}
	exists, err := util.DirExists(absPath)
	if err != nil {
		return false, ""
	}
	return exists, absPath
}

func newRepo(ctx context.Context, i includeConf) (*gitRepository, error) {
	isLocal, absPath := isLocalRepo(i.Git)
	if !isLocal {
		remoteRepo, err := newRemoteRepo(ctx, i)
		if err != nil {
			return nil, fmt.Errorf("unable to open remote repository %s: %w", i.Git, err)
		}
		return &gitRepository{repo: remoteRepo}, nil
	} else {
		localRepo, err := newLocalRepo(ctx, i, absPath)
		if err != nil {
			return nil, fmt.Errorf("unable to open local repository %s: %w", i.Git, err)
		}
		return &gitRepository{repo: localRepo}, nil
	}
}

func newRemoteRepo(ctx context.Context, i includeConf) (*git_repo.Remote, error) {
	remoteRepo, err := git_repo.OpenRemoteRepo(i.Git, i.Git, i.BasicAuth)
	if err != nil {
		return nil, fmt.Errorf("unable to open remote repository %s: %w", i.Git, err)
	}
	if err := remoteRepo.CloneAndFetch(ctx); err != nil {
		return nil, err
	}
	return remoteRepo, nil
}

func newLocalRepo(ctx context.Context, i includeConf, repoAbsPath string) (*git_repo.Local, error) {
	localRepo, err := git_repo.OpenLocalRepo(ctx, repoAbsPath, repoAbsPath, git_repo.OpenLocalRepoOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to open local repository %s: %w", i.Git, err)
	}
	return localRepo, nil
}

func getCommit(r *git.Repository, git, tag, branch, commit string) (*object.Commit, error) {
	switch {
	case commit != "":
		return commitRef(r, commit)
	case tag != "":
		return tagRef(r, tag)
	case branch != "":
		return branchRef(r, branch)
	default:
		return nil, fmt.Errorf("no commit, tag or branch specified for include %s", git)
	}
}

func commitRef(r *git.Repository, commit string) (*object.Commit, error) {
	rev := plumbing.Revision(commit)
	h, err := r.ResolveRevision(rev)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve commit %s: %w", commit, err)
	}
	return r.CommitObject(*h)
}

func tagRef(r *git.Repository, tag string) (*object.Commit, error) {
	tagRef, err := r.Reference(plumbing.NewTagReferenceName(tag), true)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag %s: %w", tag, err)
	}
	return commitRef(r, tagRef.Hash().String())
}

func branchRef(r *git.Repository, branch string) (*object.Commit, error) {
	branchRef, err := r.Reference(plumbing.NewBranchReferenceName(branch), true)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch %s: %w", branch, err)
	}
	return commitRef(r, branchRef.Hash().String())
}

func ref(git, tag, branch, commit string) (string, error) {
	switch {
	case tag != "":
		return tag, nil
	case branch != "":
		return branch, nil
	case commit != "":
		return commit, nil
	default:
		return "", fmt.Errorf("no ref specified for include %s", git)
	}
}
