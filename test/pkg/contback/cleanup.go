package contback

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/distribution/reference"
)

type cleanupImage struct {
	ID    string   `json:"id"`
	Names []string `json:"names"`
}

var errCleanupImageInUse = errors.New("test image is still in use")

const cleanupBusyGracePeriod = 30 * time.Second

type CleanupProjectOptions struct {
	Repositories []string
	BackendModes []string
}

type BuildahContainer struct {
	ID      string
	ImageID string
	Names   []string
}

// BuildahStorageSnapshot is what the Buildah storage werf builds into holds for a test project:
// the images labeled with the project and every container in the storage, whichever build
// created it.
type BuildahStorageSnapshot struct {
	ProjectImageIDs []string
	Containers      []BuildahContainer
}

// SnapshotBuildahStorage reads the storage in-process, so it sees the containers werf created
// regardless of the buildah CLI version installed on the host.
func SnapshotBuildahStorage(ctx context.Context, projectName string) (BuildahStorageSnapshot, error) {
	if !strings.HasPrefix(projectName, "werf-test-") {
		return BuildahStorageSnapshot{}, fmt.Errorf("refuse to snapshot non-test project %q", projectName)
	}
	return snapshotBuildahStorage(ctx, projectName)
}

func CleanupProject(ctx context.Context, projectName string, opts CleanupProjectOptions) error {
	if !strings.HasPrefix(projectName, "werf-test-") || strings.ContainsAny(projectName, "/:*?[] \t\n") {
		return fmt.Errorf("refuse cleanup of non-test project %q", projectName)
	}

	useDocker, useBuildah := len(opts.BackendModes) == 0, len(opts.BackendModes) == 0
	for _, mode := range opts.BackendModes {
		switch mode {
		case "", "docker":
			useDocker = true
		case "native-rootless", "native-chroot", "auto", "default":
			useBuildah = true
		default:
			useDocker, useBuildah = true, true
		}
	}
	var cleanupErrors []error
	repos := append(slices.Clone(opts.Repositories), os.Getenv("WERF_REPO"), os.Getenv("WERF_FINAL_REPO"))
	if useDocker {
		if _, err := exec.LookPath("docker"); err == nil {
			if err := cleanupDockerProjectImages(ctx, projectName, repos, func(ctx context.Context, ref string) error {
				_, err := cleanupCommand(ctx, "docker", []string{"rmi", "--no-prune", ref})
				return err
			}); err != nil {
				cleanupErrors = append(cleanupErrors, err)
			}
		} else if !errors.Is(err, exec.ErrNotFound) {
			cleanupErrors = append(cleanupErrors, fmt.Errorf("locate docker CLI: %w", err))
		}
	}
	if useBuildah && buildahAvailable() {
		if err := cleanupBuildahProject(ctx, projectName, repos); err != nil {
			cleanupErrors = append(cleanupErrors, err)
		}
	}
	return errors.Join(cleanupErrors...)
}

func cleanupDockerProjectImages(ctx context.Context, projectName string, repos []string, removeImage func(context.Context, string) error) error {
	var taggedIDs []string
	list := func() ([]string, error) {
		if len(taggedIDs) > 0 {
			dangling := inspectDanglingDockerImages(ctx, taggedIDs)
			taggedIDs = nil
			if len(dangling) > 0 {
				return dangling, nil
			}
		}
		args := []string{"images", "--filter", "label=werf=" + projectName, "--all", "--no-trunc", "--digests", "--format", "{{json .}}"}
		output, err := cleanupCommand(ctx, "docker", args)
		if err != nil {
			return nil, err
		}
		images, err := parseDockerCleanupImages(output)
		if err != nil {
			return nil, err
		}
		refs := projectImageReferences(images, projectName, repos)
		for _, image := range images {
			if len(image.Names) == 1 && image.ID != "" {
				if _, owned := slices.BinarySearch(refs, image.Names[0]); owned {
					taggedIDs = append(taggedIDs, image.ID)
				}
			}
		}
		slices.Sort(taggedIDs)
		taggedIDs = slices.Compact(taggedIDs)
		return refs, nil
	}

	return cleanupImageReferences(ctx, "docker", io.Discard, list, func(ctx context.Context, ref string) error {
		err := removeImage(ctx, ref)
		if err != nil {
			taggedIDs = nil
		}
		return err
	})
}

func inspectDanglingDockerImages(ctx context.Context, ids []string) []string {
	args := append([]string{"image", "inspect"}, ids...)
	output, err := exec.CommandContext(ctx, "docker", args...).Output()
	if err != nil {
		var exitErr *exec.ExitError
		if ctx.Err() != nil || !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
			return nil
		}
		for line := range strings.SplitSeq(strings.TrimSpace(string(exitErr.Stderr)), "\n") {
			id, missing := strings.CutPrefix(line, "Error response from daemon: No such image: ")
			if _, known := slices.BinarySearch(ids, id); !missing || !known {
				return nil
			}
		}
	}
	var images []struct {
		ID          string
		RepoTags    []string
		RepoDigests []string
	}
	if err := json.Unmarshal(output, &images); err != nil {
		return nil
	}
	var dangling []string
	for _, image := range images {
		if _, known := slices.BinarySearch(ids, image.ID); known && len(image.RepoTags) == 0 && len(image.RepoDigests) == 0 {
			dangling = append(dangling, image.ID)
		}
	}
	return dangling
}

