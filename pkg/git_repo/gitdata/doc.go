// Package gitdata manages garbage collection of werf's host-local git data
// under $WERF_HOME/local_cache. It is the only package that performs
// selective GC of local_cache data; every rule below exists because that
// directory is shared.
//
// # Base invariant
//
// local_cache has no per-werf-version path segment (see localCacheDir in
// pkg/werf): it is shared by ALL werf binaries using the same WERF_HOME,
// across minor and major versions alike. On CI runners different release
// branches are routinely built by different werf versions on one host.
// Therefore any path under local_cache either belongs to the running binary
// or to another werf version that may be in the middle of a build right now.
//
// # Cache roots
//
// Each kind of data lives in its own root with a per-root version namespace:
//
//	git_repos/<v>      version: git_repo.GitReposCacheVersion      cleaned: wipeCacheDirs + LRU
//	git_mirrors/<v>    version: git_repo.GitMirrorsCacheVersion    cleaned: wipeCacheDirs + LRU
//	git_worktrees/<v>  version: git_repo.GitWorktreesCacheVersion  cleaned: wipeCacheDirs + LRU
//	git_archives/<v>   version: GitArchivesCacheVersion            cleaned: wipeCacheDirs + LRU
//	git_patches/<v>    version: GitPatchesCacheVersion             cleaned: wipeCacheDirs + LRU
//	manifests/<v>      version: image.ManifestCacheVersion         cleaned: never
//	lru_images/<v>     version: lrumeta.LRUImagesCacheVersion      cleaned: never
//
// Nothing cleans manifests and lru_images: a version bump there leaks the old
// data forever but destroys nothing. The whole local_cache is removed only by
// `werf host purge`.
//
// # Bumping a cache version
//
// A bump is allowed only when ALL of the following hold:
//
//  1. The on-disk format changed incompatibly and the change cannot be made
//     additive.
//  2. Every werf version that can share the host has a staleness guard in its
//     wipe path (see wipeCacheDirs). A bump against releases without the
//     guard deletes their cache in the middle of a build.
//  3. The transition cost is accepted: both layouts stay on disk until the
//     old one goes stale, plus a one-time re-fetch.
//
// Never bump "to invalidate the cache" — that is deletion of another
// version's data. A deliberate bump as a weapon is acceptable only when
// reusing old data is unsafe (wrong builds, leaked credentials), and then it
// goes into release notes. If compatibility with already released guard-less
// versions is needed, use a new root instead of a bump.
//
// # Deleting and reading foreign data
//
//   - A foreign version directory is deleted only when stale, never
//     immediately (see wipeCacheDirs for the window).
//   - Never read or write inside another version's namespace.
//   - Never put a non-directory directly into a version root: collectors
//     remove such entries as invalid.
//
// # Load-bearing artifacts (backward compatibility)
//
// While a version constant is unchanged, its namespace is shared with
// released werf versions and everything below is load-bearing. Changing any
// item is an incompatible change: new root or a bump by the rules above.
//
// git_repos/<v>/<repoID>:
//
//   - <repoID> is util.Sha256Hash of Remote.getFilesystemRelativePathByEndpoint();
//     the path function must not change.
//   - <repoID> IS the bare mirror itself, not a container of mirrors: werf
//     2.74.x os.Stat's it, treats existence as "already cloned" and goes
//     straight to PlainOpen, so nesting breaks old builds permanently.
//   - <repoID>/last_access_at in common-go/pkg/util/timestamps format,
//     updated on every access; an entry with a missing or corrupt timestamp
//     gets zero last-access time and is evicted first by LRU.
//   - The mirror carries refs/remotes/origin/* and tags; never rely on a
//     refspec written into the mirror's config by another version — use
//     explicit refspecs only (see buildFetchOptions).
//
// git_mirrors/<v>/<repoID>: shallow/ is a bare shallow mirror with
// last_access_at inside; requires_full is a marker file (not a mirror) that
// pins the repo to the full mirror; a repoID dir holding only the marker is
// valid and kept.
//
// git_worktrees/<v>/{local,remote}/<name> with last_access_at inside; for
// local, <name> is sha256 of the repository's absolute path. The full-mirror
// worktree dir is the unsuffixed remote/<repoID> so that werf 2.74.x reuses
// it; the shallow-mirror worktree dir keeps the .shallow suffix. Changing the
// naming scheme does not break old versions but orphans their entries: cache
// misses plus garbage until LRU.
//
// git_archives/<v>/<repoID>/<2 hex>/<id>.tar plus <id>.meta.json with a
// LastAccessTimestamp field rewritten on every access. git_patches/<v> is the
// same plus <id>.patch and <id>.patch.<hash>.paths_list; see the doc comments
// of GetGitArchivesAndRemoveInvalid and GetGitPatchesAndRemoveInvalid.
//
// Additive changes are safe exactly as far as other versions' collectors
// tolerate them: an extra plain file inside an entry is tolerated, an extra
// directory in a version root is not. A new kind of artifact gets its own
// root with its own version, its own wipeCacheDirs call and its own LRU
// collection.
package gitdata
