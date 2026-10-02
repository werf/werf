package contback

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

type CleanupProjectOptions struct {
	Repositories []string
}

func CleanupProject(ctx context.Context, projectName string, opts CleanupProjectOptions) error {
	if !strings.HasPrefix(projectName, "werf-test-") || strings.ContainsAny(projectName, "/:*?[] \t\n") {
		return fmt.Errorf("refuse cleanup of non-test project %q", projectName)
	}

	var cleanupErrors []error
	repos := append(slices.Clone(opts.Repositories), os.Getenv("WERF_REPO"), os.Getenv("WERF_FINAL_REPO"))
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
	if buildahAvailable() {
		if err := cleanupBuildahProject(ctx, projectName, repos); err != nil {
			cleanupErrors = append(cleanupErrors, err)
		}
	}
	return errors.Join(cleanupErrors...)
}

func cleanupDockerProjectImages(ctx context.Context, projectName string, repos []string, removeImage func(context.Context, string) error) error {
	list := func() ([]string, error) {
		args := []string{"images", "--filter", "label=werf=" + projectName, "--all", "--no-trunc", "--digests", "--format", "{{json .}}"}
		output, err := cleanupCommand(ctx, "docker", args)
		if err != nil {
			return nil, err
		}
		images, err := parseDockerCleanupImages(output)
		if err != nil {
			return nil, err
		}
		return projectImageReferences(images, projectName, repos), nil
	}

	return cleanupImageReferences(ctx, "docker", list, removeImage)
}

func cleanupImageReferences(ctx context.Context, backend string, list func() ([]string, error), removeImage func(context.Context, string) error) error {
	refs, err := list()
	if err != nil {
		return fmt.Errorf("list %s test images: %w", backend, err)
	}
	for pass := 0; len(refs) > 0; pass++ {
		var cleanupErrors []error
		onlyInUseErrors := true
		for _, ref := range refs {
			if err := removeImage(ctx, ref); err != nil {
				if !errors.Is(err, errCleanupImageInUse) {
					onlyInUseErrors = false
				}
				cleanupErrors = append(cleanupErrors, fmt.Errorf("remove %s test image %q: %w", backend, ref, err))
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
		if slices.Equal(refs, remaining) && onlyInUseErrors && errors.Is(cleanupErr, errCleanupImageInUse) && pass < 100 {
			timer := time.NewTimer(time.Second)
			select {
			case <-ctx.Done():
				timer.Stop()
				return errors.Join(cleanupErr, fmt.Errorf("wait for %s test images to be released: %w", backend, ctx.Err()))
			case <-timer.C:
			}
		} else if slices.Equal(refs, remaining) || pass >= 100 {
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
			if base == projectName || base == projectName+"-final" || slices.Contains(repos, repo) {
				refs = append(refs, name)
			}
		}
	}
	slices.Sort(refs)
	return slices.Compact(refs)
}