func cleanupImageReferences(ctx context.Context, backend string, diagnostics io.Writer, list func() ([]string, error), removeImage func(context.Context, string) error) error {
	refs, err := list()
	if err != nil {
		return fmt.Errorf("list %s test images: %w", backend, err)
	}
	var busySince time.Time
	reportedBusy := make(map[string]bool)
	for pass := 0; len(refs) > 0; pass++ {
		var cleanupErrors []error
		allRefsBusy := true
		for _, ref := range refs {
			if err := removeImage(ctx, ref); err != nil {
				if !errors.Is(err, errCleanupImageInUse) {
					allRefsBusy = false
				} else {
					if busySince.IsZero() {
						busySince = time.Now()
					}
					if !reportedBusy[ref] {
						fmt.Fprintf(diagnostics, "%s cleanup: image %q is still in use: %v\n", backend, ref, err)
						reportedBusy[ref] = true
					}
				}
				cleanupErrors = append(cleanupErrors, fmt.Errorf("remove %s test image %q: %w", backend, ref, err))
			} else {
				allRefsBusy = false
			}
		}
		remaining, err := list()
		if err != nil {
			cleanupErrors = append(cleanupErrors, fmt.Errorf("verify %s test image cleanup: %w", backend, err))
			return errors.Join(cleanupErrors...)
		}
		if len(remaining) == 0 {
			return nil
		}
		cleanupErr := errors.Join(cleanupErrors...)
		unchanged := slices.Equal(refs, remaining)
		if !unchanged {
			busySince = time.Time{}
		}
		if unchanged && allRefsBusy && time.Since(busySince) < cleanupBusyGracePeriod && pass < 100 {
			timer := time.NewTimer(time.Second)
			select {
			case <-ctx.Done():
				timer.Stop()
				return errors.Join(cleanupErr, fmt.Errorf("wait for %s test images to be released: %w", backend, ctx.Err()))
			case <-timer.C:
			}
		} else if unchanged || pass >= 100 {
			cleanupErrors = append(cleanupErrors, fmt.Errorf("%s test images remain: %s", backend, strings.Join(remaining, ", ")))
			return errors.Join(cleanupErrors...)
		}
		refs = remaining
	}
	return nil
}

func cleanupCommand(ctx context.Context, backend string, args []string) ([]byte, error) {
	output, err := exec.CommandContext(ctx, backend, args...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s %v: %w: %s", backend, args, err, output)
	}
	return output, nil
}

func parseDockerCleanupImages(output []byte) ([]cleanupImage, error) {
	var images []cleanupImage
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		var row struct {
			ID         string
			Repository string
			Tag        string
			Digest     string
		}
		if err := json.Unmarshal(scanner.Bytes(), &row); err != nil {
			return nil, fmt.Errorf("decode docker image: %w", err)
		}
		var names []string
		if row.Repository != "<none>" && row.Tag != "<none>" {
			names = []string{row.Repository + ":" + row.Tag}
		} else if row.Repository != "<none>" {
			if row.Digest == "" || row.Digest == "<none>" {
				return nil, fmt.Errorf("docker image %q has a repository but no tag or digest", row.ID)
			}
			names = []string{row.Repository + "@" + row.Digest}
		}
		images = append(images, cleanupImage{ID: row.ID, Names: names})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read docker images: %w", err)
	}
	return images, nil
}

func projectImageReferences(images []cleanupImage, projectName string, repos []string) []string {
	var refs []string
	ownedRepos := make(map[string]bool)
	for _, repo := range repos {
		if named, err := reference.ParseNormalizedNamed(repo); err == nil {
			ownedRepos[reference.FamiliarName(named)] = true
		}
	}
	aliasedIDs := make(map[string]bool)
	for _, img := range images {
		if len(img.Names) != 0 {
			aliasedIDs[img.ID] = true
		}
	}
	for _, img := range images {
		if len(img.Names) == 0 {
			if img.ID != "" && !aliasedIDs[img.ID] {
				refs = append(refs, img.ID)
			}
			continue
		}
		for _, name := range img.Names {
			named, err := reference.ParseNormalizedNamed(name)
			if err != nil {
				continue
			}
			repo := reference.FamiliarName(named)
			base := path.Base(reference.Path(named))
			if base == projectName || base == projectName+"-final" || ownedRepos[repo] {
				refs = append(refs, name)
			}
		}
	}
	slices.Sort(refs)
	return slices.Compact(refs)
}
